package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/type/expr"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func resourceIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "The IAM policy resource.",
		CreateContext: resourceIAMPolicyUpsert,
		ReadContext:   dataSourceIAMPolicyRead,
		UpdateContext: resourceIAMPolicyUpsert,
		DeleteContext: resourceIAMPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: internal.ResourceNameValidation(
					// workspace policy
					fmt.Sprintf("^%s$", internal.WorkspaceName),
					// project policy
					fmt.Sprintf("^%s%s$", internal.ProjectNamePrefix, internal.ResourceIDPattern),
				),
				Description: `The IAM policy parent name for the policy, support "projects/{resource id}" or "workspaces/-"`,
			},
			"iam_policy": getIAMPolicySchema(false),
		},
	}
}

func resourceIAMPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(api.Client)
	parent := d.Get("parent").(string)

	iamPolicy, err := convertToIAMPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}
	request := &v1pb.SetIamPolicyRequest{
		Resource: parent,
		Policy:   iamPolicy,
	}
	if strings.HasPrefix(parent, internal.ProjectNamePrefix) {
		if _, err := c.SetProjectIAMPolicy(ctx, parent, request); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if _, err := c.SetWorkspaceIAMPolicy(ctx, request); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(parent)
	return dataSourceIAMPolicyRead(ctx, d, m)
}

func resourceIAMPolicyDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Unsupport delete IAM policy",
	})
	d.SetId("")

	return diags
}

func convertToV1Binding(rawSchema interface{}) (*v1pb.Binding, error) {
	rawBinding := rawSchema.(map[string]interface{})

	role := rawBinding["role"].(string)
	if !strings.HasPrefix(role, internal.RoleNamePrefix) {
		return nil, errors.Errorf("invalid role format, role must in roles/{id} format")
	}

	binding := &v1pb.Binding{
		Role: role,
	}

	members, ok := rawBinding["members"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid members")
	}
	if members.Len() == 0 {
		return nil, errors.Errorf("empty members")
	}
	for _, member := range members.List() {
		if err := internal.ValidateMemberBinding(member.(string)); err != nil {
			return nil, errors.Wrapf(err, "invalid member: %v", member)
		}
		binding.Members = append(binding.Members, member.(string))
	}

	if condition, ok := rawBinding["condition"].(*schema.Set); ok {
		if condition.Len() > 1 {
			return nil, errors.Errorf("should only set one condition")
		}
		if condition.Len() == 1 && condition.List()[0] != nil {
			conditionExpr, err := convertToV1Condition(condition.List()[0])
			if err != nil {
				return nil, err
			}
			binding.Condition = conditionExpr
		}
	} else {
		binding.Condition = &expr.Expr{
			Expression: "",
		}
	}
	return binding, nil
}

func convertToV1Condition(rawSchema interface{}) (*expr.Expr, error) {
	rawCondition := rawSchema.(map[string]interface{})
	expressions := []string{}

	if database, ok := rawCondition["database"].(string); ok && database != "" {
		expressions = append(expressions, fmt.Sprintf(`resource.database == "%s"`, database))
	}
	if schema, ok := rawCondition["schema"].(string); ok {
		expressions = append(expressions, fmt.Sprintf(`resource.schema == "%s"`, schema))
	}
	if tables, ok := rawCondition["tables"].(*schema.Set); ok && tables.Len() > 0 {
		tableList := []string{}
		for _, table := range tables.List() {
			tableList = append(tableList, fmt.Sprintf(`"%s"`, table.(string)))
		}
		expressions = append(expressions, fmt.Sprintf(`resource.table in [%s]`, strings.Join(tableList, ",")))
	}
	if rowLimit, ok := rawCondition["row_limit"].(int); ok && rowLimit > 0 {
		expressions = append(expressions, fmt.Sprintf(`request.row_limit <= %d`, rowLimit))
	}
	if expire, ok := rawCondition["expire_timestamp"].(string); ok && expire != "" {
		formattedTime, err := time.Parse(time.RFC3339, expire)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid time: %v", expire)
		}
		expressions = append(expressions, fmt.Sprintf(`request.time < timestamp("%s")`, formattedTime.Format(time.RFC3339)))
	}

	return &expr.Expr{
		Expression: strings.Join(expressions, " && "),
	}, nil
}

func convertToIAMPolicy(d *schema.ResourceData) (*v1pb.IamPolicy, error) {
	rawList, ok := d.Get("iam_policy").([]interface{})
	if !ok || len(rawList) != 1 {
		return nil, errors.Errorf("invalid iam_policy")
	}

	raw := rawList[0].(map[string]interface{})
	bindingList, ok := raw["binding"].(*schema.Set)
	if !ok {
		return nil, errors.Errorf("invalid binding")
	}

	policy := &v1pb.IamPolicy{}

	for _, raw := range bindingList.List() {
		binding, err := convertToV1Binding(raw)
		if err != nil {
			return nil, err
		}
		policy.Bindings = append(policy.Bindings, binding)
	}
	return policy, nil
}
