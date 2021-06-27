package proto

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/runtime/protoiface"
)

func (c *Component) Type() string {
	switch c.Spec.TypeUrl {
	case "proto.ClusterSpec":
		return EvaluationTypeCluster
	case "proto.ResourceSpec":
		return EvaluationTypeResource
	}
	panic("Not found")
}

func UnmarshalAny(a *any.Any) (proto.Message, error) {
	var obj protoiface.MessageV1
	switch a.TypeUrl {
	case "proto.ClusterSpec":
		obj = &ClusterSpec{}
	case "proto.ResourceSpec":
		obj = &ResourceSpec{}
	default:
		return nil, fmt.Errorf("unknown type %s", a.TypeUrl)
	}
	if err := proto.Unmarshal(a.Value, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func MustUnmarshalAny(a *any.Any) proto.Message {
	msg, err := UnmarshalAny(a)
	if err != nil {
		panic(err)
	}
	return msg
}

func MarshalAny(m proto.Message) (*any.Any, error) {
	name := proto.MessageName(m)
	b, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return &any.Any{TypeUrl: name, Value: b}, nil
}

func MustMarshalAny(m proto.Message) *any.Any {
	b, err := MarshalAny(m)
	if err != nil {
		panic(err)
	}
	return b
}

func Cmp(a, b *any.Any) (bool, error) {
	msg1, err := UnmarshalAny(a)
	if err != nil {
		return false, err
	}
	msg2, err := UnmarshalAny(b)
	if err != nil {
		return false, err
	}
	return proto.Equal(msg1, msg2), nil
}
