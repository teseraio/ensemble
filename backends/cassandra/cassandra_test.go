package cassandra

import (
	"testing"

	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/testutil"
)

func TestCassandraBootstrap(t *testing.T) {
	provider, _ := testutil.NewTestProvider(t, "cassandra", nil)

	srv := operator.TestOperator(t, provider, Factory)
	defer srv.Stop()

	uuid := provider.Apply(&testutil.TestTask{
		Name:  "A",
		Input: `{"replicas": 2}`,
	})
	provider.WaitForTask(uuid)
}
