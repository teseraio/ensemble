package k8s

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/testutil"
)

func TestK8sProviderSpec(t *testing.T) {
	p, err := K8sFactory(hclog.NewNullLogger(), nil)
	if err != nil {
		t.Fatal(err)
	}
	testutil.TestProvider(t, p)
}

func TestK8sClient_Error(t *testing.T) {
	cases := []struct {
		obj string
		err error
	}{
		{
			obj: `{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Failure","message":"The resourceVersion for the provided list is too old.","reason":"Expired","code":410}`,
			err: errExpired,
		},
		{
			obj: `{"type":"ERROR","object":{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Failure","message":"too old resource version: 10000 (343914)","reason":"Expired","code":410}}`,
			err: errExpired,
		},
	}

	for _, c := range cases {
		assert.ErrorIs(t, isError([]byte(c.obj)), c.err)
	}
}

func TestDecode_GetPod(t *testing.T) {
	readFixture := func(name string) *PodItem {
		data, err := ioutil.ReadFile("./resources/fixtures/" + name + ".json")
		assert.NoError(t, err)

		item := &PodItem{}
		assert.NoError(t, json.Unmarshal(data, &item))

		return item
	}

	t.Run("Pending", func(t *testing.T) {
		item := readFixture("api_pod_get_pending")
		assert.Equal(t, item.Status.Phase, PodPhasePending)
		assert.NotEmpty(t, item.Status.ContainerStatuses[0].State.Waiting.Reason)
	})

	t.Run("Running", func(t *testing.T) {
		item := readFixture("api_pod_get_running")
		assert.Equal(t, item.Status.Phase, PodPhaseRunning)
		assert.NotEmpty(t, item.Status.ContainerStatuses[0].State.Running.StartedAt)
	})

	t.Run("Terminated_BadArgs", func(t *testing.T) {
		item := readFixture("api_pod_get_terminated_badargs")
		assert.Equal(t, item.Status.Phase, PodPhaseFailed)

		exitCode, err := item.ExitResult()
		assert.NoError(t, err)
		assert.Equal(t, exitCode, &proto.Instance_ExitResult{
			Code:  128,
			Error: "StartError: failed to create containerd task",
		})
	})

	t.Run("Terminated_Completed", func(t *testing.T) {
		item := readFixture("api_pod_get_terminated_completed")
		assert.Equal(t, item.Status.Phase, PodPhaseSucceeded)

		exitCode, err := item.ExitResult()
		assert.NoError(t, err)
		assert.Equal(t, exitCode, &proto.Instance_ExitResult{Code: 0})
	})
}
