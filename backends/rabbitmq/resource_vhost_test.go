package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/testutil"
)

func TestVHost(t *testing.T) {
	provider, _ := testutil.NewTestProvider(t, "rabbitmq", nil)

	srv := operator.TestOperator(t, provider, Factory)
	defer srv.Stop()

	uuid := provider.Apply(&testutil.TestTask{
		Name:  "A",
		Input: `{"replicas": 1}`,
	})
	provider.WaitForTask(uuid)

	// create the vhost
	uuid = provider.Apply(&testutil.TestTask{
		Name:     "B",
		Resource: "VHost",
		Input: `{
			"name": "B"
		}`,
	})
	provider.WaitForTask(uuid)
}
