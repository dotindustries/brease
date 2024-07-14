package code

import (
	"github.com/d5/tengo/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
)

type ProtoStruct struct {
	tengo.ObjectImpl
	Value *structpb.Struct
}

func (o *ProtoStruct) String() string {
	if o.Value == nil {
		return "(empty)"
	}
	bts, err := proto.Marshal(o.Value)
	if err != nil {
		return "(error: " + err.Error() + ")"
	}
	return strings.TrimSuffix(string(bts), "\n")
}

func (o *ProtoStruct) TypeName() string {
	return "proto.Struct"
}
