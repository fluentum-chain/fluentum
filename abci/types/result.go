package types

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

const (
	CodeTypeOK uint32 = 0
)

// IsOK returns true if Code is OK.
func IsOK(code uint32) bool {
	return code == CodeTypeOK
}

// IsErr returns true if Code is something other than OK.
func IsErr(code uint32) bool {
	return code != CodeTypeOK
}

//---------------------------------------------------------------------------
// JSON marshaling helpers for protobuf types

var (
	jsonMarshaller = jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: true,
	}
	jsonUnmarshaller = jsonpb.Unmarshaler{}
)

// MarshalJSON marshals a protobuf message to JSON
func MarshalJSON(msg interface{}) ([]byte, error) {
	if protoMsg, ok := msg.(proto.Message); ok {
		s, err := jsonMarshaller.MarshalToString(protoMsg)
		return []byte(s), err
	}
	return nil, fmt.Errorf("message does not implement proto.Message")
}

// UnmarshalJSON unmarshals JSON to a protobuf message
func UnmarshalJSON(b []byte, msg interface{}) error {
	if protoMsg, ok := msg.(proto.Message); ok {
		reader := bytes.NewBuffer(b)
		return jsonUnmarshaller.Unmarshal(reader, protoMsg)
	}
	return fmt.Errorf("message does not implement proto.Message")
}
