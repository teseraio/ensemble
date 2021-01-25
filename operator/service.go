package operator

import (
	"context"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type service struct {
	s *Server
}

func (s *service) Apply(ctx context.Context, req *proto.ApplyReq) (*proto.Component, error) {
	component := req.Component

	// Validate component
	if err := s.s.validate(req.Component); err != nil {
		return nil, err
	}
	if req.DryRun {
		return component, nil
	}

	// generate the random id for this version and store
	component.Id = uuid.UUID()

	if err := s.s.State.Apply(component); err != nil {
		return nil, err
	}
	return component, nil
}
