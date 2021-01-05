package state

import "testing"

type setupFn func(*testing.T) (State, func())

// TestSuite has a suite of tests for state implementations
func TestSuite(t *testing.T, setup setupFn) {

}
