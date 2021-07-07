
## 0.1.3 (Unreleased)

FEATURES:

- backends: add VictoriaMetrics backend.
- backends: add Kafka backend.
- operator/schema: Use a new schema format to represent database and provider configuration and resources.
- operator/cluster: handle resource configuration (cpu, mem...) defined by the Provider for nodes in the cluster.
- operator/cluster: handle configuration for nodes in the cluster.

## 0.1.2 (February 8, 2021)

FEATURES:

- backends: add Dask backend.
- operator/cluster: introduce group of nodes in the cluster.
- operator/resource: a resource can include now a Init function to pre-validate the input.
- operator/resource: you can now add a 'required' tag to any field in the Resource to make it mandatory.

## 0.1.1 (January 25, 2021)

FEATURES:

- Use Boltdb to store the state

## 0.1.0 (December 9, 2020)

Initial Public Release
