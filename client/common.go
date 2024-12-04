package client

import (
	"google.golang.org/protobuf/encoding/protojson"
)

var ProtojsonUnmarshaler = protojson.UnmarshalOptions{DiscardUnknown: true}
