package operator

import (
	"fmt"

	gproto "github.com/golang/protobuf/proto"
	"github.com/teseraio/ensemble/operator/proto"
)

type resourceScheduler struct {
	state schedState
}

func (r *resourceScheduler) Process(eval *proto.Evaluation) error {

	// get the required data
	comp, err := r.state.GetComponentByID(eval.CompId)
	if err != nil {
		return err
	}
	dep, err := r.state.LoadDeployment(eval.ClusterID)
	if err != nil {
		return err
	}
	handler, err := r.state.GetHandler(dep.Backend)
	if err != nil {
		return err
	}

	// decode resource spec
	var spec proto.ResourceSpec
	if err := gproto.Unmarshal(comp.Spec.Value, &spec); err != nil {
		return err
	}

	// validate schema (TODO: Validate at the apply state)
	schema, ok := handler.GetSchemas().Resources[spec.Resource]
	if !ok {
		return fmt.Errorf("resource not found %s", spec.Resource)
	}
	if err := schema.Validate(spec.Params); err != nil {
		return err
	}

	if comp.Sequence != 1 {
		pastComp, err := r.state.GetComponentByID(comp.PrevId)
		if err != nil {
			return err
		}
		var oldSpec proto.ResourceSpec
		if err := gproto.Unmarshal(pastComp.Spec.Value, &oldSpec); err != nil {
			return err
		}

		diff := schema.Diff(spec.Params, oldSpec.Params)

		// check if any of the diffs requires force-new
		forceNew := false
		for name := range diff {
			field, err := schema.Get(name)
			if err != nil {
				return err
			}
			if field.ForceNew {
				forceNew = true
			}
		}
		if forceNew {
			req := &ApplyResourceRequest{
				Deployment: dep,
				Action:     ApplyResourceRequestDelete,
				Resource:   &spec,
			}
			if err := handler.ApplyResource(req); err != nil {
				return err
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
		Resource:   &spec,
	}
	handler.ApplyResource(req)

	return nil
}
