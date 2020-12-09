package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/testutil"
)

func TestExchange(t *testing.T) {
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
		Name:     "V",
		Resource: "VHost",
		Input: `{
			"name": "v"
		}`,
	})
	provider.WaitForTask(uuid)

	// create the exchange
	uuid = provider.Apply(&testutil.TestTask{
		Name:     "B",
		Resource: "Exchange",
		Input: `{
			"name": "B",
			"vhost": "v",
			"settings": {
				"type": "fanout"
			}
		}`,
	})
	provider.WaitForTask(uuid)
}
