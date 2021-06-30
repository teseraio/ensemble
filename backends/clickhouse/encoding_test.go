package clickhouse

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestX(t *testing.T) {
	obj := Cluster{
		Shards: []*Shard{
			{
				Replicas: []*Replica{
					{
						Host: "h",
						Port: 1234,
					},
					{
						Host: "hh",
						Port: 5678,
					},
				},
			},
		},
	}
	res, err := runTmpl("cluster", obj)
	assert.NoError(t, err)
	fmt.Println(string(res))
}
