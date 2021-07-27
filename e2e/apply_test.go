package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/grpc"
)

func TestE2E_Apply(t *testing.T) {
	_, err := kubectl("apply -f ../examples/dask-simple.yaml")
	assert.NoError(t, err)

	clt := newClient(t)

	getDeployment := func() *proto.Deployment {
		resp, err := clt.ListDeployments(context.Background(), &empty.Empty{})
		assert.NoError(t, err)

		for _, dep := range resp.Deployments {
			if dep.Name == "dask-simple" {
				return dep
			}
		}
		return nil
	}

	for i := 0; i < 60; i++ {
		dep := getDeployment()
		fmt.Println(dep)
		if dep.Status == proto.DeploymentDone {
			return
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatal("timeout")
}

func newClient(t *testing.T) proto.EnsembleServiceClient {
	conn, err := grpc.Dial("127.0.0.1:6001", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	clt := proto.NewEnsembleServiceClient(conn)
	return clt
}

func kubectl(args string) (string, error) {
	cmd := exec.Command("kubectl", strings.Split(args, " ")...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	cmd.Stdout = &outBuf
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to exec '%s': %s", err.Error(), errBuf.String())
	}
	return outBuf.String(), nil
}
