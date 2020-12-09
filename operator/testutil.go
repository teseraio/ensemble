package operator

import (
	"testing"

	"github.com/hashicorp/go-hclog"
)

func TestOperator(t *testing.T, provider Provider, factory HandlerFactory) *Server {
	config := &Config{
		Provider:         provider,
		HandlerFactories: []HandlerFactory{},
	}
	if factory != nil {
		config.HandlerFactories = append(config.HandlerFactories, factory)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "ensemble",
		Level: hclog.Info,
	})
	srv, err := NewServer(logger, config)
	if err != nil {
		t.Fatal(err)
	}
	return srv
}
