package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/durationpb"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

func TestWorkspaceProfileSettingSchemaSeparatesRequestAndRoleExpiration(t *testing.T) {
	workspaceProfile := getWorkspaceProfileSetting(false).Elem.(*schema.Resource)

	requestExpiration, ok := workspaceProfile.Schema["maximum_request_expiration_in_seconds"]
	if !ok {
		t.Fatal("maximum_request_expiration_in_seconds schema is missing")
	}
	if !strings.Contains(requestExpiration.Description, "data access requests") {
		t.Fatalf("maximum_request_expiration_in_seconds description = %q, want data access scope", requestExpiration.Description)
	}
	if strings.Contains(requestExpiration.Description, "role grants") {
		t.Fatalf("maximum_request_expiration_in_seconds description = %q, should not mention role grants", requestExpiration.Description)
	}

	roleExpiration, ok := workspaceProfile.Schema["maximum_role_expiration_in_seconds"]
	if !ok {
		t.Fatal("maximum_role_expiration_in_seconds schema is missing")
	}
	if !strings.Contains(roleExpiration.Description, "request role") {
		t.Fatalf("maximum_role_expiration_in_seconds description = %q, want request role scope", roleExpiration.Description)
	}
}

func TestFlattenWorkspaceProfileSettingSeparatesRequestAndRoleExpiration(t *testing.T) {
	got := flattenWorkspaceProfileSetting(&v1pb.WorkspaceProfileSetting{
		MaximumRequestExpiration: &durationpb.Duration{Seconds: 3600},
		MaximumRoleExpiration:    &durationpb.Duration{Seconds: 7200},
	})

	profile := got[0].(map[string]interface{})
	if profile["maximum_request_expiration_in_seconds"] != 3600 {
		t.Fatalf("maximum_request_expiration_in_seconds = %#v, want 3600", profile["maximum_request_expiration_in_seconds"])
	}
	if profile["maximum_role_expiration_in_seconds"] != 7200 {
		t.Fatalf("maximum_role_expiration_in_seconds = %#v, want 7200", profile["maximum_role_expiration_in_seconds"])
	}
}
