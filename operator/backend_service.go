package operator

import (
	"context"
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/teseraio/ensemble/operator/proto"
)

type backendService struct {
	proto.UnimplementedBackendServiceServer

	srv *Server

	lock     sync.Mutex
	channels []chan proto.GetInstanceUpdatesResp
}

func (b *backendService) createWatcher() chan proto.GetInstanceUpdatesResp {
	b.lock.Lock()
	ch := make(chan proto.GetInstanceUpdatesResp, 10)
	b.channels = append(b.channels, ch)
	b.lock.Unlock()

	return ch
}

func (b *backendService) UpsertInstance(ctx context.Context, req *proto.UpsertInstanceReq) (*proto.UpsertInstanceResp, error) {
	i := req.Instance.Copy()

	fmt.Println("--- Upsert instance: ", i.DeploymentID, i.ID)

	if err := b.srv.State.UpsertNode(i); err != nil {
		return nil, err
	}

	b.lock.Lock()
	for _, c := range b.channels {
		select {
		case c <- proto.GetInstanceUpdatesResp{Id: i.ID, Cluster: i.DeploymentID}:
		default:
		}
	}
	b.lock.Unlock()

	return &proto.UpsertInstanceResp{}, nil
}

func (b *backendService) GetInstance(ctx context.Context, req *proto.GetInstanceReq) (*proto.GetInstanceResp, error) {
	fmt.Println("_ GET INSTANCE", req.Cluster, req.Id)

	instance, err := b.srv.State.LoadInstance(req.Cluster, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.GetInstanceResp{Instance: instance}, nil
}

func (b *backendService) GetInstanceUpdates(req *empty.Empty, stream proto.BackendService_GetInstanceUpdatesServer) error {
	ch := b.createWatcher()
	for {
		msg := <-ch
		stream.Send(&msg)
	}
	return nil
}

func (b *backendService) GetDeploymentByID(ctx context.Context, req *proto.GetDeploymentByIDReq) (*proto.GetDeploymentByIDResp, error) {
	dep, err := b.srv.State.LoadDeployment(req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.GetDeploymentByIDResp{Deployment: dep}, nil
}
