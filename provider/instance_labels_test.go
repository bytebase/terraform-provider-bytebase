package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestInstanceLabelsSchema(t *testing.T) {
	assertInstanceLabelsSchema(t, "bytebase_instance", resourceInstance().Schema["labels"], true, true)
	assertInstanceLabelsSchema(t, "data.bytebase_instance", dataSourceInstance().Schema["labels"], false, true)

	instanceListSchema := dataSourceInstanceList().Schema["instances"].Elem.(*schema.Resource).Schema
	assertInstanceLabelsSchema(t, "data.bytebase_instance_list.instances", instanceListSchema["labels"], false, true)
}

func assertInstanceLabelsSchema(t *testing.T, name string, labels *schema.Schema, wantOptional, wantComputed bool) {
	t.Helper()

	if labels == nil {
		t.Fatalf("%s labels schema is missing", name)
	}
	if labels.Type != schema.TypeMap {
		t.Fatalf("%s labels type = %v, want %v", name, labels.Type, schema.TypeMap)
	}
	if labels.Optional != wantOptional {
		t.Fatalf("%s labels Optional = %v, want %v", name, labels.Optional, wantOptional)
	}
	if labels.Computed != wantComputed {
		t.Fatalf("%s labels Computed = %v, want %v", name, labels.Computed, wantComputed)
	}
	elem, ok := labels.Elem.(*schema.Schema)
	if !ok {
		t.Fatalf("%s labels Elem = %T, want *schema.Schema", name, labels.Elem)
	}
	if elem.Type != schema.TypeString {
		t.Fatalf("%s labels Elem type = %v, want %v", name, elem.Type, schema.TypeString)
	}
}
