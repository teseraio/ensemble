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

var testPortRangeBegin = uint64(6000)

type TestServer struct {
	t      *testing.T
	srv    *operator.Server
	state  *boltdb.BoltDB
	path   string
	docker *Client
	clt    proto.EnsembleServiceClient
}

type taskWrap struct {
	id   string
	name string
}

func (t *TestServer) Apply(c *proto.Component) string {
	cc, err := t.clt.Apply(context.Background(), c)
	if err != nil {
		t.t.Fatal(err)
	}
	return cc.Id
}

func (t *TestServer) Destroy(i int) {
	t.docker.Destroy(i)
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

func TestOperator(t *testing.T, factory operator.HandlerFactory) *TestServer {
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
	if factory != nil {
		config.HandlerFactories = append(config.HandlerFactories, factory)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "ensemble",
		Level: hclog.Info,
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
		state:  state.(*boltdb.BoltDB),
		docker: provider,
		path:   path,
		clt:    proto.NewEnsembleServiceClient(conn),
	}
	return tt
}
