package operator

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type service struct {
	s *Server
}

func (s *service) Apply(ctx context.Context, component *proto.Component) (*proto.Component, error) {
	// generate the random id for this version
	component.Id = uuid.UUID()

	// Apply the component
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
	// TODO
	return nil, nil
}

func (s *service) GetDeployment(ctx context.Context, req *proto.GetDeploymentReq) (*proto.Deployment, error) {
	dep, err := s.s.State.LoadDeployment(req.Cluster)
	if err != nil {
		return nil, err
	}
	return dep, nil
}
