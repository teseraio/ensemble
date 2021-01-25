package proto

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const urlPrefix = "ensembleoss.io/"

func MarshalAny(m proto.Message) (*any.Any, error) {
	b, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &any.Any{TypeUrl: urlPrefix + proto.MessageName(m), Value: b}, nil
}

func MustMarshalAny(m proto.Message) *any.Any {
	b, err := MarshalAny(m)
	if err != nil {
		panic(err)
	}
	return b
}
