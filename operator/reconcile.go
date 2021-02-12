package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
)

// reconciler does a first reconcilization and determines which nodes have to be deployed or changed
// and creates a plan that describes this changes
type allocReconciler struct {
	c      *proto.Cluster
	nodes  []*proto.Instance
	result *proto.Plan
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
	a.result = &proto.Plan{
		Steps: []*proto.Plan_Step{},
	}
	allocMatrix := a.createAllocMatrix()

	if a.c.Stop {
		// the cluster has been removed, remove all the instances
		a.result.Delete = true
		return
	}

	for group, instances := range allocMatrix {
		a.reconcileGroup(group, instances)
	}
}

func (a *allocReconciler) reconcileGroup(group string, instances []*proto.Instance) {
	grp := a.c.LookupGroup(group)
	if grp == nil {
		// the group was removed, remove all the instances
		a.result.Steps = append(a.result.Steps, &proto.Plan_Step{
			Group:  group,
			Action: &proto.Plan_Step_ActionDelete_{},
		})
		return
	}

	numInstances := int64(len(instances))
	isScale := grp.Count != numInstances
	if isScale {
		if grp.Count < numInstances {
			// scale down
			a.result.Steps = append(a.result.Steps, &proto.Plan_Step{
				Group: group,
				Action: &proto.Plan_Step_ActionScaleDown_{
					ActionScaleDown: &proto.Plan_Step_ActionScaleDown{
						NumNodes: numInstances - grp.Count,
					},
				},
			})
		} else {
			// scale up
			a.result.Steps = append(a.result.Steps, &proto.Plan_Step{
				Group: group,
				Action: &proto.Plan_Step_ActionScaleUp_{
					ActionScaleUp: &proto.Plan_Step_ActionScaleUp{
						NumNodes: grp.Count - numInstances,
					},
				},
			})
		}
	}

	// check if there is any update
	curRevision := grp.Revision
	for _, inst := range instances {
		if inst.Revision < curRevision {
			fmt.Println("- update -")
		}
	}
}
