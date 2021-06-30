SHELL := /bin/bash

build-releases:
	./scripts/build.sh

build-dev-docker:
	rm -rf ./bin
	go build --mod=vendor -o ./bin/ensemble main.go
	docker build -t ensemble:dev .

protoc:
	protoc --go_out=. --go-grpc_out=. ./operator/proto/*.proto

bindata:
	go generate ./k8s
	go generate ./backends/clickhouse
