package provider

import (
	"strings"
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

func TestFlattenDatabaseCatalog_ObjectSchema(t *testing.T) {
	catalog := &v1pb.DatabaseCatalog{
		Name: "instances/x/databases/y/catalog",
		Schemas: []*v1pb.SchemaCatalog{{
			Name: "",
			Tables: []*v1pb.TableCatalog{{
				Name: "users_index",
				Kind: &v1pb.TableCatalog_ObjectSchema{
					ObjectSchema: &v1pb.ObjectSchema{
						Type: v1pb.ObjectSchema_OBJECT,
						Kind: &v1pb.ObjectSchema_StructKind_{
							StructKind: &v1pb.ObjectSchema_StructKind{
								Properties: map[string]*v1pb.ObjectSchema{
									"email": {Type: v1pb.ObjectSchema_STRING, SemanticType: "abc"},
								},
							},
						},
					},
				},
			}},
		}},
	}
	out := flattenDatabaseCatalog(catalog)
	if len(out) != 1 {
		t.Fatalf("expected 1 catalog, got %d", len(out))
	}
	schemas := out[0].(map[string]any)["schemas"].(*schema.Set).List()
	if len(schemas) == 0 {
		t.Fatal("no schemas in flattened catalog")
	}
	tables := schemas[0].(map[string]any)["tables"].(*schema.Set).List()
	if len(tables) == 0 {
		t.Fatal("no tables in flattened catalog")
	}
	rawTable := tables[0].(map[string]any)
	got, _ := rawTable["object_schema_json"].(string)
	if got == "" {
		t.Fatal("object_schema_json not populated in flattened state")
	}
	// Ensure the canonical form contains the expected semantic type.
	if !strings.Contains(got, `"abc"`) {
		t.Errorf("canonical JSON missing semantic type abc: %s", got)
	}
}

func TestFlattenDatabaseCatalog_ColumnsPathUnchanged(t *testing.T) {
	catalog := &v1pb.DatabaseCatalog{
		Name: "instances/x/databases/y/catalog",
		Schemas: []*v1pb.SchemaCatalog{{
			Name: "",
			Tables: []*v1pb.TableCatalog{{
				Name: "users",
				Kind: &v1pb.TableCatalog_Columns_{
					Columns: &v1pb.TableCatalog_Columns{
						Columns: []*v1pb.ColumnCatalog{{
							Name:         "email",
							SemanticType: "email-mask",
						}},
					},
				},
			}},
		}},
	}
	out := flattenDatabaseCatalog(catalog)
	schemas := out[0].(map[string]any)["schemas"].(*schema.Set).List()
	tables := schemas[0].(map[string]any)["tables"].(*schema.Set).List()
	rawTable := tables[0].(map[string]any)

	// object_schema_json should be empty for a Columns-variant table.
	if got, _ := rawTable["object_schema_json"].(string); got != "" {
		t.Errorf("expected empty object_schema_json for columns table, got %q", got)
	}
	// Columns must still be populated.
	cols := rawTable["columns"].(*schema.Set).List()
	if len(cols) != 1 {
		t.Fatalf("expected 1 column, got %d", len(cols))
	}
}
