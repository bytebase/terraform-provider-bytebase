package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/expr"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

var (
	_ resource.ResourceWithConfigure      = (*iamPolicyFrameworkResource)(nil)
	_ resource.ResourceWithValidateConfig = (*iamPolicyFrameworkResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*iamPolicyFrameworkResource)(nil)
	_ resource.ResourceWithImportState    = (*iamPolicyFrameworkResource)(nil)
)

type iamPolicyFrameworkResource struct{ client api.Client }

type iamPolicyFrameworkModel struct {
	ID        types.String `tfsdk:"id"`
	Parent    types.String `tfsdk:"parent"`
	Etag      types.String `tfsdk:"etag"`
	Bindings  types.Map    `tfsdk:"bindings"`
	IAMPolicy types.List   `tfsdk:"iam_policy"`
}

func newIAMPolicyFrameworkResource() resource.Resource { return &iamPolicyFrameworkResource{} }

func (*iamPolicyFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func iamFrameworkConditionType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"database": types.StringType, "schema": types.StringType, "expire_timestamp": types.StringType,
		"tables": types.SetType{ElemType: types.StringType}, "environment_ids": types.SetType{ElemType: types.StringType},
	}}
}

func iamFrameworkBindingType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"role": types.StringType, "members": types.SetType{ElemType: types.StringType}, "condition": iamFrameworkConditionType(),
	}}
}

func iamFrameworkLegacyBindingType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"role": types.StringType, "members": types.SetType{ElemType: types.StringType},
		"condition": types.SetType{ElemType: iamFrameworkConditionType()},
	}}
}

func iamFrameworkLegacyPolicyType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{"binding": types.SetType{ElemType: iamFrameworkLegacyBindingType()}}}
}

func iamFrameworkConditionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"database":         schema.StringAttribute{Optional: true, Description: "The accessible database full name in instances/{instance resource id}/databases/{database name} format. Only works for roles with bb.sql.x permissions."},
		"schema":           schema.StringAttribute{Optional: true, Description: "The accessible schema in the database. Must configure with the database."},
		"tables":           schema.SetAttribute{Optional: true, ElementType: types.StringType, Description: "The accessible tables. Must configure with the database."},
		"expire_timestamp": schema.StringAttribute{Optional: true, Description: "The expiration timestamp in RFC3339 format."},
		"environment_ids":  schema.SetAttribute{Optional: true, ElementType: types.StringType, Description: "The environment IDs allowed for DDL/DML operations in the SQL Editor."},
	}
}

func iamFrameworkLegacyConditionAttributes() map[string]schema.Attribute {
	attributes := iamFrameworkConditionAttributes()
	// Match SDK zero values without inheriting a removed nonempty condition field.
	for name, attribute := range attributes {
		switch attribute := attribute.(type) {
		case schema.StringAttribute:
			attribute.Computed = true
			attribute.Default = stringdefault.StaticString("")
			attributes[name] = attribute
		case schema.SetAttribute:
			attribute.Computed = true
			attribute.Default = setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{}))
			attributes[name] = attribute
		}
	}
	return attributes
}

func (*iamPolicyFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The authoritative IAM policy resource. Configure bindings or the legacy iam_policy block, but not both.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true},
			"parent": schema.StringAttribute{Optional: true, Computed: true, Description: `The policy parent: projects/{resource id} or workspaces/{workspace id}. Defaults to the workspace.`},
			"etag":   schema.StringAttribute{Computed: true, Description: "The server policy etag used to reject conflicting updates."},
			"bindings": schema.MapNestedAttribute{
				Optional: true, Computed: true,
				Description: "The complete policy, indexed by stable user-chosen binding keys.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"role":      schema.StringAttribute{Required: true, Description: "The role full name in roles/{id} format."},
					"members":   schema.SetAttribute{Required: true, ElementType: types.StringType, Description: "The complete set of members granted this role."},
					"condition": schema.SingleNestedAttribute{Optional: true, Attributes: iamFrameworkConditionAttributes()},
				}},
			},
		},
		Blocks: map[string]schema.Block{
			"iam_policy": schema.ListNestedBlock{
				Description: "The legacy complete IAM policy. Use bindings for keyed binding plans.",
				Validators:  []validator.List{iamFrameworkPolicySizeValidator{}},
				NestedObject: schema.NestedBlockObject{Blocks: map[string]schema.Block{
					"binding": schema.SetNestedBlock{NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"role":    schema.StringAttribute{Optional: true},
							"members": schema.SetAttribute{Optional: true, ElementType: types.StringType},
						},
						Blocks: map[string]schema.Block{
							"condition": schema.SetNestedBlock{NestedObject: schema.NestedBlockObject{Attributes: iamFrameworkLegacyConditionAttributes()}},
						},
					}},
				}},
			},
		},
	}
}

type iamFrameworkPolicySizeValidator struct{}

func (iamFrameworkPolicySizeValidator) Description(context.Context) string {
	return "must contain at most one policy block"
}
func (v iamFrameworkPolicySizeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (iamFrameworkPolicySizeValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if !req.ConfigValue.IsUnknown() && len(req.ConfigValue.Elements()) > 1 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid IAM policy", "Configure at most one iam_policy block.")
	}
}

func (r *iamPolicyFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (*iamPolicyFrameworkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var model iamPolicyFrameworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !model.Bindings.IsNull() && !model.IAMPolicy.IsUnknown() && len(model.IAMPolicy.Elements()) > 0 {
		resp.Diagnostics.AddAttributeError(path.Root("bindings"), "Conflicting IAM policy inputs", "Configure either bindings or iam_policy, not both.")
		return
	}
	if !model.Parent.IsNull() && !model.Parent.IsUnknown() && !regexp.MustCompile(`^(|projects/`+internal.ResourceIDPattern+`|workspaces/`+internal.ResourceIDPattern+`)$`).MatchString(model.Parent.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("parent"), "Invalid IAM policy parent", "Resource id doesn't must any patterns: projects/{resource id} or workspaces/{workspace id}.")
	}
	bindings, known, err := iamFrameworkBindings(model)
	if err != nil {
		resp.Diagnostics.AddError("Invalid IAM policy", err.Error())
		return
	}
	if known && !model.Bindings.IsNull() {
		if err := validateIAMBindingIdentities(bindings); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("bindings"), "Duplicate IAM bindings", err.Error())
		}
	}
}

func (r *iamPolicyFrameworkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var config iamPolicyFrameworkModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.Bindings.IsNull() && !config.IAMPolicy.IsUnknown() && len(config.IAMPolicy.Elements()) > 0 {
		// Optional/computed enables keyed import; legacy configurations must not adopt it.
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("bindings"), types.MapNull(iamFrameworkBindingType()))...)
	}
	if config.Parent.IsNull() && r.client != nil {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("parent"), types.StringValue(r.client.GetWorkspaceName()))...)
	}
	if r.client != nil {
		var previous iamPolicyFrameworkModel
		if !req.State.Raw.IsNull() {
			resp.Diagnostics.Append(req.State.Get(ctx, &previous)...)
		}
		if !config.Parent.IsUnknown() && !previous.ID.IsNull() && previous.ID.ValueString() == r.resolveParent(config) {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), previous.ID)...)
		}
	}
}

func (r *iamPolicyFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model iamPolicyFrameworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.getPolicy(ctx, r.resolveParent(model))
	if err != nil {
		resp.Diagnostics.AddError("Cannot read IAM policy before initial write", err.Error())
		return
	}
	if err := r.upsert(ctx, &model, current.GetEtag()); err != nil {
		resp.Diagnostics.AddError("Cannot set IAM policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *iamPolicyFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model, previous iamPolicyFrameworkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &previous)...)
	if resp.Diagnostics.HasError() {
		return
	}
	etag := previous.Etag.ValueString()
	if r.resolveParent(model) != r.resolveParent(previous) {
		current, err := r.getPolicy(ctx, r.resolveParent(model))
		if err != nil {
			resp.Diagnostics.AddError("Cannot read target IAM policy before update", err.Error())
			return
		}
		etag = current.GetEtag()
	} else if previous.Etag.IsNull() || previous.Etag.IsUnknown() {
		resp.Diagnostics.AddError("IAM policy refresh required", "Refresh this resource before updating it so Terraform can obtain the policy etag and detect conflicting writes.")
		return
	}
	if err := r.upsert(ctx, &model, etag); err != nil {
		resp.Diagnostics.AddError("Cannot set IAM policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *iamPolicyFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model iamPolicyFrameworkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	parent := r.resolveParent(model)
	policy, err := r.getPolicy(ctx, parent)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read IAM policy", err.Error())
		return
	}
	if err := iamFrameworkSetPolicy(&model, parent, policy); err != nil {
		resp.Diagnostics.AddError("Cannot read IAM policy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (*iamPolicyFrameworkResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Unsupport delete IAM policy", "The policy is removed from Terraform state without changing its grants in Bytebase.")
	resp.State.RemoveResource(ctx)
}

func (*iamPolicyFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := iamPolicyFrameworkModel{
		ID: types.StringValue(req.ID), Parent: types.StringValue(req.ID), Etag: types.StringNull(),
		Bindings:  types.MapValueMust(iamFrameworkBindingType(), map[string]attr.Value{}),
		IAMPolicy: types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{}),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *iamPolicyFrameworkResource) resolveParent(model iamPolicyFrameworkModel) string {
	return internal.ResolveWorkspaceParent(model.Parent.ValueString(), r.client.GetWorkspaceName())
}

func (r *iamPolicyFrameworkResource) getPolicy(ctx context.Context, parent string) (*v1pb.IamPolicy, error) {
	var policy *v1pb.IamPolicy
	var err error
	if strings.HasPrefix(parent, internal.ProjectNamePrefix) {
		policy, err = r.client.GetProjectIAMPolicy(ctx, parent)
	} else {
		policy, err = r.client.GetWorkspaceIAMPolicy(ctx)
	}
	if err == nil && policy == nil {
		return nil, fmt.Errorf("server returned an empty IAM policy")
	}
	return policy, err
}

func (r *iamPolicyFrameworkResource) upsert(ctx context.Context, model *iamPolicyFrameworkModel, etag string) error {
	bindings, known, err := iamFrameworkBindings(*model)
	if err != nil {
		return err
	}
	if !known {
		return fmt.Errorf("IAM policy bindings must be known before apply")
	}
	if !model.Bindings.IsNull() {
		if err := validateIAMBindingIdentities(bindings); err != nil {
			return err
		}
	}
	policy := &v1pb.IamPolicy{Etag: etag}
	keys := make([]string, 0, len(bindings))
	for key := range bindings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		binding := bindings[key]
		role, err := r.client.GetRole(ctx, binding.Role)
		if err != nil {
			return fmt.Errorf("failed to get role %s: %w", binding.Role, err)
		}
		if binding.Condition != nil && strings.Contains(binding.Condition.Expression, internal.CELAttributeResourceDatabase) && !roleContainsAnyPermission(role,
			"bb.sql.select", "bb.sql.export", "bb.sql.ddl", "bb.sql.dml", "bb.sql.explain", "bb.sql.info") {
			return fmt.Errorf(`role "%s" without "bb.sql." permissions shouldn't configure the database condition`, binding.Role)
		}
		policy.Bindings = append(policy.Bindings, binding)
	}
	parent := r.resolveParent(*model)
	request := &v1pb.SetIamPolicyRequest{Resource: parent, Policy: policy}
	var updated *v1pb.IamPolicy
	if strings.HasPrefix(parent, internal.ProjectNamePrefix) {
		updated, err = r.client.SetProjectIAMPolicy(ctx, parent, request)
	} else {
		updated, err = r.client.SetWorkspaceIAMPolicy(ctx, request)
	}
	if err != nil {
		return err
	}
	return iamFrameworkSetPolicy(model, parent, updated)
}

// Unknowns are expected during validation and planning, including nested values.
func iamFrameworkValueKnown(value attr.Value) bool {
	if value.IsUnknown() {
		return false
	}
	if value.IsNull() {
		return true
	}
	switch value := value.(type) {
	case types.Object:
		for _, child := range value.Attributes() {
			if !iamFrameworkValueKnown(child) {
				return false
			}
		}
	case types.Map:
		for _, child := range value.Elements() {
			if !iamFrameworkValueKnown(child) {
				return false
			}
		}
	case types.List:
		for _, child := range value.Elements() {
			if !iamFrameworkValueKnown(child) {
				return false
			}
		}
	case types.Set:
		for _, child := range value.Elements() {
			if !iamFrameworkValueKnown(child) {
				return false
			}
		}
	}
	return true
}

func iamFrameworkBindings(model iamPolicyFrameworkModel) (map[string]*v1pb.Binding, bool, error) {
	bindings := map[string]*v1pb.Binding{}
	if !model.Bindings.IsNull() {
		if !iamFrameworkValueKnown(model.Bindings) {
			return bindings, false, nil
		}
		for key, value := range model.Bindings.Elements() {
			if value.IsNull() {
				return nil, true, fmt.Errorf("binding %q must not be null", key)
			}
			binding, err := iamFrameworkBinding(value.(types.Object), false)
			if err != nil {
				return nil, true, fmt.Errorf("binding %q: %w", key, err)
			}
			bindings[key] = binding
		}
		return bindings, true, nil
	}
	if !iamFrameworkValueKnown(model.IAMPolicy) {
		return bindings, false, nil
	}
	if len(model.IAMPolicy.Elements()) != 1 {
		return nil, true, fmt.Errorf("configure bindings or exactly one iam_policy block")
	}
	policy := model.IAMPolicy.Elements()[0].(types.Object)
	for i, value := range policy.Attributes()["binding"].(types.Set).Elements() {
		binding, err := iamFrameworkBinding(value.(types.Object), true)
		if err != nil {
			return nil, true, err
		}
		bindings[fmt.Sprintf("%d", i)] = binding
	}
	return bindings, true, nil
}

func iamFrameworkStringSet(value types.Set) *sdkschema.Set {
	values := make([]interface{}, 0, len(value.Elements()))
	for _, element := range value.Elements() {
		values = append(values, element.(types.String).ValueString())
	}
	return sdkschema.NewSet(sdkschema.HashString, values)
}

func iamFrameworkBinding(object types.Object, legacy bool) (*v1pb.Binding, error) {
	if !legacy {
		return iamFrameworkKeyedBinding(object)
	}
	attributes := object.Attributes()
	role := attributes["role"].(types.String).ValueString()
	members := attributes["members"].(types.Set)
	for _, member := range members.Elements() {
		value := member.(types.String).ValueString()
		if value != "allUsers" && !strings.HasPrefix(value, "user:") && !strings.HasPrefix(value, "group:") && !strings.HasPrefix(value, "serviceAccount:") && !strings.HasPrefix(value, "workloadIdentity:") {
			return nil, fmt.Errorf("invalid member %q", value)
		}
	}
	raw := map[string]interface{}{"role": role, "members": iamFrameworkStringSet(members)}
	var condition types.Object
	conditionSet := attributes["condition"].(types.Set)
	if len(conditionSet.Elements()) > 1 {
		return nil, fmt.Errorf("should only set one condition")
	}
	if len(conditionSet.Elements()) == 1 {
		condition = conditionSet.Elements()[0].(types.Object)
	}
	conditions := []interface{}{}
	if !condition.IsNull() && condition.AttributeTypes(context.Background()) != nil {
		rawCondition := map[string]interface{}{}
		for name, value := range condition.Attributes() {
			switch value := value.(type) {
			case types.String:
				rawCondition[name] = value.ValueString()
			case types.Set:
				rawCondition[name] = iamFrameworkStringSet(value)
			}
		}
		if database, _ := rawCondition["database"].(string); database == "" {
			schema, _ := rawCondition["schema"].(string)
			tables, _ := rawCondition["tables"].(*sdkschema.Set)
			if schema != "" || tables != nil && tables.Len() > 0 {
				return nil, fmt.Errorf("condition schema and tables require a nonempty database")
			}
		}
		conditions = append(conditions, rawCondition)
	}
	raw["condition"] = sdkschema.NewSet(conditionHash, conditions)
	return convertToV1Binding(raw)
}

func iamFrameworkKeyedBinding(object types.Object) (*v1pb.Binding, error) {
	attributes := object.Attributes()
	role := attributes["role"].(types.String).ValueString()
	if !strings.HasPrefix(role, internal.RoleNamePrefix) {
		return nil, fmt.Errorf("invalid role format, role must in roles/{id} format")
	}
	binding := &v1pb.Binding{Role: role}
	for _, member := range attributes["members"].(types.Set).Elements() {
		value := member.(types.String).ValueString()
		if value != "allUsers" && !strings.HasPrefix(value, "user:") && !strings.HasPrefix(value, "group:") && !strings.HasPrefix(value, "serviceAccount:") && !strings.HasPrefix(value, "workloadIdentity:") {
			return nil, fmt.Errorf("invalid member %q", value)
		}
		binding.Members = append(binding.Members, value)
	}
	if len(binding.Members) == 0 {
		return nil, fmt.Errorf("empty members")
	}
	sort.Strings(binding.Members)
	condition := attributes["condition"].(types.Object)
	if condition.IsNull() {
		return binding, nil
	}
	conditionAttributes := condition.Attributes()
	for _, field := range []string{"tables", "environment_ids"} {
		for _, element := range conditionAttributes[field].(types.Set).Elements() {
			if element.IsNull() {
				return nil, fmt.Errorf("condition %s must not contain null elements", field)
			}
		}
	}
	database := conditionAttributes["database"].(types.String).ValueString()
	schema := conditionAttributes["schema"].(types.String).ValueString()
	tables := conditionAttributes["tables"].(types.Set)
	if database == "" && (schema != "" || len(tables.Elements()) > 0) {
		return nil, fmt.Errorf("condition schema and tables require a nonempty database")
	}
	clauses := []string{}
	appendString := func(attribute, value string) {
		encoded, _ := json.Marshal(value)
		clauses = append(clauses, attribute+" == "+string(encoded))
	}
	appendSet := func(attribute string, set types.Set, includeEmpty bool) {
		if set.IsNull() || len(set.Elements()) == 0 && !includeEmpty {
			return
		}
		values := make([]string, 0, len(set.Elements()))
		for _, item := range set.Elements() {
			values = append(values, item.(types.String).ValueString())
		}
		sort.Strings(values)
		encoded, _ := json.Marshal(values)
		clauses = append(clauses, attribute+" in "+string(encoded))
	}
	if database != "" {
		appendString(internal.CELAttributeResourceDatabase, database)
		if schema != "" {
			appendString(internal.CELAttributeResourceSchemaName, schema)
		}
		appendSet(internal.CELAttributeResourceTableName, tables, false)
	}
	if expiry := conditionAttributes["expire_timestamp"].(types.String).ValueString(); expiry != "" {
		parsed, err := time.Parse(time.RFC3339Nano, expiry)
		if err != nil {
			return nil, fmt.Errorf("invalid time %q: %w", expiry, err)
		}
		encoded, _ := json.Marshal(parsed.Format(time.RFC3339Nano))
		clauses = append(clauses, internal.CELAttributeRequestTime+" < timestamp("+string(encoded)+")")
	}
	appendSet(internal.CELAttributeResourceEnvironmentID, conditionAttributes["environment_ids"].(types.Set), true)
	if len(clauses) > 0 {
		binding.Condition = &expr.Expr{Expression: strings.Join(clauses, " && ")}
	}
	return binding, nil
}

func iamFrameworkSetPolicy(model *iamPolicyFrameworkModel, parent string, policy *v1pb.IamPolicy) error {
	if policy == nil {
		return fmt.Errorf("server returned an empty IAM policy")
	}
	model.ID = types.StringValue(parent)
	if model.Parent.IsNull() || model.Parent.IsUnknown() {
		model.Parent = types.StringValue(parent)
	}
	model.Etag = types.StringValue(policy.Etag)
	if !model.Bindings.IsNull() {
		previous, _, err := iamFrameworkBindings(*model)
		if err != nil {
			return err
		}
		values := map[string]attr.Value{}
		for key, binding := range reconcileIAMBindings(policy.Bindings, previous) {
			var template types.Object
			if old, ok := model.Bindings.Elements()[key]; ok {
				template = old.(types.Object)
			}
			value, err := iamFrameworkFlattenBinding(binding, template, false)
			if err != nil {
				return err
			}
			values[key] = value
		}
		model.Bindings = types.MapValueMust(iamFrameworkBindingType(), values)
		model.IAMPolicy = types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{})
		return nil
	}
	values := []attr.Value{}
	for _, binding := range policy.Bindings {
		value, err := iamFrameworkFlattenBinding(binding, types.Object{}, true)
		if err != nil {
			return err
		}
		values = append(values, value)
	}
	model.IAMPolicy = types.ListValueMust(iamFrameworkLegacyPolicyType(), []attr.Value{
		types.ObjectValueMust(iamFrameworkLegacyPolicyType().AttrTypes, map[string]attr.Value{"binding": types.SetValueMust(iamFrameworkLegacyBindingType(), values)}),
	})
	return nil
}

func iamFrameworkFlattenBinding(binding *v1pb.Binding, template types.Object, legacy bool) (types.Object, error) {
	if !legacy {
		return iamFrameworkFlattenKeyedBinding(binding, template)
	}
	flattened, err := flattenIAMPolicy(&v1pb.IamPolicy{Bindings: []*v1pb.Binding{binding}})
	if err != nil {
		return types.Object{}, err
	}
	raw := flattened[0].(map[string]interface{})["binding"].(*sdkschema.Set).List()[0].(map[string]interface{})
	members := []attr.Value{}
	for _, member := range binding.Members {
		members = append(members, types.StringValue(member))
	}
	attributes := map[string]attr.Value{"role": types.StringValue(binding.Role), "members": types.SetValueMust(types.StringType, members)}
	condition := types.ObjectNull(iamFrameworkConditionType().AttrTypes)
	if set, ok := raw["condition"].(*sdkschema.Set); ok && set.Len() == 1 {
		conditionAttrs := map[string]attr.Value{}
		rawCondition := set.List()[0].(map[string]interface{})
		for name, kind := range iamFrameworkConditionType().AttrTypes {
			if kind.Equal(types.StringType) {
				conditionAttrs[name] = types.StringValue("")
				if value, ok := rawCondition[name].(string); ok {
					conditionAttrs[name] = types.StringValue(value)
				}
			} else {
				conditionAttrs[name] = types.SetValueMust(types.StringType, []attr.Value{})
				if value, ok := rawCondition[name].(*sdkschema.Set); ok {
					values := []attr.Value{}
					for _, item := range value.List() {
						values = append(values, types.StringValue(item.(string)))
					}
					conditionAttrs[name] = types.SetValueMust(types.StringType, values)
				}
			}
		}
		condition = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, conditionAttrs)
	}
	conditions := []attr.Value{}
	if !condition.IsNull() {
		conditions = append(conditions, condition)
	}
	attributes["condition"] = types.SetValueMust(iamFrameworkConditionType(), conditions)
	return types.ObjectValueMust(iamFrameworkLegacyBindingType().AttrTypes, attributes), nil
}

func iamFrameworkFlattenKeyedBinding(binding *v1pb.Binding, template types.Object) (types.Object, error) {
	condition := binding.GetCondition()
	unsupported := func() (types.Object, error) {
		return types.Object{}, fmt.Errorf("binding for %q has a condition that cannot be represented by bindings; review and update the remote policy to use supported structured conditions before managing it with this resource", binding.Role)
	}
	parsed, ok := parseIAMConditionExpression(condition.GetExpression())
	if !ok || condition.GetTitle() != "" || condition.GetDescription() != "" {
		return unsupported()
	}
	members := make([]attr.Value, 0, len(binding.Members))
	for _, member := range binding.Members {
		members = append(members, types.StringValue(member))
	}
	attributes := map[string]attr.Value{
		"role": types.StringValue(binding.Role), "members": types.SetValueMust(types.StringType, members),
		"condition": types.ObjectNull(iamFrameworkConditionType().AttrTypes),
	}
	var previous types.Object
	if !template.IsNull() && template.AttributeTypes(context.Background()) != nil {
		previous = template.Attributes()["condition"].(types.Object)
	}
	if len(parsed) > 0 || !previous.IsNull() && previous.AttributeTypes(context.Background()) != nil {
		conditionAttributes := map[string]attr.Value{}
		for name, kind := range iamFrameworkConditionType().AttrTypes {
			if kind.Equal(types.StringType) {
				conditionAttributes[name] = types.StringNull()
				if value, ok := parsed[name].(string); ok {
					conditionAttributes[name] = types.StringValue(value)
				}
			} else {
				conditionAttributes[name] = types.SetNull(types.StringType)
				if values, ok := parsed[name].([]string); ok {
					elements := make([]attr.Value, 0, len(values))
					for _, value := range values {
						elements = append(elements, types.StringValue(value))
					}
					conditionAttributes[name] = types.SetValueMust(types.StringType, elements)
				}
			}
			if previous.IsNull() || previous.AttributeTypes(context.Background()) == nil {
				continue
			}
			prior := previous.Attributes()[name]
			if conditionAttributes[name].IsNull() {
				if value, ok := prior.(types.String); ok && !value.IsUnknown() && value.ValueString() == "" {
					conditionAttributes[name] = prior
				}
				if value, ok := prior.(types.Set); ok && !value.IsUnknown() && len(value.Elements()) == 0 {
					conditionAttributes[name] = prior
				}
			} else if name == "expire_timestamp" && !prior.IsNull() && !prior.IsUnknown() {
				currentTime, currentErr := time.Parse(time.RFC3339Nano, conditionAttributes[name].(types.String).ValueString())
				previousTime, previousErr := time.Parse(time.RFC3339Nano, prior.(types.String).ValueString())
				if currentErr == nil && previousErr == nil && currentTime.Equal(previousTime) {
					conditionAttributes[name] = prior
				}
			}
		}
		attributes["condition"] = types.ObjectValueMust(iamFrameworkConditionType().AttrTypes, conditionAttributes)
	}
	object := types.ObjectValueMust(iamFrameworkBindingType().AttrTypes, attributes)
	reconstructed, err := iamFrameworkKeyedBinding(object)
	if err != nil || iamBindingIdentity(binding) != iamBindingIdentity(reconstructed) {
		return unsupported()
	}
	return object, nil
}
