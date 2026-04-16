package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func TestConvertToV1TableCatalog_ObjectSchemaJSON(t *testing.T) {
	raw := map[string]any{
		"name":               "users_index",
		"classification":     "",
		"columns":            schema.NewSet(columnHash, nil),
		"object_schema_json": `{"type":"OBJECT","structKind":{"properties":{"email":{"type":"STRING","semanticType":"abc"}}}}`,
	}
	got, err := convertToV1TableCatalog(raw)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if got.GetObjectSchema() == nil {
		t.Fatal("expected ObjectSchema variant, got nil")
	}
	if got.GetObjectSchema().GetType() != v1pb.ObjectSchema_OBJECT {
		t.Errorf("wrong type: %v", got.GetObjectSchema().GetType())
	}
	if got.GetColumns() != nil {
		t.Errorf("expected Columns variant unset, got %+v", got.GetColumns())
	}
}

func TestConvertToV1TableCatalog_ColumnsOnly_StillWorks(t *testing.T) {
	col := map[string]any{
		"name":           "id",
		"semantic_type":  "",
		"classification": "",
		"labels":         map[string]any{},
	}
	raw := map[string]any{
		"name":               "users",
		"classification":     "",
		"columns":            schema.NewSet(columnHash, []any{col}),
		"object_schema_json": "",
	}
	got, err := convertToV1TableCatalog(raw)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if got.GetColumns() == nil {
		t.Fatal("expected Columns variant")
	}
	if len(got.GetColumns().Columns) != 1 {
		t.Errorf("expected 1 column, got %d", len(got.GetColumns().Columns))
	}
	if got.GetObjectSchema() != nil {
		t.Errorf("expected ObjectSchema variant unset, got %+v", got.GetObjectSchema())
	}
}

func TestConvertToV1TableCatalog_BothSet_IsError(t *testing.T) {
	col := map[string]any{
		"name": "id", "semantic_type": "", "classification": "",
		"labels": map[string]any{},
	}
	raw := map[string]any{
		"name":               "mixed",
		"classification":     "",
		"columns":            schema.NewSet(columnHash, []any{col}),
		"object_schema_json": `{"type":"OBJECT"}`,
	}
	_, err := convertToV1TableCatalog(raw)
	if err == nil {
		t.Fatal("expected error for mutually-exclusive columns + object_schema_json")
	}
}

func TestConvertToV1TableCatalog_NeitherSet_EmitsEmptyColumns(t *testing.T) {
	// After Task 5 relaxed columns to Optional, an HCL table with only
	// a name is valid. The convert function should still produce a
	// valid TableCatalog — pick the Columns variant with zero columns.
	raw := map[string]any{
		"name":               "empty",
		"classification":     "",
		"columns":            schema.NewSet(columnHash, nil),
		"object_schema_json": "",
	}
	got, err := convertToV1TableCatalog(raw)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if got.GetColumns() == nil {
		t.Fatal("expected Columns variant (possibly empty)")
	}
	if len(got.GetColumns().Columns) != 0 {
		t.Errorf("expected zero columns, got %d", len(got.GetColumns().Columns))
	}
}
