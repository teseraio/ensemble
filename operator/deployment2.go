package operator

import (
	"github.com/teseraio/ensemble/operator/proto"
)

type deployment2 struct {
}

func (d *deployment2) reconcile() {

}

type reconciler2 struct {
	dep        *proto.Deployment
	spec       *proto.ClusterSpec2
	res        []*update
	isComplete bool
}

func (r *reconciler2) appendUpdate(instance *proto.Instance, status string) {
	r.res = append(r.res, &update{status: status, instance: instance})
}

type allocSet []*proto.Instance

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

func (r *reconciler2) computePlacements(grp *proto.ClusterSpec2_Group, untainted []*proto.Instance) (place allocSet) {
	place = allocSet{}

	for i := len(untainted) + 1; i < int(grp.Count); i++ {
		place = append(place, &proto.Instance{}) // Add index
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

	set := allocSet(r.dep.Instances)

	// detect the stopped nodes (TODO: migrate)
	reschedule, untainted := set.reschedule()

	// compute stop
	stop := r.computeStop(r.spec.Group, reschedule, untainted)
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
	destructive, untainted = computeUpdates(r.spec.Group, untainted)

	// rolling update
	updates := allocSet{}
	if len(destructive) > 0 && len(canaries) == 0 {
		num := min(len(destructive), maxParallel)
		for _, instance := range destructive[:num] {
			updates = append(updates, instance)
		}
	}
	isRolling := len(updates) != 0

	// check if all the instances are healthy
	allHealthy := true
	for _, i := range untainted {
		if !i.Healthy {
			allHealthy = false
			break
		}
	}

	// compute placements
	place := r.computePlacements(r.spec.Group, untainted)

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

	r.isComplete = len(destructive)+len(place)+len(reschedule) == 0 && !isRolling
}

func computeUpdates(grp *proto.ClusterSpec2_Group, alloc allocSet) (destructive allocSet, untainted allocSet) {
	untainted = allocSet{}
	destructive = allocSet{}

	for _, i := range alloc {
		if i.Revision != grp.Revision {
			destructive = append(destructive, i)
		} else {
			untainted = append(untainted, i)
		}
	}
	return
}

/*
// evictAndPlace is used to mark allocations for evicts and add them to the placement queue
func evictAndPlace(diff *diffAlloc, allocs []*proto.Instance, limit *int) {
	n := len(allocs)
	for i := 0; i < n && i < *limit; i++ {
		a := allocs[i]
		diff.add = append(diff.add, a)
	}
	if n <= *limit {
		*limit -= n
	} else {
		*limit = 0
	}
}

func diffAllocs(instances []*proto.Instance, required map[string]*proto.ClusterSpec2_Group) *diffAlloc {
	out := &diffAlloc{}
	existing := map[string]struct{}{}

	for _, instance := range instances {
		name := instance.Name
		existing[name] = struct{}{}

		grp, ok := required[name]
		if !ok {
			out.del = append(out.del, instance)
			// remove
			continue
		}

		if instance.Revision != grp.Revision {
			// update
			out.update = append(out.update, instance)
			continue
		}
	}

	for name, _ := range required {
		if _, ok := existing[name]; !ok {
			// in place update
			out.add = append(out.add, &proto.Instance{Name: name})
		}
	}
	return out
}

func materializeGroup(spec *proto.ClusterSpec2) map[string]*proto.ClusterSpec2_Group {
	out := map[string]*proto.ClusterSpec2_Group{}
	for i := int64(0); i < spec.Group.Count; i++ {
		name := fmt.Sprintf("%s.%d", spec.Group.Name, i)
		out[name] = spec.Group
	}
	return out
}
*/
