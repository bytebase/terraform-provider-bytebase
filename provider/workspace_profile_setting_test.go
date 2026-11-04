package provider

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestWorkspaceProfileSettingInactiveSessionTimeout(t *testing.T) {
	for _, computed := range []bool{false, true} {
		field := getWorkspaceProfileSetting(computed).Elem.(*schema.Resource).Schema["inactive_session_timeout_in_seconds"]
		if field == nil {
			t.Fatal("inactive_session_timeout_in_seconds schema is missing")
		}
		if field.Type != schema.TypeInt || !field.Optional || !field.Computed {
			t.Fatalf("timeout schema must use optional computed seconds: %+v", field)
		}
	}
	for _, seconds := range []int64{0, 1800} {
		got := flattenWorkspaceProfileSetting(&v1pb.WorkspaceProfileSetting{InactiveSessionTimeout: &durationpb.Duration{Seconds: seconds}})
		if value := got[0].(map[string]interface{})["inactive_session_timeout_in_seconds"]; value != int(seconds) {
			t.Fatalf("flatten timeout = %v, want %d", value, seconds)
		}
	}
	if got := flattenWorkspaceProfileSetting(&v1pb.WorkspaceProfileSetting{})[0].(map[string]interface{})["inactive_session_timeout_in_seconds"]; got != 0 {
		t.Fatalf("unset server timeout = %v, want disabled zero", got)
	}
}

func TestWorkspaceProfileSettingInactiveSessionTimeoutUpdateMasks(t *testing.T) {
	for _, tc := range []struct {
		name    string
		profile map[string]interface{}
		seconds int64
		masks   []string
	}{
		{name: "omitted", profile: map[string]interface{}{"query_timeout_in_seconds": 60}, masks: []string{"value.workspace_profile.query_timeout"}},
		{name: "disabled", profile: map[string]interface{}{"inactive_session_timeout_in_seconds": 0}, masks: []string{"value.workspace_profile.inactive_session_timeout"}},
		{name: "configured", profile: map[string]interface{}{"inactive_session_timeout_in_seconds": 1800}, seconds: 1800, masks: []string{"value.workspace_profile.inactive_session_timeout"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &settingTestClient{response: &v1pb.Setting{
				Name:  "settings/WORKSPACE_PROFILE",
				Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_WorkspaceProfile{WorkspaceProfile: &v1pb.WorkspaceProfileSetting{}}},
			}}
			applySettingTestConfig(t, map[string]interface{}{
				"name": "settings/WORKSPACE_PROFILE", "workspace_profile": []interface{}{tc.profile},
			}, nil, client)
			if !reflect.DeepEqual(client.masks, tc.masks) {
				t.Fatalf("masks = %v, want %v", client.masks, tc.masks)
			}
			timeout := client.patch.GetValue().GetWorkspaceProfile().GetInactiveSessionTimeout()
			if tc.name == "omitted" {
				if timeout != nil {
					t.Fatalf("omitted timeout sent as %v", timeout)
				}
			} else if timeout == nil || timeout.Seconds != tc.seconds {
				t.Fatalf("timeout = %v, want %d seconds", timeout, tc.seconds)
			}
		})
	}
}

func TestWorkspaceProfileSettingDataSourceReadsInactiveSessionTimeout(t *testing.T) {
	d := schema.TestResourceDataRaw(t, dataSourceSetting().Schema, map[string]interface{}{"name": "settings/WORKSPACE_PROFILE"})
	client := &settingTestClient{response: &v1pb.Setting{
		Name: "settings/WORKSPACE_PROFILE",
		Value: &v1pb.SettingValue{Value: &v1pb.SettingValue_WorkspaceProfile{WorkspaceProfile: &v1pb.WorkspaceProfileSetting{
			InactiveSessionTimeout: &durationpb.Duration{Seconds: 1800},
		}}},
	}}
	if diags := dataSourceSettingRead(context.Background(), d, client); diags.HasError() {
		t.Fatalf("data source read: %v", diags)
	}
	if got := d.Get("workspace_profile.0.inactive_session_timeout_in_seconds"); got != 1800 {
		t.Fatalf("read timeout = %v, want 1800", got)
	}
}

func TestWorkspaceProfileSettingDeleteResetsInactiveSessionTimeout(t *testing.T) {
	d := resourceSetting().Data(&terraform.InstanceState{ID: "settings/WORKSPACE_PROFILE"})
	client := &settingTestClient{response: &v1pb.Setting{Name: "settings/WORKSPACE_PROFILE"}}
	if diags := resourceSettingDelete(context.Background(), d, client); diags.HasError() {
		t.Fatalf("delete workspace profile: %v", diags)
	}
	for _, mask := range client.masks {
		if mask == "value.workspace_profile.inactive_session_timeout" {
			if timeout := client.patch.GetValue().GetWorkspaceProfile().GetInactiveSessionTimeout(); timeout != nil {
				t.Fatalf("reset timeout = %v, want unset", timeout)
			}
			return
		}
	}
	t.Fatal("delete workspace profile did not reset inactivity timeout")
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
