package provider

import (
	"context"
	"errors"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"google.golang.org/genproto/googleapis/type/expr"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestIAMPolicyFrameworkSchema(t *testing.T) {
	ctx := context.Background()
	r := newIAMPolicyFrameworkResource()
	var response resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &response)
	if response.Schema.Version != 0 {
		t.Fatal("legacy state must remain schema version zero")
	}
	bindings, ok := response.Schema.Attributes["bindings"].(schema.MapNestedAttribute)
	if !ok || !bindings.Optional || !bindings.Computed {
		t.Fatal("bindings must be an optional/computed nested map")
	}
	if _, ok := bindings.NestedObject.Attributes["condition"].(schema.SingleNestedAttribute); !ok {
		t.Fatal("keyed conditions must be single objects")
	}
	legacy, ok := response.Schema.Blocks["iam_policy"].(schema.ListNestedBlock)
	if !ok {
		t.Fatal("legacy policy must remain a list block")
	}
	binding, ok := legacy.NestedObject.Blocks["binding"].(schema.SetNestedBlock)
	if !ok {
		t.Fatal("legacy bindings must remain set blocks")
	}
	if _, ok := binding.NestedObject.Blocks["condition"].(schema.SetNestedBlock); !ok {
		t.Fatal("legacy conditions must remain set blocks")
	}
}

type frameworkIAMClient struct {
	api.Client
	policy          *v1pb.IamPolicy
	writes          []*v1pb.SetIamPolicyRequest
	project         string
	writeError      error
	permissions     []string
	projectPolicies map[string]*v1pb.IamPolicy
}

func (*frameworkIAMClient) GetWorkspaceName() string { return "workspaces/test" }
func (c *frameworkIAMClient) GetRole(context.Context, string) (*v1pb.Role, error) {
	return &v1pb.Role{Permissions: c.permissions}, nil
}
func (c *frameworkIAMClient) GetWorkspaceIAMPolicy(context.Context) (*v1pb.IamPolicy, error) {
	return c.policy, nil
}
func (c *frameworkIAMClient) GetProjectIAMPolicy(_ context.Context, parent string) (*v1pb.IamPolicy, error) {
	c.project = parent
	if c.projectPolicies != nil {
		return c.projectPolicies[parent], nil
	}
	return c.policy, nil
}
func (c *frameworkIAMClient) SetWorkspaceIAMPolicy(_ context.Context, request *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	c.writes = append(c.writes, proto.Clone(request).(*v1pb.SetIamPolicyRequest))
	if c.writeError != nil {
		return nil, c.writeError
	}
	c.policy = proto.Clone(request.Policy).(*v1pb.IamPolicy)
	c.policy.Etag = "etag-next"
	return c.policy, nil
}
func (c *frameworkIAMClient) SetProjectIAMPolicy(ctx context.Context, parent string, request *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	c.project = parent
	return c.SetWorkspaceIAMPolicy(ctx, request)
}

func frameworkIAMSchema(t *testing.T) schema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	newIAMPolicyFrameworkResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	return response.Schema
}

func frameworkIAMState(t *testing.T, model iamPolicyFrameworkModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: frameworkIAMSchema(t)}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatal(diags)
	}
	return state
}

func frameworkIAMMap(t *testing.T, role string, members ...string) types.Map {
	t.Helper()
	memberValues := make([]attr.Value, len(members))
	for i, member := range members {
		memberValues[i] = types.StringValue(member)
	}
	return types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{
		"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, map[string]attr.Value{
			"role": types.StringValue(role), "members": types.SetValueMust(types.StringType, memberValues),
			"condition": types.ObjectNull(iamFrameworkConditionType().AttrTypes),
		}),
	})
}

func frameworkIAMModel(t *testing.T, parent string) iamPolicyFrameworkModel {
	t.Helper()
	return iamPolicyFrameworkModel{
		ID: types.StringValue(parent), Parent: types.StringValue(parent), Etag: types.StringValue("etag-old"),
		Bindings:  frameworkIAMMap(t, "roles/reader", "user:a@example.com"),
		IAMPolicy: types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{}),
	}
}

func TestIAMPolicyFrameworkUpdateWritesEtagAndFullPolicy(t *testing.T) {
	ctx := context.Background()
	client := &frameworkIAMClient{}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: client}, &resource.ConfigureResponse{})
	old := frameworkIAMModel(t, "projects/example")
	plan := old
	plan.Bindings = frameworkIAMMap(t, "roles/reader", "user:a@example.com", "user:b@example.com")
	plan.Etag = types.StringUnknown()
	state := frameworkIAMState(t, old)
	planned := frameworkIAMState(t, plan)
	response := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	if len(client.writes) != 1 || client.writes[0].Policy.Etag != "etag-old" {
		t.Fatalf("expected one etag-protected write, got %v", client.writes)
	}
	if client.project != "projects/example" || len(client.writes[0].Policy.Bindings) != 1 || len(client.writes[0].Policy.Bindings[0].Members) != 2 {
		t.Fatal("update did not write the full policy to the project")
	}
	var result iamPolicyFrameworkModel
	if diags := response.State.Get(ctx, &result); diags.HasError() {
		t.Fatal(diags)
	}
	if result.Etag.ValueString() != "etag-next" || len(result.Bindings.Elements()) != 1 {
		t.Fatalf("unexpected state: %#v", result)
	}
}

func TestIAMPolicyFrameworkConflictIsNotRetried(t *testing.T) {
	ctx := context.Background()
	client := &frameworkIAMClient{writeError: errors.New("etag conflict")}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = client
	state := frameworkIAMState(t, frameworkIAMModel(t, "workspaces/test"))
	response := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: state.Schema, Raw: state.Raw}}, &response)
	if !response.Diagnostics.HasError() || len(client.writes) != 1 {
		t.Fatalf("expected an unretried conflict, got %v, %v", client.writes, response.Diagnostics)
	}
}

func TestIAMPolicyFrameworkReadPreservesKeysAndIncludesDrift(t *testing.T) {
	ctx := context.Background()
	client := &frameworkIAMClient{policy: &v1pb.IamPolicy{Etag: "etag-read", Bindings: []*v1pb.Binding{
		{Role: "roles/reader", Members: []string{"user:a@example.com"}},
		{Role: "roles/reader", Members: []string{"user:b@example.com"}},
		{Role: "roles/writer", Members: []string{"user:c@example.com"}},
	}}}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = client
	state := frameworkIAMState(t, frameworkIAMModel(t, "workspaces/test"))
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result iamPolicyFrameworkModel
	response.State.Get(ctx, &result)
	if len(result.Bindings.Elements()) != 2 {
		t.Fatalf("remote bindings were hidden: %s", result.Bindings)
	}
	reader := result.Bindings.Elements()["reader"].(types.Object)
	if len(reader.Attributes()["members"].(types.Set).Elements()) != 2 {
		t.Fatal("duplicate remote identities lost members")
	}
}

func TestIAMPolicyFrameworkUnknownValidation(t *testing.T) {
	ctx := context.Background()
	r := newIAMPolicyFrameworkResource().(resource.ResourceWithValidateConfig)
	for _, unknown := range []types.Map{types.MapUnknown(iamFrameworkBindingType()), types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"future": types.ObjectUnknown(iamFrameworkBindingType().AttrTypes)})} {
		model := frameworkIAMModel(t, "workspaces/test")
		model.Bindings = unknown
		state := frameworkIAMState(t, model)
		response := resource.ValidateConfigResponse{}
		r.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: state.Schema, Raw: state.Raw}}, &response)
		if response.Diagnostics.HasError() {
			t.Fatal(response.Diagnostics)
		}
	}
}

func TestIAMPolicyFrameworkImportAndDelete(t *testing.T) {
	ctx := context.Background()
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	schema := frameworkIAMSchema(t)
	state := tfsdk.State{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), nil)}
	importResponse := resource.ImportStateResponse{State: state}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "projects/example"}, &importResponse)
	if importResponse.Diagnostics.HasError() {
		t.Fatal(importResponse.Diagnostics)
	}
	var result iamPolicyFrameworkModel
	importResponse.State.Get(ctx, &result)
	if result.ID.ValueString() != "projects/example" || result.Parent.ValueString() != "projects/example" || result.Bindings.IsNull() {
		t.Fatalf("bad import state: %#v", result)
	}
	deleteResponse := resource.DeleteResponse{State: importResponse.State}
	r.Delete(ctx, resource.DeleteRequest{State: importResponse.State}, &deleteResponse)
	if deleteResponse.Diagnostics.HasError() || len(deleteResponse.Diagnostics.Warnings()) != 1 {
		t.Fatalf("delete must warn without changing remote policy: %v", deleteResponse.Diagnostics)
	}
}

func TestIAMPolicyFrameworkEmptyConditionRoundTrip(t *testing.T) {
	model := frameworkIAMModel(t, "workspaces/test")
	object := model.Bindings.Elements()["reader"].(types.Object)
	attributes := object.Attributes()
	empty := map[string]attr.Value{
		"database": types.StringNull(), "schema": types.StringNull(), "expire_timestamp": types.StringNull(),
		"tables":          types.SetValueMust(types.StringType, []attr.Value{}),
		"environment_ids": types.SetNull(types.StringType),
	}
	attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, empty)
	model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{
		"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes),
	})
	want := model.Bindings
	bindings, known, err := iamFrameworkBindings(model)
	if err != nil || !known {
		t.Fatalf("unexpected conversion: %v, %v", known, err)
	}
	if bindings["reader"].Condition != nil && bindings["reader"].Condition.Expression != "" {
		t.Fatalf("empty condition changed effective scope: %v", bindings["reader"].Condition)
	}
	if err := iamFrameworkSetPolicy(&model, "workspaces/test", &v1pb.IamPolicy{Bindings: []*v1pb.Binding{{Role: "roles/reader", Members: []string{"user:a@example.com"}}}}); err != nil {
		t.Fatal(err)
	}
	if !model.Bindings.Equal(want) {
		t.Fatalf("empty condition did not round trip: got %s, want %s", model.Bindings, want)
	}
}

func TestIAMPolicyFrameworkRejectsLossyRemoteConditions(t *testing.T) {
	for _, condition := range []*expr.Expr{
		{Expression: `resource.database != "instances/test/databases/main"`},
		{Expression: `resource.database == "instances/test/databases/main" || true`},
		{Expression: `resource.database == "instances/test/databases/main"`, Title: "important"},
	} {
		model := frameworkIAMModel(t, "workspaces/test")
		err := iamFrameworkSetPolicy(&model, "workspaces/test", &v1pb.IamPolicy{Bindings: []*v1pb.Binding{{Role: "roles/reader", Members: []string{"user:a@example.com"}, Condition: condition}}})
		if err == nil {
			t.Fatalf("unsupported condition was silently weakened: %v", condition)
		}
	}
}

func TestIAMPolicyFrameworkLegacyPlanClearsComputedMap(t *testing.T) {
	ctx := context.Background()
	model := frameworkIAMModel(t, "workspaces/test")
	model.Bindings = types.MapNull(iamFrameworkBindingType())
	model.IAMPolicy = types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{
		types.ObjectValueMust(iamFrameworkLegacyPolicyType().AttrTypes, map[string]attr.Value{"binding": types.SetValueMust(iamFrameworkLegacyBindingType(), []attr.Value{})}),
	})
	configured := frameworkIAMState(t, model)
	plannedModel := model
	plannedModel.Bindings = types.MapUnknown(iamFrameworkBindingType())
	planned := frameworkIAMState(t, plannedModel)
	response := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}
	r := newIAMPolicyFrameworkResource().(resource.ResourceWithModifyPlan)
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: configured.Schema, Raw: configured.Raw},
		Plan:   tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw},
	}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result iamPolicyFrameworkModel
	response.Plan.Get(ctx, &result)
	if !result.Bindings.IsNull() {
		t.Fatal("legacy policy accidentally adopted the optional/computed map")
	}
}

func TestIAMPolicyFrameworkValidatesDuplicateAndConflictingInputs(t *testing.T) {
	ctx := context.Background()
	for _, conflict := range []bool{false, true} {
		model := frameworkIAMModel(t, "workspaces/test")
		values := model.Bindings.Elements()
		values["duplicate"] = values["reader"]
		model.Bindings = types.MapValueMust(iamFrameworkBindingType(), values)
		if conflict {
			model.IAMPolicy = types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{
				types.ObjectValueMust(iamFrameworkLegacyPolicyType().AttrTypes, map[string]attr.Value{"binding": types.SetValueMust(iamFrameworkLegacyBindingType(), []attr.Value{})}),
			})
		}
		state := frameworkIAMState(t, model)
		response := resource.ValidateConfigResponse{}
		newIAMPolicyFrameworkResource().(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: state.Schema, Raw: state.Raw}}, &response)
		if !response.Diagnostics.HasError() {
			t.Fatalf("expected invalid configuration (conflict=%v) to be rejected", conflict)
		}
	}
}

func TestIAMPolicyFrameworkUnknownNestedValidation(t *testing.T) {
	ctx := context.Background()
	for _, field := range []string{"role", "members", "condition"} {
		model := frameworkIAMModel(t, "workspaces/test")
		attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
		switch field {
		case "role":
			attributes[field] = types.StringUnknown()
		case "members":
			attributes[field] = types.SetUnknown(types.StringType)
		case "condition":
			attributes[field] = types.ObjectUnknown(iamFrameworkConditionType().AttrTypes)
		}
		model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
		state := frameworkIAMState(t, model)
		response := resource.ValidateConfigResponse{}
		newIAMPolicyFrameworkResource().(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: state.Schema, Raw: state.Raw}}, &response)
		if response.Diagnostics.HasError() {
			t.Fatalf("%s should defer validation until known: %v", field, response.Diagnostics)
		}
	}
}

func TestIAMPolicyFrameworkDatabaseConditionChecksRolePermission(t *testing.T) {
	ctx := context.Background()
	model := frameworkIAMModel(t, "workspaces/test")
	attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
	condition := map[string]attr.Value{
		"database": types.StringValue("instances/test/databases/main"), "schema": types.StringNull(), "expire_timestamp": types.StringNull(),
		"tables": types.SetNull(types.StringType), "environment_ids": types.SetNull(types.StringType),
	}
	attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, condition)
	model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
	client := &frameworkIAMClient{}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = client
	if err := r.upsert(ctx, &model, "etag-old"); err == nil || len(client.writes) != 0 {
		t.Fatal("role without SQL permission wrote a database condition")
	}
	client.permissions = []string{"bb.sql.export"}
	if err := r.upsert(ctx, &model, "etag-old"); err != nil {
		t.Fatal(err)
	}
}

func TestIAMPolicyFrameworkParentChangesDoNotRetainOldPlannedID(t *testing.T) {
	ctx := context.Background()
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = &frameworkIAMClient{}
	previous := frameworkIAMState(t, frameworkIAMModel(t, "projects/old"))
	model := frameworkIAMModel(t, "projects/new")
	configured := frameworkIAMState(t, model)
	model.ID = types.StringUnknown()
	planned := frameworkIAMState(t, model)
	response := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		State: previous, Config: tfsdk.Config{Schema: configured.Schema, Raw: configured.Raw},
		Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw},
	}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result iamPolicyFrameworkModel
	response.Plan.Get(ctx, &result)
	if !result.ID.IsUnknown() {
		t.Fatalf("old ID was retained for a changed target: %s", result.ID)
	}
}

func TestIAMPolicyFrameworkPreservesConfiguredWorkspaceParent(t *testing.T) {
	ctx := context.Background()
	for _, configured := range []string{"", "workspaces/alias"} {
		model := frameworkIAMModel(t, configured)
		client := &frameworkIAMClient{}
		r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
		r.client = client
		if err := r.upsert(ctx, &model, "etag-old"); err != nil {
			t.Fatal(err)
		}
		if model.Parent.ValueString() != configured || model.ID.ValueString() != "workspaces/test" || client.writes[0].Resource != "workspaces/test" {
			t.Fatalf("configured parent must be preserved while API/ID resolve to the workspace: %#v", model)
		}
	}
}

func TestIAMPolicyFrameworkChangedParentUsesTargetEtag(t *testing.T) {
	ctx := context.Background()
	client := &frameworkIAMClient{projectPolicies: map[string]*v1pb.IamPolicy{
		"projects/old": {Etag: "etag-old"}, "projects/new": {Etag: "etag-new"},
	}}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = client
	previous := frameworkIAMState(t, frameworkIAMModel(t, "projects/old"))
	model := frameworkIAMModel(t, "projects/new")
	model.Etag = types.StringUnknown()
	model.ID = types.StringUnknown()
	planned := frameworkIAMState(t, model)
	response := resource.UpdateResponse{State: previous}
	r.Update(ctx, resource.UpdateRequest{State: previous, Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	if len(client.writes) != 1 || client.writes[0].Policy.Etag != "etag-new" {
		t.Fatalf("changed parent used another policy's etag: %v", client.writes)
	}
}

func TestIAMPolicyFrameworkScopedConditionRequiresDatabase(t *testing.T) {
	ctx := context.Background()
	for _, field := range []string{"schema", "tables"} {
		model := frameworkIAMModel(t, "workspaces/test")
		attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
		condition := map[string]attr.Value{
			"database": types.StringNull(), "schema": types.StringNull(), "expire_timestamp": types.StringNull(),
			"tables": types.SetNull(types.StringType), "environment_ids": types.SetNull(types.StringType),
		}
		if field == "schema" {
			condition[field] = types.StringValue("sales")
		} else {
			condition[field] = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("orders")})
		}
		attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, condition)
		model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
		client := &frameworkIAMClient{permissions: []string{"bb.sql.select"}}
		r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
		r.client = client
		if err := r.upsert(ctx, &model, "etag-old"); err == nil || len(client.writes) != 0 {
			t.Fatalf("%s without database silently became unconditional", field)
		}
	}
}

func TestIAMPolicyFrameworkCreateReadsExistingPolicyEtag(t *testing.T) {
	ctx := context.Background()
	client := &frameworkIAMClient{policy: &v1pb.IamPolicy{Etag: "etag-existing"}}
	r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
	r.client = client
	model := frameworkIAMModel(t, "workspaces/test")
	model.Etag = types.StringUnknown()
	model.ID = types.StringUnknown()
	planned := frameworkIAMState(t, model)
	response := resource.CreateResponse{State: planned}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	if len(client.writes) != 1 || client.writes[0].Policy.Etag != "etag-existing" {
		t.Fatalf("initial write was not etag protected: %v", client.writes)
	}
}

func TestIAMPolicyFrameworkPreservesHashCollidingMembers(t *testing.T) {
	model := frameworkIAMModel(t, "workspaces/test")
	model.Bindings = frameworkIAMMap(t, "roles/reader", "user:zutdocbjiv@example.com", "user:xvmxlgbdlh@example.com")
	bindings, _, err := iamFrameworkBindings(model)
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings["reader"].Members) != 2 {
		t.Fatalf("native Terraform set lost a valid member to SDK hash collision: %v", bindings["reader"].Members)
	}
}

func TestIAMPolicyFrameworkPreciseConditionRoundTrip(t *testing.T) {
	model := frameworkIAMModel(t, "workspaces/test")
	attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
	condition := map[string]attr.Value{
		"database": types.StringValue("instances/test/databases/main"), "schema": types.StringNull(),
		"expire_timestamp": types.StringValue("2030-01-01T00:00:00.123Z"),
		"tables":           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("user:zutdocbjiv@example.com"), types.StringValue("user:xvmxlgbdlh@example.com")}),
		"environment_ids":  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("user:zutdocbjiv@example.com"), types.StringValue("user:xvmxlgbdlh@example.com")}),
	}
	attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, condition)
	model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
	want := model.Bindings
	bindings, _, err := iamFrameworkBindings(model)
	if err != nil {
		t.Fatal(err)
	}
	if err := iamFrameworkSetPolicy(&model, "workspaces/test", &v1pb.IamPolicy{Bindings: []*v1pb.Binding{bindings["reader"]}}); err != nil {
		t.Fatal(err)
	}
	if !model.Bindings.Equal(want) {
		t.Fatalf("condition values were changed or discarded: got %s, want %s", model.Bindings, want)
	}
}

func TestIAMPolicyFrameworkEmptyEnvironmentSetIsNotUnconditional(t *testing.T) {
	model := frameworkIAMModel(t, "workspaces/test")
	attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
	condition := map[string]attr.Value{
		"database": types.StringNull(), "schema": types.StringNull(), "expire_timestamp": types.StringNull(),
		"tables": types.SetNull(types.StringType), "environment_ids": types.SetValueMust(types.StringType, []attr.Value{}),
	}
	attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, condition)
	model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
	want := model.Bindings
	bindings, _, err := iamFrameworkBindings(model)
	if err != nil {
		t.Fatal(err)
	}
	if bindings["reader"].GetCondition().GetExpression() != "resource.environment_id in []" {
		t.Fatal("explicit empty environment restriction became unconditional")
	}
	if err := iamFrameworkSetPolicy(&model, "workspaces/test", &v1pb.IamPolicy{Bindings: []*v1pb.Binding{bindings["reader"]}}); err != nil {
		t.Fatal(err)
	}
	if !model.Bindings.Equal(want) {
		t.Fatalf("explicit empty environment set did not round trip: %s", model.Bindings)
	}
}

func TestIAMPolicyFrameworkMissingEtagRequiresRefresh(t *testing.T) {
	ctx := context.Background()
	for _, etag := range []types.String{types.StringNull(), types.StringUnknown()} {
		model := frameworkIAMModel(t, "workspaces/test")
		model.Etag = etag
		state := frameworkIAMState(t, model)
		client := &frameworkIAMClient{}
		r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
		r.client = client
		response := resource.UpdateResponse{State: state}
		r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: state.Schema, Raw: state.Raw}}, &response)
		if !response.Diagnostics.HasError() || len(client.writes) != 0 {
			t.Fatal("update without a known etag should require refresh rather than write unprotected")
		}
	}
}

func TestIAMPolicyFrameworkNullConditionSetElementsDoNotWrite(t *testing.T) {
	ctx := context.Background()
	for _, field := range []string{"tables", "environment_ids"} {
		model := frameworkIAMModel(t, "workspaces/test")
		attributes := model.Bindings.Elements()["reader"].(types.Object).Attributes()
		condition := map[string]attr.Value{
			"database": types.StringValue("instances/test/databases/main"), "schema": types.StringNull(), "expire_timestamp": types.StringNull(),
			"tables": types.SetNull(types.StringType), "environment_ids": types.SetNull(types.StringType),
		}
		condition[field] = types.SetValueMust(types.StringType, []attr.Value{types.StringNull()})
		attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, condition)
		model.Bindings = types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{"reader": types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)})
		client := &frameworkIAMClient{permissions: []string{"bb.sql.select"}}
		r := newIAMPolicyFrameworkResource().(*iamPolicyFrameworkResource)
		r.client = client
		if err := r.upsert(ctx, &model, "etag-old"); err == nil || len(client.writes) != 0 {
			t.Fatalf("null %s element was silently changed to an empty string", field)
		}
	}
}
