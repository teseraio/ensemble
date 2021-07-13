package testutil

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/operator/state/boltdb"
	"google.golang.org/grpc"
)

func IsE2EEnabled(t *testing.T) {
	if os.Getenv("E2E_ENABLED") != "true" {
		t.Skip("Tests only enabled in e2e mode")
	}
}

var testPortRangeBegin = uint64(6000)

type TestServer struct {
	t      *testing.T
	srv    *operator.Server
	state  *boltdb.BoltDB
	path   string
	docker *Client
	clt    proto.EnsembleServiceClient
}

func (t *TestServer) GetDeployment(name string) *proto.Deployment {
	deps, err := t.state.ListDeployments()
	if err != nil {
		t.t.Fatal(err)
	}
	for _, dep := range deps {
		if dep.Name == name {
			d, err := t.state.LoadDeployment(dep.Id)
			if err != nil {
				t.t.Fatal(err)
			}
			return d
		}
	}
	t.t.Fatal("deployment not found")
	return nil
}

func (t *TestServer) Remove(id string) error {
	return t.docker.Remove(id)
}

func (t *TestServer) Apply(c *proto.Component) string {
	cc, err := t.clt.Apply(context.Background(), c)
	if err != nil {
		t.t.Fatal(err)
	}
	return cc.Id
}

func (t *TestServer) WaitForTask(id string) {
	ch := t.state.Wait(id)
	<-ch
}

func (t *TestServer) Close() {
	t.srv.Stop()
	t.docker.Clean()
	if err := os.Remove(t.path); err != nil {
		t.t.Fatal(err)
	}
}

func TestOperator(t *testing.T, factories ...operator.HandlerFactory) *TestServer {
	path := "/tmp/db-" + uuid.UUID()

	state, err := boltdb.Factory(map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}

	provider, err := NewDockerClient()
	if err != nil {
		t.Fatal(err)
	}

	grpcPort := atomic.AddUint64(&testPortRangeBegin, 1)
	grpcAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: int(grpcPort)}

	config := &operator.Config{
		Provider:         provider,
		State:            state,
		HandlerFactories: []operator.HandlerFactory{},
		GRPCAddr:         grpcAddr,
	}
	for _, factory := range factories {
		config.HandlerFactories = append(config.HandlerFactories, factory)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "ensemble",
		Level: hclog.Debug,
	})
	srv, err := operator.NewServer(logger, config)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := grpc.Dial(grpcAddr.String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	tt := &TestServer{
		t:      t,
		srv:    srv,
		state:  state,
		docker: provider,
		path:   path,
		clt:    proto.NewEnsembleServiceClient(conn),
	}
	return tt
}
