package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type deployment2 struct {
}

func (d *deployment2) reconcile() {

}

// interface required by external actors that modify the behaviour
type reconcilerImpl interface {
}

type reconciler2 struct {
	dep  *proto.Deployment
	spec *proto.ClusterSpec2
	res  []*update
	done bool
}

func (r *reconciler2) appendUpdate(instance *proto.Instance, status string) {
	r.res = append(r.res, &update{status: status, instance: instance})
}

type allocSet []*proto.Instance

func (a *allocSet) byGroup(name string) (byGroup allocSet) {
	byGroup = allocSet{}

	for _, i := range *a {
		if i.Group.Name == name {
			byGroup = append(byGroup, i)
		}
	}
	return
}

func (a *allocSet) reschedule() (down allocSet, untainted allocSet) {
	// returns the nodes that need rescheduling
	down = allocSet{}
	untainted = allocSet{}

	for _, i := range *a {
		// TODO: Migrated
		if i.Status == proto.Instance_FAILED {
			down = append(down, i)
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

func (r *reconciler2) computeStop(grp *proto.ClusterSpec2_Group, reschedule allocSet, untainted allocSet) (stop allocSet) {
	stop = allocSet{}

	remove := len(untainted) - int(grp.Count)
	for i := remove; i > 0; i-- {
		stop = append(stop, untainted[i])
	}

	return stop
}

func (r *reconciler2) computePlacements(grp *proto.ClusterSpec2_Group, untainted, destructive allocSet) (place allocSet) {
	place = allocSet{}

	total := len(untainted) + len(destructive)

	fmt.Println("- total -")
	fmt.Println(len(untainted))
	fmt.Println(total)
	fmt.Println(grp.Count)

	for i := total; i < int(grp.Count); i++ {
		instance := &proto.Instance{
			ID:      uuid.UUID(),
			Cluster: r.spec.Name,
			Index:   int64(i),
			Name:    fmt.Sprintf("%s%d", r.spec.Name, i),
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

func (r *reconciler2) Compute() {
	r.res = []*update{}

	for _, grp := range r.spec.Groups {
		if !r.computeGroup(grp) {
			break
		}
	}
}

func (r *reconciler2) computeGroup(grp *proto.ClusterSpec2_Group) bool {
	set := allocSet(r.dep.Instances)
	set = set.byGroup(grp.Name)

	// detect the stopped nodes (TODO: migrate)
	reschedule, untainted := set.reschedule()

	// compute stop
	stop := r.computeStop(grp, reschedule, untainted)
	untainted = untainted.difference(stop)

	// get the pending and promoted canaries from the set
	var canaries, promoted allocSet
	canaries, promoted, untainted = set.canaries()

	// promote healthy canaries and add them to the untainted set
	for _, i := range promoted {
		r.appendUpdate(i, "promote")
	}
	untainted = untainted.join(promoted)

	// destructive updates (TODO: inplace)
	var destructive allocSet
	destructive, untainted = computeUpdates(grp, untainted)

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
		for _, p := range place {
			r.appendUpdate(p, "add")
		}

		// place the updates
		for _, p := range updates {
			r.appendUpdate(p, "update")
		}
	}

	if allHealthy {
		r.done = true
	} else {
		r.done = false
	}

	isComplete := len(destructive)+len(place)+len(reschedule) == 0 && !isRolling
	return isComplete
}

func computeUpdates(grp *proto.ClusterSpec2_Group, alloc allocSet) (destructive allocSet, untainted allocSet) {
	untainted = allocSet{}
	destructive = allocSet{}

	for _, i := range alloc {
		if i.Group.Revision != grp.Revision {
			destructive = append(destructive, i)
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}
