package server

import (
	"github.com/teseraio/ensemble/backends/cassandra"
	"github.com/teseraio/ensemble/backends/dask"
	"github.com/teseraio/ensemble/backends/kafka"
	"github.com/teseraio/ensemble/backends/rabbitmq"
	"github.com/teseraio/ensemble/backends/victoriametrics"
	"github.com/teseraio/ensemble/backends/zookeeper"
	"github.com/teseraio/ensemble/operator"
)

// BuiltinBackends is the list of available builtin backends
var BuiltinBackends []operator.HandlerFactory

func registerBackend(factory operator.HandlerFactory) {
	if len(BuiltinBackends) == 0 {
		BuiltinBackends = []operator.HandlerFactory{}
	}
	BuiltinBackends = append(BuiltinBackends, factory)
}

func init() {
	registerBackend(dask.Factory)
	registerBackend(rabbitmq.Factory)
	registerBackend(cassandra.Factory)
	registerBackend(zookeeper.Factory)
	registerBackend(kafka.Factory)
	registerBackend(victoriametrics.Factory)
}
