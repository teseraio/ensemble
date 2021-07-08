package clickhouse

import (
	"fmt"

	"github.com/teseraio/ensemble/lib/template"
)

//go:generate go-bindata -pkg clickhouse -o ./bindata.go ./resources

func runTmpl(name string, obj interface{}) ([]byte, error) {
	content, err := Asset(fmt.Sprintf("resources/%s.template", name))
	if err != nil {
		return nil, err
	}
	return template.RunTmpl(string(content), obj)
}

type Cluster struct {
	Name      string
	Shards    []*Shard
	Zookeeper string
}

type Shard struct {
	Replicas []*Replica
}

type Replica struct {
	Host string
	Port uint64
}
