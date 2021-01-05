package operator

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct {
	s *Server
}

func (s *service) UpsertCluster(ctx context.Context, req *proto.ClusterSpec) (*empty.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *service) UpsertResource(ctx context.Context, req *proto.ResourceSpec) (*empty.Empty, error) {
	return &emptypb.Empty{}, nil
}
