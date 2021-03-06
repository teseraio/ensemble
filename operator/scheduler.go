package operator

import (
	"fmt"
	"reflect"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type schedState interface {
	LoadDeployment(id string) (*proto.Deployment, error)
	GetComponent(id string) (*proto.Component, error)
}

type scheduler struct {
	state     schedState
	handlerFn func(backend string) (Handler, error)
}

func (s *scheduler) Process(eval *proto.Evaluation) error {
	deployment, err := s.state.LoadDeployment(eval.ClusterID)
	if err != nil {
		return err
	}
	component, err := s.state.GetComponent(deployment.CompID)
	if err != nil {
		return err
	}

	var clusterSpec proto.ClusterSpec
	if err := gproto.Unmarshal(component.Spec.Value, &clusterSpec); err != nil {
		return err
	}

	handler, err := s.handlerFn(clusterSpec.Backend)
	if err != nil {
		return err
	}
	fmt.Println(handler)

	// reconcile the state
	rec := &reconciler{
		dep:  deployment,
		spec: &clusterSpec,
	}
	rec.Compute()

	return nil
}

type reconciler struct {
	delete bool
	dep    *proto.Deployment
	spec   *proto.ClusterSpec
	res    []*update
	done   bool
}

func (r *reconciler) appendUpdate(instance *proto.Instance, status string) {
	r.res = append(r.res, &update{status: status, instance: instance})
}

type allocSet []*proto.Instance

func (a *allocSet) byGroup(typ string) (byGroup allocSet) {
	byGroup = allocSet{}

	for _, i := range *a {
		if i.Group.Type == typ {
			byGroup = append(byGroup, i)
		}
	}
	return
}

func (a *allocSet) filterByStopping() (allocSet, allocSet) {
	return a.filter(func(i *proto.Instance) bool {
		return i.Desired == "Stop"
	})
}

func (a *allocSet) filter(f func(i *proto.Instance) bool) (b allocSet, c allocSet) {
	b = allocSet{}
	c = allocSet{}

	for _, i := range *a {
		if f(i) {
			b = append(b, i)
		} else {
			c = append(c, i)
		}
	}
	return
}

const maxAttempts = 3

func (a *allocSet) reschedule() (down allocSet, lost allocSet, untainted allocSet) {
	// returns the nodes that need rescheduling
	down = allocSet{}
	untainted = allocSet{}
	lost = allocSet{}

	for _, i := range *a {
		// TODO: Migrated
		if i.Status == proto.Instance_FAILED {
			if i.Reschedule == nil {
				i.Reschedule = &proto.Instance_Reschedule{}
			}
			ii := i.Copy()
			if i.Reschedule.Attempts < maxAttempts {
				ii.Reschedule.Attempts++
				down = append(down, ii)
			} else {
				lost = append(lost)
			}
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}

func (a *allocSet) canaries() (canaries allocSet, healthy allocSet, untainted allocSet) {
	canaries = allocSet{}
	untainted = allocSet{}
	healthy = allocSet{}

	for _, i := range *a {
		if i.Canary {
			if i.Healthy {
				// promote canary
				i.Canary = false
				healthy = append(healthy, i)
			} else {
				canaries = append(canaries, i)
			}
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}

func (a *allocSet) add(other allocSet) {
	existing := map[string]struct{}{}
	for _, item := range *a {
		existing[item.ID] = struct{}{}
	}
	for _, item := range other {
		if _, ok := existing[item.ID]; !ok {
			(*a) = append(*a, item)
		}
	}
}

func (a *allocSet) join(other allocSet) (res allocSet) {
	res = allocSet{}
	res.add(*a)
	res.add(other)
	return
}

func (a *allocSet) difference(others allocSet) (res allocSet) {
	res = allocSet{}

	for _, item := range *a {
		exists := false
		for _, i := range others {
			if i.ID == item.ID {
				exists = true
			}
		}
		if !exists {
			res = append(res, item)
		}
	}
	return
}

type diffAlloc struct {
	add, del, update allocSet
}

type update struct {
	status   string
	instance *proto.Instance
}

func (r *reconciler) computeStop(grp *proto.ClusterSpec_Group, reschedule allocSet, untainted allocSet) (stop allocSet) {
	stop = allocSet{}
	remove := len(untainted) + len(reschedule) - int(grp.Count)

	if remove <= 0 {
		return
	}
	for i := 0; i < len(reschedule); i++ {
		stop = append(stop, reschedule[i])
		remove--
		if remove == 0 {
			return
		}
	}

	for i := 0; i < len(untainted); i++ {
		// TODO: Sort by health check
		stop = append(stop, untainted[i])
		remove--
		if remove == 0 {
			return
		}
	}
	return stop
}

func (r *reconciler) computePlacements(grp *proto.ClusterSpec_Group, untainted, destructive allocSet) (place allocSet) {
	place = allocSet{}
	total := len(untainted) + len(destructive)

	for i := total; i < int(grp.Count); i++ {
		id := uuid.UUID8()

		/*
			var name string
			if grp.Type != "" {
				// if the group has a name <name>-<group>-<indx>
				name = fmt.Sprintf("%s-%s-%d", id, grp.Type, i)
			} else {
				// if the group does not have a name just <name>-<indx>
				name = fmt.Sprintf("%s", id)
			}
		*/
		// fmt.Println(name)

		instance := &proto.Instance{
			ID:      id,
			Cluster: r.spec.Name,
			Index:   int64(i),
			Name:    id,
			Group:   grp,
			Spec:    &proto.NodeSpec{},
		}
		place = append(place, instance) // Add index
	}
	return
}

var maxParallel = 2

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func (r *reconciler) print() {
	for _, i := range r.res {
		fmt.Printf("Res: %s %s (%s) (%s)\n", i.status, i.instance.ID, i.instance.Group.Type, i.instance.FullName())
	}
}

func (r *reconciler) gather(status string) []*proto.Instance {
	res := []*proto.Instance{}
	for _, i := range r.res {
		if i.status == status {
			res = append(res, i.instance)
		}
	}
	return res
}

func (r *reconciler) check(s string) (count int) {
	for _, i := range r.res {
		if i.status == s {
			count++
		}
	}
	return
}

func (r *reconciler) Compute() {
	r.res = []*update{}

	if r.delete {
		// remove all the running instances
		pending := false
		for _, i := range r.dep.Instances {
			if i.Status == proto.Instance_RUNNING {
				if i.Desired != "Stop" {
					ii := i.Copy()
					ii.Desired = "Stop"
					r.appendUpdate(ii, "stop")
				} else {
					pending = true
				}
			}
		}
		if len(r.res) == 0 && !pending {
			// delete completed
			r.done = true
		}
		return
	}

	for _, grp := range r.spec.Groups {
		if !r.computeGroup(grp) {
			break
		}
	}
}

func (r *reconciler) computeGroup(grp *proto.ClusterSpec_Group) bool {
	set := allocSet(r.dep.Instances)
	set = set.byGroup(grp.Type)

	// detect the stopped nodes (TODO: migrate)
	reschedule, lost, untainted := set.reschedule()

	// avoid the nodes that are already being stopped
	_, untainted = untainted.filterByStopping()

	var stop allocSet
	if grp.Count == 0 {
		// purge the group
		stop = untainted
	} else {
		// scale down
		stop = r.computeStop(grp, reschedule, untainted)
	}

	// remove the reschedule nodes if we are stopping any
	reschedule = reschedule.difference(stop)
	for _, i := range reschedule {
		r.appendUpdate(i, "reschedule")
	}

	// stop nodes
	for _, i := range stop {
		ii := i.Copy()
		ii.Desired = "Stop"

		r.appendUpdate(ii, "stop")
	}
	if len(stop) != 0 {
		return false
	}

	untainted = untainted.difference(stop)

	// get the pending and promoted canaries from the set
	var canaries, promoted allocSet
	canaries, promoted, untainted = set.canaries()

	// promote healthy canaries and add them to the untainted set
	for _, i := range promoted {
		ii := i.Copy()
		ii.Canary = false

		r.appendUpdate(ii, "promote")
	}

	untainted = untainted.join(promoted)

	// destructive updates (TODO: inplace)
	var destructive allocSet
	destructive, untainted = computeUpdates(r.spec, grp, untainted)

	// rolling update
	updates := allocSet{}
	if len(destructive) > 0 && len(canaries) == 0 {
		num := min(len(destructive), maxParallel)
		for _, instance := range destructive[:num] {

			// set the other node as pending to be removed
			ii := instance.Copy()
			ii.Healthy = false
			ii.Group = grp
			ii.Ip = ""
			ii.Canary = true

			updates = append(updates, ii)

			// mark is as down the original instance
			instance.Desired = "DOWN"
		}
	}
	isRolling := len(updates) != 0

	// we can make the canaries part of the untainted
	untainted = untainted.join(canaries)

	// check if all the instances are healthy
	allHealthy := true
	for _, i := range untainted {
		if !i.Healthy {
			allHealthy = false
			break
		}
	}

	// compute placements
	place := r.computePlacements(grp, untainted, destructive) // sketchy right now

	if allHealthy {
		// only place new allocs for scale up if the cluster is stable
		if len(updates) == 0 {
			for _, p := range place {
				r.appendUpdate(p, "add")
			}
		}

		// place the updates
		for _, p := range updates {
			r.appendUpdate(p, "update")
		}
	}

	r.done = false
	if allHealthy {
		if len(updates) == 0 && len(place) == 0 && len(lost) == 0 {
			r.done = true
		}
	}

	isComplete := len(destructive)+len(place)+len(reschedule)+len(lost) == 0 && !isRolling
	return isComplete
}

func computeUpdates(spec *proto.ClusterSpec, grp *proto.ClusterSpec_Group, alloc allocSet) (destructive allocSet, untainted allocSet) {
	untainted = allocSet{}
	destructive = allocSet{}

	for _, i := range alloc {
		if spec.Sequence != i.Sequence {
			// check if the changes are destructive
			if areDiff(grp, i.Group) {
				destructive = append(destructive, i)
			} else {
				untainted = append(untainted, i)
			}
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}

func areDiff(grp *proto.ClusterSpec_Group, other *proto.ClusterSpec_Group) bool {
	if !reflect.DeepEqual(grp.Config, other.Config) {
		return true
	}
	if !reflect.DeepEqual(grp.Resources, other.Resources) {
		return true
	}
	return false
}
