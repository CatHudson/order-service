package protoxjson

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var defaultUnmarshaler = NewUnmarshaler() //nolint: gochecknoglobals // global by design

// NewUnmarshaler returns configured protojson unmarshaler.
func NewUnmarshaler() protojson.UnmarshalOptions {
	return protojson.UnmarshalOptions{
		// Without this option we lose compatibility.
		DiscardUnknown: true,
	}
}

// Unmarshal unmarshals the given JSON data into the given proto.Message using
// default unmarshaler.
func Unmarshal(b []byte, m proto.Message) error {
	return defaultUnmarshaler.Unmarshal(b, m)
}
