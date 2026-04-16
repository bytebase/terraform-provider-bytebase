package internal

import (
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func TestNormalizeObjectSchemaJSON_ProducesStableOutput(t *testing.T) {
	// Two inputs with different whitespace and key order should normalize
	// to the same canonical JSON.
	inputA := `{"type":"OBJECT","structKind":{"properties":{"email":{"type":"STRING","semanticType":"abc"}}}}`
	inputB := `{
  "structKind": {
    "properties": {
      "email": { "semanticType": "abc", "type": "STRING" }
    }
  },
  "type": "OBJECT"
}`

	gotA, err := NormalizeObjectSchemaJSON(inputA)
	if err != nil {
		t.Fatalf("normalize A: %v", err)
	}
	gotB, err := NormalizeObjectSchemaJSON(inputB)
	if err != nil {
		t.Fatalf("normalize B: %v", err)
	}
	if gotA != gotB {
		t.Errorf("canonical forms differ:\nA=%s\nB=%s", gotA, gotB)
	}
}

func TestNormalizeObjectSchemaJSON_RejectsInvalidProto(t *testing.T) {
	_, err := NormalizeObjectSchemaJSON(`{"type":"NOT_A_REAL_TYPE"}`)
	if err == nil {
		t.Fatal("expected error for invalid enum value, got nil")
	}
}

func TestNormalizeObjectSchemaJSON_RejectsInvalidJSON(t *testing.T) {
	_, err := NormalizeObjectSchemaJSON(`not json at all`)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestNormalizeObjectSchemaJSON_EmptyString(t *testing.T) {
	got, err := NormalizeObjectSchemaJSON("")
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output for empty input, got %q", got)
	}
}

func TestParseObjectSchemaJSON_RoundTripsThroughProto(t *testing.T) {
	input := `{"type":"OBJECT","structKind":{"properties":{"x":{"type":"STRING","semanticType":"s"}}}}`
	schema, err := ParseObjectSchemaJSON(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if schema.GetType() != v1pb.ObjectSchema_OBJECT {
		t.Errorf("expected OBJECT, got %v", schema.GetType())
	}
	props := schema.GetStructKind().GetProperties()
	if props["x"].GetSemanticType() != "s" {
		t.Errorf("expected semanticType s, got %q", props["x"].GetSemanticType())
	}
}

func TestMarshalObjectSchemaToJSON_DeterministicOrder(t *testing.T) {
	// Build a proto with multiple map keys and verify marshal is
	// byte-identical across calls AND that keys are sorted.
	schema := &v1pb.ObjectSchema{
		Type: v1pb.ObjectSchema_OBJECT,
		Kind: &v1pb.ObjectSchema_StructKind_{
			StructKind: &v1pb.ObjectSchema_StructKind{
				Properties: map[string]*v1pb.ObjectSchema{
					"zeta":  {Type: v1pb.ObjectSchema_STRING, SemanticType: "z"},
					"alpha": {Type: v1pb.ObjectSchema_STRING, SemanticType: "a"},
				},
			},
		},
	}
	a, err := MarshalObjectSchemaToJSON(schema)
	if err != nil {
		t.Fatalf("marshal a: %v", err)
	}
	b, err := MarshalObjectSchemaToJSON(schema)
	if err != nil {
		t.Fatalf("marshal b: %v", err)
	}
	if a != b {
		t.Errorf("marshal not deterministic: %q vs %q", a, b)
	}
	// Alpha must appear before zeta since we sort map keys.
	if strings.Index(a, "alpha") > strings.Index(a, "zeta") {
		t.Errorf("expected sorted map keys in output, got %s", a)
	}
}
