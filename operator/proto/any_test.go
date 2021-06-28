package proto

import (
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestEncodeDecodeAny(t *testing.T) {
	cases := []proto.Message{
		&ClusterSpec{
			Version: "",
		},
	}
	for _, c := range cases {
		MarshalAny(c)
	}
}

func TestCmp(t *testing.T) {
	a := MustMarshalAny(&ResourceSpec{
		Cluster:  "rc1",
		Resource: "User",
		Params: BlockSpec(&Spec_Block{
			Attrs: map[string]*Spec{
				"username": LiteralSpec(&Spec_Literal{
					Value: "a",
				}),
				"password": LiteralSpec(&Spec_Literal{
					Value: "b",
				}),
			},
		}),
	})
	b := MustMarshalAny(&ResourceSpec{
		Cluster:  "rc1",
		Resource: "User",
		Params: BlockSpec(&Spec_Block{
			Attrs: map[string]*Spec{
				"username": LiteralSpec(&Spec_Literal{
					Value: "a",
				}),
				"password": LiteralSpec(&Spec_Literal{
					Value: "b",
				}),
			},
		}),
	})
	Cmp(a, b)
}
