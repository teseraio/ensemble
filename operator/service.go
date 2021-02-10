package operator

import (
	"context"
	"fmt"

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

	fmt.Println("-- seq --")
	fmt.Println(seq)

	// TODO: Validate and check which type of object it is
	// Right now only works for clusters

	// create an evaluation
	eval := &proto.Evaluation{
		Id:          uuid.UUID(),
		Status:      proto.Evaluation_PENDING,
		TriggeredBy: proto.Evaluation_SPECCHANGE,
		ClusterID:   component.Name,
		Generation:  seq,
	}
	if err := s.s.State.AddEvaluation(eval); err != nil {
		return nil, err
	}
	return component, nil
}
