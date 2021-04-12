package operator

import (
	"context"
	"fmt"

	gproto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type service struct {
	proto.UnimplementedEnsembleServiceServer

	s *Server
}

func (s *service) Apply(ctx context.Context, component *proto.Component) (*proto.Component, error) {
	// Apply the component
	component.Id = uuid.UUID()

	fmt.Println("---xxxx--")
	fmt.Println(component)

	var spec proto.ClusterSpec
	if err := gproto.Unmarshal(component.Spec.Value, &spec); err != nil {
		panic(err)
	}

	fmt.Println("-- spec --")
	fmt.Println(spec)

	providerSpec := s.s.Provider.Resources()

	for _, grp := range spec.Groups {
		if grp.Storage == nil {
			grp.Storage = proto.EmptySpec()
		}
		if grp.Resources == nil {
			grp.Resources = proto.EmptySpec()
		}

		fmt.Println("-- grp --")
		fmt.Println(grp)
		fmt.Println(grp.Resources)
		fmt.Println(providerSpec.Resources)

		fmt.Println(providerSpec.Resources.Validate(grp.Resources))

		fmt.Println("-- storage --")
		fmt.Println(providerSpec.Storage)
		fmt.Println(grp.Storage)
		fmt.Println(providerSpec.Storage.Validate(grp.Storage))
	}

	seq, err := s.s.State.Apply(component)
	if err != nil {
		return nil, err
	}
	if seq == 0 {
		// it was not updated
		return component, nil
	}
	return component, nil
}

func (s *service) ListDeployments(ctx context.Context, _ *empty.Empty) (*proto.ListDeploymentsResp, error) {
	deps, err := s.s.State.ListDeployments()
	if err != nil {
		return nil, err
	}
	resp := &proto.ListDeploymentsResp{
		Deployments: deps,
	}
	return resp, nil
}

func (s *service) GetDeployment(ctx context.Context, req *proto.GetDeploymentReq) (*proto.Deployment, error) {
	dep, err := s.s.State.LoadDeployment(req.Cluster)
	if err != nil {
		return nil, err
	}
	return dep, nil
}
