package operator

import (
	"context"

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

	if err := s.s.validateComponent(component); err != nil {
		return nil, err
	}
	component, err := s.s.State.Apply(component)
	if err != nil {
		return nil, err
	}
	if component == nil {
		return &proto.Component{}, nil
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
