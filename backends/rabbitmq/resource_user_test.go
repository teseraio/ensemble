package rabbitmq

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/testutil"
)

func TestUser(t *testing.T) {
	provider, _ := testutil.NewTestProvider(t, "rabbitmq", nil)

	srv := operator.TestOperator(t, provider, Factory)
	defer srv.Stop()

	uuid := provider.Apply(&testutil.TestTask{
		Name:  "A",
		Input: `{"replicas": 1}`,
	})
	provider.WaitForTask(uuid)

	// create the user
	uuid = provider.Apply(&testutil.TestTask{
		Name:     "B",
		Resource: "User",
		Input: `{
			"username": "B",
			"password": "xxx"
		}`,
	})
	provider.WaitForTask(uuid)

	// update the password
	uuid = provider.Apply(&testutil.TestTask{
		Name:     "B",
		Resource: "User",
		Input: `{
			"username": "B",
			"password": "yyy"
		}`,
	})
	provider.WaitForTask(uuid)
}
