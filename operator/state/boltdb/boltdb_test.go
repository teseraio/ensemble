package boltdb

import (
	"os"
	"testing"

	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/state"
)

func setupFn(t *testing.T) (state.State, func()) {
	path := "/tmp/db-" + uuid.UUID()

	st, err := Factory(map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}
	closeFn := func() {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	return st, closeFn
}

func TestSuite(t *testing.T) {
	state.TestSuite(t, setupFn)
}
