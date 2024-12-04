package client

import (
	"google.golang.org/protobuf/encoding/protojson"
)

// ProtojsonUnmarshaler is the unmarshal for protocol.
var ProtojsonUnmarshaler = protojson.UnmarshalOptions{DiscardUnknown: true}
