package operator

import (
	"context"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type service struct {
	s *Server
}

func (s *service) Apply(ctx context.Context, component *proto.Component) (*proto.Component, error) {
	// generate the random id for this version
	component.Id = uuid.UUID()

	// TODO: Validate
	if err := s.s.State.Apply(component); err != nil {
		return nil, err
	}
	return component, nil
}
