package protoxjson

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var defaultMarshaler = NewMarshaler() //nolint: gochecknoglobals // global by design

// NewMarshaler returns configured protojson marshaler.
func NewMarshaler() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		// We use snake_case field names instead of camelCase names.
		UseProtoNames: true,
		// We want full messages, not partial.
		EmitUnpopulated: true,
	}
}

// Marshal marshals the given proto.Message in the JSON format using default
// marshaler.
func Marshal(m proto.Message) ([]byte, error) {
	return defaultMarshaler.Marshal(m)
}
