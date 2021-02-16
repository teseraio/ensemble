package operator

import (
	"github.com/teseraio/ensemble/operator/proto"
)

// reconciler does a first reconcilization and determines which nodes have to be deployed or changed
// and creates a plan that describes this changes
type allocReconciler struct {
	c      *proto.Cluster
	nodes  []*proto.Instance
	result *reconcileResult
}

type reconcileResult struct {
	name string

	// whether this group is being removed
	delete bool

	// scale up nodes
	add *proto.Instance

	// scale down nodes
	del *proto.Instance
}

func (a *allocReconciler) createAllocMatrix() map[string][]*proto.Instance {
	allocMatrix := map[string][]*proto.Instance{}
	for _, node := range a.nodes {
		set := node.Group
		if _, ok := allocMatrix[set]; !ok {
			allocMatrix[set] = []*proto.Instance{}
		}
		allocMatrix[set] = append(allocMatrix[set], node)
	}
	for _, group := range a.c.Groups {
		if _, ok := allocMatrix[group.Nodeset]; !ok {
			allocMatrix[group.Nodeset] = []*proto.Instance{}
		}
	}
	return allocMatrix
}

func (a *allocReconciler) reconcile() {
	allocMatrix := a.createAllocMatrix()

	if a.c.Stop {
		// the cluster has been removed, remove all the instances
		a.result.delete = true
		return
	}

	for group, instances := range allocMatrix {
		res := a.reconcileGroup(group, instances)
		a.result = res
	}
}

func (a *allocReconciler) reconcileGroup(group string, instances []*proto.Instance) (reconcile *reconcileResult) {
	grp := a.c.LookupGroup(group)
	if grp == nil {
		// the group was removed, remove all the instances
		reconcile.delete = true
		return
	}

	reconcile = &reconcileResult{
		name: group,
	}

	numInstances := int64(len(instances))
	isScale := grp.Count != numInstances
	if isScale {
		if grp.Count < numInstances {
			// scale down
			a.result.add = &proto.Instance{}
		} else {
			// scale up
			reconcile.scaleUp = grp.Count - numInstances
		}
	}

	// check if there is any update
	curRevision := grp.Revision
	for _, inst := range instances {
		if inst.Revision < curRevision {
			reconcile.updateNodes = append(reconcile.updateNodes, inst)
		}
	}
	return
}
