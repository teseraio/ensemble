package operator

import (
	"fmt"
	"strconv"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type updateReconciler interface {
	EvaluateConfig(spec *proto.NodeSpec, config map[string]string) error
	Initialize(grp *proto.Group, n []*proto.Instance, target *proto.Instance) (*proto.NodeSpec, error)
}

// reconciler does a first reconcilization and determines which nodes have to be deployed or changed
// and creates a plan that describes this changes
type allocReconciler struct {
	name       string
	c          *proto.ClusterSpec
	dep        *proto.Deployment
	result     *reconcileResult
	reconciler updateReconciler
}

type reconcileResult struct {
	name string

	// whether this group is being removed
	delete bool

	// scale up nodes
	add []*proto.Instance

	// scale down nodes
	del []*proto.Instance

	// list of update instances
	update []*proto.Instance

	// failed instances also need to update
	failed []*proto.Instance

	state *proto.Deployment_GroupState
}

func (a *allocReconciler) createAllocMatrix() map[string][]*proto.Instance {
	allocMatrix := map[string][]*proto.Instance{}
	for _, node := range a.dep.Instances {
		set := node.Group
		if _, ok := allocMatrix[set]; !ok {
			allocMatrix[set] = []*proto.Instance{}
		}
		allocMatrix[set] = append(allocMatrix[set], node)
	}
	for _, group := range a.c.Groups {
		if _, ok := allocMatrix[group.Name]; !ok {
			allocMatrix[group.Name] = []*proto.Instance{}
		}
	}
	return allocMatrix
}

func (a *allocReconciler) reconcile() {
	allocMatrix := a.createAllocMatrix()

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

	config := grp.Config

	// get the initial spec for the whole group

	reconcile = &reconcileResult{
		name:   group,
		add:    []*proto.Instance{},
		failed: []*proto.Instance{},
		update: []*proto.Instance{},
	}

	numInstances := int64(len(instances))
	isScale := grp.Count != numInstances
	if isScale {
		if grp.Count > numInstances {
			// scale up
			for i := numInstances; i < grp.Count; i++ {
				node := &proto.Instance{
					ID:      uuid.UUID(),
					Name:    a.name + "_" + strconv.Itoa(int(i)),
					Cluster: a.name,
					Status:  proto.Instance_PENDING,
					Index:   i,
				}
				fmt.Println("-- ...")
				reconcile.add = append(reconcile.add, node)
			}
		} else {
			// scale down
		}
	}

	reconcile.state = &proto.Deployment_GroupState{
		Desired: grp.Count,
	}

	// the nodes we are about to add we need to initialize them
	// split with the other ones so that we can start this after all the groups are done

	// add also the nodes part of the cluster already
	nn := []*proto.Instance{}
	for _, n := range a.dep.Instances {
		nn = append(nn, n)
	}
	for _, n := range reconcile.add {
		nn = append(nn, n)
	}

	for _, n := range reconcile.add {
		if a.reconciler != nil {
			a.reconciler.Initialize(nil, nn, n)

			// evaluate the config
			if err := a.reconciler.EvaluateConfig(n.Spec, config); err != nil {
				panic(err)
			}
		}
	}

	// check failed nodes
	for _, instance := range instances {
		if instance.Status == proto.Instance_FAILED {
			if instance.Count < 2 {
				node := &proto.Instance{
					ID:      uuid.UUID(),
					Name:    instance.Name,
					Cluster: a.name,
					Status:  proto.Instance_PENDING,
					Count:   instance.Count + 1,
					Spec:    instance.Spec,
				}

				reconcile.failed = append(reconcile.failed, instance)

				fmt.Printf("Restart instance %s %d\n", node.Name, node.Count)
				reconcile.add = append(reconcile.add, node)
			} else {
				fmt.Println("=================================> OUT")
				// do not do anything
			}
		}
	}

	oneFailed := false
	for _, instance := range instances {
		if instance.Status == proto.Instance_FAILED {
			oneFailed = true
		}
	}
	if !oneFailed {
		// the deployment is ready, evaluation is done
	}

	// check if there is any update
	curRevision := a.c.Generation

	revision := 0
	for _, instance := range instances {
		if instance.Revision < curRevision {
			// update to the next revision
			revision++

			node := &proto.Instance{
				ID:      uuid.UUID(),
				Name:    instance.Name,
				Cluster: a.name,
				Status:  proto.Instance_PENDING,
				Spec:    instance.Spec,
				Prev:    instance.ID,
			}
			if err := a.reconciler.EvaluateConfig(node.Spec, config); err != nil {
				panic(err)
			}
			reconcile.update = append(reconcile.update, node)
		}
	}

	reconcile.state.Update = int64(revision)

	return
}
