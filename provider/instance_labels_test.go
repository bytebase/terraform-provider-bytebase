package provider

import (
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
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

func TestDataSourceCloudSQLIPTypeSupport(t *testing.T) {
	assertStringSchema(t, "bytebase_instance.data_sources.cloud_sql_ip_type", resourceInstance().Schema["data_sources"].Elem.(*schema.Resource).Schema["cloud_sql_ip_type"], true, false)
	assertStringSchema(t, "data.bytebase_instance.data_sources.cloud_sql_ip_type", getDataSourceComputedSchema()["cloud_sql_ip_type"], false, true)

	dataSource, err := convertToV1DataSource(map[string]interface{}{
		"id":                  "admin",
		"type":                v1pb.DataSourceType_ADMIN.String(),
		"authentication_type": v1pb.DataSource_GOOGLE_CLOUD_SQL_IAM.String(),
		"cloud_sql_ip_type":   v1pb.DataSource_PRIVATE.String(),
	})
	if err != nil {
		t.Fatalf("convertToV1DataSource returned error: %v", err)
	}
	if dataSource.CloudSqlIpType != v1pb.DataSource_PRIVATE {
		t.Fatalf("CloudSqlIpType = %v, want %v", dataSource.CloudSqlIpType, v1pb.DataSource_PRIVATE)
	}
}

func TestProjectReadParitySchema(t *testing.T) {
	for _, field := range []string{
		"execution_retry_policy",
		"ci_sampling_size",
		"parallel_tasks_per_rollout",
	} {
		assertComputedSchema(t, "data.bytebase_project."+field, dataSourceProject().Schema[field])

		projectListSchema := dataSourceProjectList().Schema["projects"].Elem.(*schema.Resource).Schema
		assertComputedSchema(t, "data.bytebase_project_list.projects."+field, projectListSchema[field])
	}
}

func TestInstanceStatusSchema(t *testing.T) {
	for _, resource := range []struct {
		name   string
		schema map[string]*schema.Schema
	}{
		{"bytebase_instance", resourceInstance().Schema},
		{"data.bytebase_instance", dataSourceInstance().Schema},
		{"data.bytebase_instance_list.instances", dataSourceInstanceList().Schema["instances"].Elem.(*schema.Resource).Schema},
	} {
		assertStringSchema(t, resource.name+".state", resource.schema["state"], false, true)
		assertStringSchema(t, resource.name+".last_sync_time", resource.schema["last_sync_time"], false, true)
		assertComputedSchema(t, resource.name+".roles", resource.schema["roles"])
	}
}

func TestDatabaseStatusSchema(t *testing.T) {
	for _, resource := range []struct {
		name   string
		schema map[string]*schema.Schema
	}{
		{"bytebase_database", resourceDatabase().Schema},
		{"data.bytebase_database", dataSourceDatabase().Schema},
		{"data.bytebase_database_list.databases", dataSourceDatabaseList().Schema["databases"].Elem.(*schema.Resource).Schema},
	} {
		assertStringSchema(t, resource.name+".release", resource.schema["release"], false, true)
		assertStringSchema(t, resource.name+".effective_environment", resource.schema["effective_environment"], false, true)
		assertComputedSchema(t, resource.name+".instance_resource", resource.schema["instance_resource"])
		assertComputedSchema(t, resource.name+".backup_available", resource.schema["backup_available"])
		assertStringSchema(t, resource.name+".sync_status", resource.schema["sync_status"], false, true)
		assertStringSchema(t, resource.name+".sync_error", resource.schema["sync_error"], false, true)
	}
}

func assertStringSchema(t *testing.T, name string, got *schema.Schema, wantOptional, wantComputed bool) {
	t.Helper()
	assertSchemaPresence(t, name, got, wantComputed)
	if got.Type != schema.TypeString {
		t.Fatalf("%s type = %v, want %v", name, got.Type, schema.TypeString)
	}
	if got.Optional != wantOptional {
		t.Fatalf("%s Optional = %v, want %v", name, got.Optional, wantOptional)
	}
}

func assertComputedSchema(t *testing.T, name string, got *schema.Schema) {
	t.Helper()
	assertSchemaPresence(t, name, got, true)
}

func assertSchemaPresence(t *testing.T, name string, got *schema.Schema, wantComputed bool) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s schema is missing", name)
	}
	if got.Computed != wantComputed {
		t.Fatalf("%s Computed = %v, want %v", name, got.Computed, wantComputed)
	}
}
