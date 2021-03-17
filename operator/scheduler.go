package operator

import (
	"fmt"
	"reflect"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type reconciler struct {
	delete bool
	dep    *proto.Deployment
	spec   *proto.ClusterSpec
	res    *reconcileResult
}

type allocSet []*proto.Instance

func (a *allocSet) byName() (res map[string]*proto.Instance) {
	res = map[string]*proto.Instance{}

	for _, i := range *a {
		res[i.Name] = i
	}
	return
}

func (a *allocSet) byGroup(typ string) (byGroup allocSet) {
	byGroup = allocSet{}

	for _, i := range *a {
		if i.Group.Type == typ {
			byGroup = append(byGroup, i)
		}
	}
	return
}

func (a *allocSet) filterByStatus(status ...proto.Instance_Status) (allocSet, allocSet) {
	return a.filter(func(i *proto.Instance) bool {
		for _, s := range status {
			if i.Status == s {
				return true
			}
		}
		return false
	})
}

func (a *allocSet) filterByStopping() (allocSet, allocSet) {
	return a.filter(func(i *proto.Instance) bool {
		return i.Desired == proto.InstanceDesiredStopped
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
				lost = append(lost, ii)
			}
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}

func (a *allocSet) canaries() (canaries allocSet, add allocSet, healthy allocSet, untainted allocSet) {
	canaries = allocSet{}
	untainted = allocSet{}
	healthy = allocSet{}
	add = allocSet{}

	for _, i := range *a {
		//fmt.Println("_ i _")
		//fmt.Println(i.Desired)

		if i.Canary {
			if i.Status == proto.Instance_STOPPED {
				add = append(add, i)
				continue
			}
			if i.Desired == proto.InstanceDesiredStopped {
				// its a destructive canary that is shutting down
				canaries = append(canaries, i)
				continue
			}
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

type reconcileResult struct {
	groupUpdates []groupUpdate
	place        []instancePlaceResult
	stop         []instanceStopResult
	promote      []*proto.Instance
	out          []*proto.Instance
	done         bool
}

func (r *reconcileResult) print() {
	fmt.Printf("place: %d, stop: %d, promote: %d\n", len(r.place), len(r.stop), len(r.promote))
}

type instanceStopResult struct {
	instance *proto.Instance
	update   bool
	group    *proto.ClusterSpec_Group
}

type instancePlaceResult struct {
	name       string
	instance   *proto.Instance
	group      *proto.ClusterSpec_Group
	update     bool
	reschedule bool
}

type groupUpdate struct {
	name   string
	status string
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

func (r *reconciler) computePlacements(grp *proto.ClusterSpec_Group, untainted, destructive allocSet, placedCanaries int) (place []instancePlaceResult) {
	place = []instancePlaceResult{}
	total := len(untainted) + len(destructive) + placedCanaries

	for i := total; i < int(grp.Count); i++ {
		id := uuid.UUID8()
		indx := i + 1 // index starts with 1

		// name of the node
		var name string
		if grp.Type == "" {
			name = fmt.Sprintf("%s-%d", id, indx)
		} else {
			name = fmt.Sprintf("%s-%s-%d", id, grp.Type, indx)
		}
		place = append(place, instancePlaceResult{
			name:  name,
			group: grp,
		})
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

func (r *reconciler) Compute() {
	// r.res = []*update{}
	r.res = &reconcileResult{}

	if r.delete {
		// remove all the running instances
		pending := false
		for _, i := range r.dep.Instances {
			if i.Status == proto.Instance_RUNNING {
				if i.Desired != proto.InstanceDesiredStopped {
					r.res.stop = append(r.res.stop, instanceStopResult{
						instance: i,
					})
				} else {
					pending = true
				}
			}
		}
		if len(r.res.stop) == 0 && !pending {
			// delete completed
			r.res.done = true
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
	var stopping allocSet
	stopping, untainted = untainted.filterByStopping()

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
		r.res.place = append(r.res.place, instancePlaceResult{
			instance:   i,
			reschedule: true,
		})
	}

	// stop nodes
	for _, i := range stop {
		r.res.stop = append(r.res.stop, instanceStopResult{
			instance: i,
		})
	}
	if len(stop) != 0 {
		return false
	}

	untainted = untainted.difference(stop)

	// get the pending and promoted canaries from the set
	var canaries, readyToAllocate, promoted allocSet
	canaries, readyToAllocate, promoted, untainted = set.canaries()

	var canaryPlacements []instancePlaceResult
	for _, i := range readyToAllocate {
		r.res.out = append(r.res.out, i)

		canaryPlacements = append(canaryPlacements, instancePlaceResult{
			instance: i,
			group:    grp,
			update:   true,
		})
	}

	// any canary placement is ready to be allocated
	for _, i := range canaryPlacements {
		r.res.place = append(r.res.place, i)
	}

	// promote healthy canaries and add them to the untainted set
	for _, i := range promoted {
		r.res.promote = append(r.res.promote, i)
	}

	untainted = untainted.join(promoted)

	// destructive updates (TODO: inplace updates)
	var destructive allocSet
	destructive, untainted = computeUpdates(r.spec, grp, untainted)

	// rolling update
	updates := []instanceStopResult{}

	areCanaries := len(canaries) + len(readyToAllocate)
	if len(destructive) > 0 && areCanaries == 0 {
		num := min(len(destructive), maxParallel)
		for _, instance := range destructive[:num] {
			updates = append(updates, instanceStopResult{
				instance: instance,
				group:    grp,
				update:   true,
			})
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
	place := r.computePlacements(grp, untainted, destructive, len(readyToAllocate)) // sketchy right now

	if allHealthy {
		// only place new allocs for scale up if the cluster is stable
		if len(updates) == 0 {
			for _, p := range place {
				r.res.place = append(r.res.place, p)
			}
		}

		// place the updates
		for _, p := range updates {
			r.res.stop = append(r.res.stop, p)
		}
	}

	r.res.done = false
	if allHealthy {
		if len(reschedule) == 0 && len(readyToAllocate) == 0 && len(stopping) == 0 && len(updates) == 0 && len(place) == 0 && len(lost) == 0 {
			r.res.done = true
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
