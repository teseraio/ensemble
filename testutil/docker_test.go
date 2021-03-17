package testutil

import (
	"testing"
)

func TestDockerProviderSpec(t *testing.T) {
	p, err := NewDockerClient()
	if err != nil {
		t.Fatal(err)
	}
	TestProvider(t, p)
}
