package operator

import (
	"fmt"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/operator/proto"
)

type ResourceScheduler struct {
	state schedState
}

func (r *ResourceScheduler) Process(eval *proto.Evaluation) (*proto.Plan, error) {
	dep, err := r.state.LoadDeployment(eval.DeploymentID)
	if err != nil {
		return nil, err
	}
	handler, err := r.state.GetHandler(dep.Backend)
	if err != nil {
		return nil, err
	}

	comp, err := r.state.GetComponentByID(eval.DeploymentID, eval.ComponentID, eval.Sequence)
	if err != nil {
		return nil, err
	}
	msg, err := proto.UnmarshalAny(comp.Spec)
	if err != nil {
		return nil, err
	}
	spec, ok := msg.(*proto.ResourceSpec)
	if !ok {
		return nil, fmt.Errorf("component is not of Resource type")
	}

	schema, ok := handler.GetSchemas().Resources[spec.Resource]
	if !ok {
		return nil, fmt.Errorf("resource '%s' not found", spec.Resource)
	}
	if err := schema.Validate(spec.Params); err != nil {
		return nil, fmt.Errorf("failed to validate Resource schema: %v", err)
	}

	if comp.Sequence != 1 {
		pastComp, err := r.state.GetComponentByID(eval.DeploymentID, eval.ComponentID, eval.Sequence-1)
		if err != nil {
			return nil, err
		}
		var oldSpec proto.ResourceSpec
		if err := gproto.Unmarshal(pastComp.Spec.Value, &oldSpec); err != nil {
			return nil, err
		}

		diff := schema.Diff(spec.Params, oldSpec.Params)

		// check if any of the diffs requires force-new
		forceNew := false
		for name := range diff {
			field, err := schema.Get(name)
			if err != nil {
				return nil, err
			}
			if field.ForceNew {
				forceNew = true
			}
		}
		if forceNew {
			req := &ApplyResourceRequest{
				Deployment: dep,
				Action:     ApplyResourceRequestDelete,
				Resource:   spec,
			}
			if err := handler.ApplyResource(req); err != nil {
				return nil, err
			}
		}
	}

	var action string
	if comp.Action == proto.Component_DELETE {
		action = ApplyResourceRequestDelete
	} else {
		action = ApplyResourceRequestReconcile
	}
	req := &ApplyResourceRequest{
		Deployment: dep,
		Action:     action,
		Resource:   spec,
	}
	if err := handler.ApplyResource(req); err != nil {
		return nil, err
	}

	plan := &proto.Plan{
		Done: true,
	}
	return plan, nil
}
