package internal

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

// ParseObjectSchemaJSON unmarshals the user-provided JSON string into a
// v1pb.ObjectSchema. Returns an error with a user-friendly message if the
// JSON is malformed or does not match the proto shape.
func ParseObjectSchemaJSON(raw string) (*v1pb.ObjectSchema, error) {
	if raw == "" {
		return nil, nil
	}
	var schema v1pb.ObjectSchema
	opts := protojson.UnmarshalOptions{DiscardUnknown: false}
	if err := opts.Unmarshal([]byte(raw), &schema); err != nil {
		return nil, errors.Wrap(err, "invalid object_schema_json")
	}
	return &schema, nil
}

// MarshalObjectSchemaToJSON serializes an ObjectSchema into a canonical
// JSON string. We route through encoding/json after protojson so map keys
// are sorted and whitespace is stripped — protojson itself does not
// guarantee deterministic map-key order.
func MarshalObjectSchemaToJSON(schema *v1pb.ObjectSchema) (string, error) {
	if schema == nil {
		return "", nil
	}
	pj, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(schema)
	if err != nil {
		return "", errors.Wrap(err, "protojson marshal")
	}
	return canonicalizeJSON(pj)
}

// NormalizeObjectSchemaJSON is the round-trip: parse user JSON through the
// proto type (type-aware validation) and emit the canonical form. Used by
// StateFunc so the same value stored in state matches what we'd read back
// from the server.
func NormalizeObjectSchemaJSON(raw string) (string, error) {
	schema, err := ParseObjectSchemaJSON(raw)
	if err != nil {
		return "", err
	}
	return MarshalObjectSchemaToJSON(schema)
}

// canonicalizeJSON reparses raw JSON into a generic value and re-marshals
// with sorted map keys. encoding/json emits map keys in sorted order when
// marshaling map[string]any — we exploit that.
func canonicalizeJSON(raw []byte) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", errors.Wrap(err, "canonicalize: reparse")
	}
	v = sortMapKeys(v)
	out, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "canonicalize: remarshal")
	}
	return string(out), nil
}

// sortMapKeys walks the decoded value and replaces any map[string]any with
// a value whose marshaling order is deterministic. encoding/json already
// sorts map[string]any keys, but we still need to descend into nested
// arrays and maps to cover the full tree.
func sortMapKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(t))
		for _, k := range keys {
			out[k] = sortMapKeys(t[k])
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = sortMapKeys(e)
		}
		return out
	default:
		return v
	}
}
