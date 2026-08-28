package provider

import (
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

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

func TestDataSourceGCPResourceIDSupport(t *testing.T) {
	resourceDataSourceSchema := resourceInstance().Schema["data_sources"].Elem.(*schema.Resource).Schema
	assertStringSchema(t, "bytebase_instance.data_sources.project_id", resourceDataSourceSchema["project_id"], true, false)
	assertStringSchema(t, "bytebase_instance.data_sources.instance_id", resourceDataSourceSchema["instance_id"], true, false)

	computedDataSourceSchema := getDataSourceComputedSchema()
	assertStringSchema(t, "data.bytebase_instance.data_sources.project_id", computedDataSourceSchema["project_id"], false, true)
	assertStringSchema(t, "data.bytebase_instance.data_sources.instance_id", computedDataSourceSchema["instance_id"], false, true)

	dataSource, err := convertToV1DataSource(map[string]interface{}{
		"id":          "admin",
		"type":        v1pb.DataSourceType_ADMIN.String(),
		"project_id":  "gcp-project",
		"instance_id": "spanner-instance",
	})
	if err != nil {
		t.Fatalf("convertToV1DataSource returned error: %v", err)
	}
	if dataSource.ProjectId != "gcp-project" {
		t.Fatalf("ProjectId = %q, want %q", dataSource.ProjectId, "gcp-project")
	}
	if dataSource.InstanceId != "spanner-instance" {
		t.Fatalf("InstanceId = %q, want %q", dataSource.InstanceId, "spanner-instance")
	}

	flattened, err := flattenDataSourceList(nil, []*v1pb.DataSource{dataSource}, v1pb.Engine_SPANNER)
	if err != nil {
		t.Fatalf("flattenDataSourceList returned error: %v", err)
	}
	raw := flattened[0].(map[string]interface{})
	if raw["project_id"] != "gcp-project" {
		t.Fatalf("flattened project_id = %q, want %q", raw["project_id"], "gcp-project")
	}
	if raw["instance_id"] != "spanner-instance" {
		t.Fatalf("flattened instance_id = %q, want %q", raw["instance_id"], "spanner-instance")
	}
}

func TestDataSourceClusterRemovedFromSchema(t *testing.T) {
	resourceDataSourceSchema := resourceInstance().Schema["data_sources"].Elem.(*schema.Resource).Schema
	if _, ok := resourceDataSourceSchema["cluster"]; ok {
		t.Fatal("bytebase_instance.data_sources.cluster is present after the protocol field was removed")
	}

	if _, ok := getDataSourceComputedSchema()["cluster"]; ok {
		t.Fatal("data.bytebase_instance.data_sources.cluster is present after the protocol field was removed")
	}
}

func TestProjectInstanceParentSchema(t *testing.T) {
	resourceParent, ok := resourceInstance().Schema["parent"]
	if !ok {
		t.Fatal("bytebase_instance.parent is missing")
	}
	assertStringSchema(t, "bytebase_instance.parent", resourceParent, true, true)
	if !resourceParent.ForceNew {
		t.Fatal("bytebase_instance.parent must force replacement")
	}

	dataSourceParent, ok := dataSourceInstance().Schema["parent"]
	if !ok {
		t.Fatal("data.bytebase_instance.parent is missing")
	}
	assertStringSchema(t, "data.bytebase_instance.parent", dataSourceParent, true, false)

	listParent, ok := dataSourceInstanceList().Schema["parent"]
	if !ok {
		t.Fatal("data.bytebase_instance_list.parent is missing")
	}
	assertStringSchema(t, "data.bytebase_instance_list.parent", listParent, true, false)

	listedInstanceParent, ok := dataSourceInstanceList().Schema["instances"].Elem.(*schema.Resource).Schema["parent"]
	if !ok {
		t.Fatal("data.bytebase_instance_list.instances.parent is missing")
	}
	assertStringSchema(t, "data.bytebase_instance_list.instances.parent", listedInstanceParent, false, true)
}

func TestVaultExternalSecretTLSSchemaAndConversion(t *testing.T) {
	resourceVaultSchema := resourceInstance().Schema["data_sources"].Elem.(*schema.Resource).Schema["external_secret"].Elem.(*schema.Resource).Schema["vault"].Elem.(*schema.Resource).Schema
	assertSensitiveOptionalComputedSchema(t, "bytebase_instance.data_sources.external_secret.vault.vault_ssl_ca", resourceVaultSchema["vault_ssl_ca"])
	assertSensitiveOptionalComputedSchema(t, "bytebase_instance.data_sources.external_secret.vault.vault_ssl_cert", resourceVaultSchema["vault_ssl_cert"])
	assertSensitiveOptionalComputedSchema(t, "bytebase_instance.data_sources.external_secret.vault.vault_ssl_key", resourceVaultSchema["vault_ssl_key"])
	assertBoolSchema(t, "bytebase_instance.data_sources.external_secret.vault.skip_vault_tls_verification", resourceVaultSchema["skip_vault_tls_verification"], true, true)

	computedVaultSchema := getExternalSecretSchema().Elem.(*schema.Resource).Schema["vault"].Elem.(*schema.Resource).Schema
	assertSensitiveComputedSchema(t, "data.bytebase_instance.data_sources.external_secret.vault.vault_ssl_ca", computedVaultSchema["vault_ssl_ca"])
	assertSensitiveComputedSchema(t, "data.bytebase_instance.data_sources.external_secret.vault.vault_ssl_cert", computedVaultSchema["vault_ssl_cert"])
	assertSensitiveComputedSchema(t, "data.bytebase_instance.data_sources.external_secret.vault.vault_ssl_key", computedVaultSchema["vault_ssl_key"])
	assertBoolSchema(t, "data.bytebase_instance.data_sources.external_secret.vault.skip_vault_tls_verification", computedVaultSchema["skip_vault_tls_verification"], false, true)

	dataSource, err := convertToV1DataSource(map[string]interface{}{
		"id":                  "admin",
		"type":                v1pb.DataSourceType_ADMIN.String(),
		"authentication_type": v1pb.DataSource_PASSWORD.String(),
		"external_secret": []interface{}{
			map[string]interface{}{
				"vault": []interface{}{
					map[string]interface{}{
						"url":                         "https://vault.example.com:8200",
						"token":                       "vault-token",
						"token_type":                  v1pb.DataSourceExternalSecret_PLAIN.String(),
						"engine_name":                 "secret",
						"secret_name":                 "database/postgres",
						"password_key_name":           "password",
						"vault_ssl_ca":                "ca-pem",
						"vault_ssl_cert":              "cert-pem",
						"vault_ssl_key":               "key-pem",
						"skip_vault_tls_verification": true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("convertToV1DataSource returned error: %v", err)
	}
	externalSecret := dataSource.GetExternalSecret()
	if externalSecret.GetVaultSslCa() != "ca-pem" {
		t.Fatalf("VaultSslCa = %q, want %q", externalSecret.GetVaultSslCa(), "ca-pem")
	}
	if externalSecret.GetVaultSslCert() != "cert-pem" {
		t.Fatalf("VaultSslCert = %q, want %q", externalSecret.GetVaultSslCert(), "cert-pem")
	}
	if externalSecret.GetVaultSslKey() != "key-pem" {
		t.Fatalf("VaultSslKey = %q, want %q", externalSecret.GetVaultSslKey(), "key-pem")
	}
	if !externalSecret.GetSkipVaultTlsVerification() {
		t.Fatal("SkipVaultTlsVerification = false, want true")
	}
}

func TestFlattenVaultExternalSecretPreservesInputOnlyTLS(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceInstance().Schema, map[string]interface{}{
		"resource_id": "vault-instance",
		"title":       "Vault instance",
		"engine":      v1pb.Engine_POSTGRES.String(),
		"data_sources": []interface{}{
			map[string]interface{}{
				"id":                  "admin",
				"type":                v1pb.DataSourceType_ADMIN.String(),
				"authentication_type": v1pb.DataSource_PASSWORD.String(),
				"external_secret": []interface{}{
					map[string]interface{}{
						"vault": []interface{}{
							map[string]interface{}{
								"url":                         "https://vault.example.com:8200",
								"token":                       "vault-token",
								"token_type":                  v1pb.DataSourceExternalSecret_PLAIN.String(),
								"engine_name":                 "secret",
								"secret_name":                 "database/postgres",
								"password_key_name":           "password",
								"vault_ssl_ca":                "old-ca-pem",
								"vault_ssl_cert":              "old-cert-pem",
								"vault_ssl_key":               "old-key-pem",
								"skip_vault_tls_verification": true,
							},
						},
					},
				},
			},
		},
	})

	flattened, err := flattenDataSourceList(d, []*v1pb.DataSource{
		{
			Id:   "admin",
			Type: v1pb.DataSourceType_ADMIN,
			ExternalSecret: &v1pb.DataSourceExternalSecret{
				SecretType:               v1pb.DataSourceExternalSecret_VAULT_KV_V2,
				Url:                      "https://vault.example.com:8200",
				AuthType:                 v1pb.DataSourceExternalSecret_TOKEN,
				AuthOption:               &v1pb.DataSourceExternalSecret_Token{Token: ""},
				TokenType:                v1pb.DataSourceExternalSecret_PLAIN,
				EngineName:               "secret",
				SecretName:               "database/postgres",
				PasswordKeyName:          "password",
				SkipVaultTlsVerification: true,
			},
		},
	}, v1pb.Engine_POSTGRES)
	if err != nil {
		t.Fatalf("flattenDataSourceList returned error: %v", err)
	}
	raw := flattened[0].(map[string]interface{})
	externalSecret := raw["external_secret"].([]any)[0].(map[string]interface{})
	vault := externalSecret["vault"].([]any)[0].(map[string]interface{})
	if vault["vault_ssl_ca"] != "old-ca-pem" {
		t.Fatalf("vault_ssl_ca = %q, want %q", vault["vault_ssl_ca"], "old-ca-pem")
	}
	if vault["vault_ssl_cert"] != "old-cert-pem" {
		t.Fatalf("vault_ssl_cert = %q, want %q", vault["vault_ssl_cert"], "old-cert-pem")
	}
	if vault["vault_ssl_key"] != "old-key-pem" {
		t.Fatalf("vault_ssl_key = %q, want %q", vault["vault_ssl_key"], "old-key-pem")
	}
	if !vault["skip_vault_tls_verification"].(bool) {
		t.Fatalf("skip_vault_tls_verification = %v, want true", vault["skip_vault_tls_verification"])
	}
}

func TestSyncDatabasesUsesProtocolWrapper(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceInstance().Schema, map[string]interface{}{
		"sync_databases": []interface{}{"db1", "db2"},
	})

	got := getSyncDatabases(d)
	if got == nil {
		t.Fatal("getSyncDatabases returned nil, want wrapper")
	}
	gotDatabases := map[string]bool{}
	for _, database := range got.Databases {
		gotDatabases[database] = true
	}
	if len(gotDatabases) != 2 || !gotDatabases["db1"] || !gotDatabases["db2"] {
		t.Fatalf("Databases = %#v, want [db1 db2]", got.Databases)
	}
}

func TestEmptySyncDatabasesMeansAllDatabases(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceInstance().Schema, map[string]interface{}{})

	if got := getSyncDatabases(d); got != nil {
		t.Fatalf("getSyncDatabases returned %#v, want nil", got)
	}
}

func TestFlattenSyncDatabasesHandlesNilWrapper(t *testing.T) {
	if got := flattenSyncDatabases(nil); got != nil {
		t.Fatalf("flattenSyncDatabases(nil) = %#v, want nil", got)
	}

	got := flattenSyncDatabases(&v1pb.SyncDatabases{Databases: []string{"db1", "db2"}})
	if len(got) != 2 || got[0] != "db1" || got[1] != "db2" {
		t.Fatalf("flattenSyncDatabases(wrapper) = %#v, want [db1 db2]", got)
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

func TestEngineValidationDoesNotAllowSQLite(t *testing.T) {
	_, errors := internal.EngineValidation("SQLITE", "engine")
	if len(errors) == 0 {
		t.Fatal("EngineValidation accepted SQLITE, want validation error")
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

func assertBoolSchema(t *testing.T, name string, got *schema.Schema, wantOptional, wantComputed bool) {
	t.Helper()
	assertSchemaPresence(t, name, got, wantComputed)
	if got.Type != schema.TypeBool {
		t.Fatalf("%s type = %v, want %v", name, got.Type, schema.TypeBool)
	}
	if got.Optional != wantOptional {
		t.Fatalf("%s Optional = %v, want %v", name, got.Optional, wantOptional)
	}
}

func assertSensitiveOptionalComputedSchema(t *testing.T, name string, got *schema.Schema) {
	t.Helper()
	assertStringSchema(t, name, got, true, true)
	if !got.Sensitive {
		t.Fatalf("%s Sensitive = false, want true", name)
	}
	if got.DiffSuppressFunc == nil {
		t.Fatalf("%s DiffSuppressFunc is nil, want suppressSensitiveFieldDiff", name)
	}
}

func assertSensitiveComputedSchema(t *testing.T, name string, got *schema.Schema) {
	t.Helper()
	assertStringSchema(t, name, got, false, true)
	if !got.Sensitive {
		t.Fatalf("%s Sensitive = false, want true", name)
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
