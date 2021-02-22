package operator

import (
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state"
)

// Deployment keeps track of the instances running in a specific deployment
type Deployment struct {
	s *Server
	p Provider

	dep  *proto.Deployment
	spec *proto.ClusterSpec

	instances map[string]*nodeRunner
	state     state.State

	// status of the deployment
	status string
	stopCh chan struct{}

	allocUpdates chan struct{}
}

func (d *Deployment) setup() {
	d.instances = map[string]*nodeRunner{}
}

func (d *Deployment) getResult(r *reconcileResult) {
	// this result could change the deployment in memory
	fmt.Println("- res -")
	fmt.Println(r)

	d.dep.Groups = map[string]*proto.Deployment_GroupState{
		"": r.state,
	}

	// HOW DO WE HANDLE ERRORS OUTSIDE THIS LOOP?
	// FOR EXAMPLE WHEN WE STOP A NODE
	// OR IF THE ROLLOUT FAILS?

	// add all the nodes first
	for _, add := range r.add {
		d.Update(add)
	}

	// this is a special case
	for _, node := range r.update {
		// first try to remove the other one
		d.instances[node.Name].remove()

		// run the other one
		d.instances[node.Name].notify(node)
	}

	/*
		// this is the reconcile loop
		for {
			// pick the next things to be allocated
			if !d.handleAllocs() {
				// this allocation is done
				return
			}

			// wait for the results to be ready
			for {
				select {
				case <-d.allocUpdates:
				}

				// check that the things we expected happened
				ready := d.xx()
				if ready {

				}
			}
		}
	*/
}

func (d *Deployment) handleAllocs() bool {
	return false
}

func (d *Deployment) run() {
	d.instances = map[string]*nodeRunner{}
	d.allocUpdates = make(chan struct{})

	for {
		select {
		case <-d.allocUpdates:
			// read to check if there is anything wrong right now

			ready := d.xx()
			if ready {
				// do something, maybe create a new eval
				fmt.Println("_ IS READY _")
			}

		case <-d.stopCh:
		}
	}
}

func (d *Deployment) instancesByGroup(name string) []*proto.Instance {
	res := []*proto.Instance{}
	for _, i := range d.instances {
		if i.instance.Group == name {
			res = append(res, i.instance)
		}
	}
	return res
}

func (d *Deployment) xx() bool {
	// check all the stuff going on
	// get all groups
	for _, grp := range d.spec.Groups {
		instances := d.instancesByGroup(grp.Name)

		// check if the instances are running or pending or what

		state := d.dep.Groups[grp.Name]

		count := int64(0)
		for _, i := range instances {
			if i.Status == proto.Instance_RUNNING {
				count++
			}
		}
		if state.Desired != count {
			return false
		}
	}
	return true
}

func (d *Deployment) notifyUpdate(op *proto.InstanceUpdate) {
	for _, i := range d.instances {
		if i.instance.ID == op.ID {
			i.notifyUpdate(op)
		}
	}
}

// Update is used to notify updates in the instances of the deployment
func (d *Deployment) Update(instance *proto.Instance) {
	id := instance.Name

	if _, ok := d.instances[id]; !ok {
		d.instances[id] = &nodeRunner{instance: instance, p: d.p, s: d.s, allocUpdates: d.allocUpdates}
		go d.instances[id].run()
	}
	d.instances[id].notify(instance)
}

// run hooks

type nodeRunner struct {
	// this one tracks the state of the node, with the corresponding instance
	instance     *proto.Instance
	p            Provider
	s            *Server
	allocUpdates chan struct{}
}

func (i *nodeRunner) run() {

}

func (i *nodeRunner) notify(instance *proto.Instance) {
	i.instance = instance

	// check the state
	if instance.Status == proto.Instance_PENDING {
		i.p.CreateResource(instance)
	}
}

func (i *nodeRunner) remove() {
	i.p.DeleteResource(i.instance)
}

func (i *nodeRunner) notifyUpdate(op *proto.InstanceUpdate) {

	switch obj := op.Event.(type) {
	case *proto.InstanceUpdate_Conf:
		// this is like a success so use it like that
		i.instance.Status = proto.Instance_RUNNING
		i.instance.Ip = obj.Conf.Ip
		i.instance.Handler = obj.Conf.Handler

	case *proto.InstanceUpdate_Status:
		// failed
		i.instance.Status = proto.Instance_FAILED
	}

	if err := i.s.State.UpsertNode(i.instance); err != nil {
		panic(err)
	}

	// and create a new evaluation
	i.allocUpdates <- struct{}{}
}
