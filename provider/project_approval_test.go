package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const projectApprovalField = "allow_last_plan_editor_approval"

func TestProjectLastPlanEditorApprovalSchema(t *testing.T) {
	resourceSchema := resourceProjct().Schema[projectApprovalField]
	if resourceSchema == nil {
		t.Fatal("project resource is missing allow_last_plan_editor_approval")
	}
	if resourceSchema.Type != schema.TypeBool || !resourceSchema.Optional || !resourceSchema.Computed || resourceSchema.Default != nil {
		t.Fatal("project approval must be an optional, computed boolean without a default so omission preserves the server value")
	}
	listSchema := dataSourceProjectList().Schema["projects"].Elem.(*schema.Resource).Schema
	for name, fields := range map[string]map[string]*schema.Schema{
		"project":      dataSourceProject().Schema,
		"project_list": listSchema,
	} {
		field := fields[projectApprovalField]
		if field == nil || field.Type != schema.TypeBool || !field.Computed || field.Optional {
			t.Errorf("%s approval must be a computed-only boolean", name)
		}
	}
}

func TestProjectLastPlanEditorApprovalCreate(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			client := &projectApprovalClient{}
			state := applyProjectApprovalConfig(t, nil, client, "Approval project", &value)
			if client.created == nil || client.created.AllowLastPlanEditorApproval != value {
				t.Fatalf("create request approval = %v, want %v", client.created, value)
			}
			if !slices.Contains(client.updateMasks, projectApprovalField) {
				t.Fatalf("create update mask = %v, want explicit approval field", client.updateMasks)
			}
			if got := state.Attributes[projectApprovalField]; got != fmt.Sprint(value) {
				t.Fatalf("approval state = %q, want %v", got, value)
			}
		})
	}
}

func TestProjectLastPlanEditorApprovalUpdateTrueToFalse(t *testing.T) {
	client := &projectApprovalClient{}
	enabled := true
	state := applyProjectApprovalConfig(t, nil, client, "Approval project", &enabled)
	client.updateMasks = nil
	disabled := false
	state = applyProjectApprovalConfig(t, state, client, "Approval project", &disabled)
	if !slices.Contains(client.updateMasks, projectApprovalField) {
		t.Fatalf("update mask = %v, want approval field when clearing true to false", client.updateMasks)
	}
	if client.project.AllowLastPlanEditorApproval || state.Attributes[projectApprovalField] != "false" {
		t.Fatalf("approval was not cleared: server = %v, state = %q", client.project.AllowLastPlanEditorApproval, state.Attributes[projectApprovalField])
	}
}

func TestProjectLastPlanEditorApprovalOmittedPreservesServerValue(t *testing.T) {
	client := &projectApprovalClient{project: &v1pb.Project{
		Name:                        "projects/project-approval",
		Title:                       "Approval project",
		State:                       v1pb.State_ACTIVE,
		AllowLastPlanEditorApproval: true,
	}}
	state := applyProjectApprovalConfig(t, nil, client, "Approval project", nil)
	if slices.Contains(client.updateMasks, projectApprovalField) || !client.project.AllowLastPlanEditorApproval {
		t.Fatalf("adopting existing project changed omitted approval: mask = %v, value = %v", client.updateMasks, client.project.AllowLastPlanEditorApproval)
	}
	client.updateMasks = nil
	state = applyProjectApprovalConfig(t, state, client, "Updated project", nil)
	if slices.Contains(client.updateMasks, projectApprovalField) || !client.project.AllowLastPlanEditorApproval {
		t.Fatalf("updating title changed omitted approval: mask = %v, value = %v", client.updateMasks, client.project.AllowLastPlanEditorApproval)
	}
	if got := state.Attributes[projectApprovalField]; got != "true" {
		t.Fatalf("omitted approval state = %q, want server value true", got)
	}
}

func TestProjectLastPlanEditorApprovalRead(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			client := &projectApprovalClient{project: &v1pb.Project{
				Name:                        "projects/project-approval",
				Title:                       "Approval project",
				AllowLastPlanEditorApproval: value,
			}}
			for _, resource := range []*schema.Resource{resourceProjct(), dataSourceProject()} {
				data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{"resource_id": "project-approval"})
				data.SetId(client.project.Name)
				if diags := resource.ReadWithoutTimeout(context.Background(), data, client); diags.HasError() {
					t.Fatalf("read failed: %v", diags)
				}
				if got := data.Get(projectApprovalField); got != value {
					t.Errorf("read approval = %v, want %v", got, value)
				}
			}
			list := dataSourceProjectList()
			data := schema.TestResourceDataRaw(t, list.Schema, nil)
			if diags := list.ReadWithoutTimeout(context.Background(), data, client); diags.HasError() {
				t.Fatalf("list read failed: %v", diags)
			}
			projects := data.Get("projects").([]interface{})
			if len(projects) != 1 {
				t.Fatalf("listed %d projects, want 1", len(projects))
			}
			if got := projects[0].(map[string]interface{})[projectApprovalField]; got != value {
				t.Errorf("listed approval = %v, want %v", got, value)
			}
		})
	}
}

func applyProjectApprovalConfig(t *testing.T, state *terraform.InstanceState, client *projectApprovalClient, title string, approval *bool) *terraform.InstanceState {
	t.Helper()
	resource := resourceProjct()
	if resource.Schema[projectApprovalField] == nil {
		t.Fatal("project resource is missing allow_last_plan_editor_approval")
	}
	config := map[string]interface{}{
		"resource_id": "project-approval",
		"title":       title,
	}
	attrs := map[string]cty.Value{}
	for name, typ := range resource.CoreConfigSchema().ImpliedType().AttributeTypes() {
		attrs[name] = cty.NullVal(typ)
	}
	attrs["resource_id"] = cty.StringVal("project-approval")
	attrs["title"] = cty.StringVal(title)
	if approval != nil {
		config[projectApprovalField] = *approval
		attrs[projectApprovalField] = cty.BoolVal(*approval)
	}
	diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), client)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diff == nil {
		t.Fatal("expected a project change")
	}
	diff.RawConfig = cty.ObjectVal(attrs)
	result, diags := resource.Apply(context.Background(), state, diff, client)
	if diags.HasError() {
		t.Fatalf("apply failed: %v", diags)
	}
	return result
}

type projectApprovalClient struct {
	api.Client
	project     *v1pb.Project
	created     *v1pb.Project
	updateMasks []string
}

func (c *projectApprovalClient) GetProject(context.Context, string) (*v1pb.Project, error) {
	if c.project == nil {
		return nil, errors.New("project not found")
	}
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectApprovalClient) CreateProject(_ context.Context, _ string, project *v1pb.Project) (*v1pb.Project, error) {
	c.created = proto.Clone(project).(*v1pb.Project)
	c.project = proto.Clone(project).(*v1pb.Project)
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectApprovalClient) UpdateProject(_ context.Context, project *v1pb.Project, masks []string) (*v1pb.Project, error) {
	c.updateMasks = append(c.updateMasks, masks...)
	if slices.Contains(masks, projectApprovalField) {
		c.project.AllowLastPlanEditorApproval = project.AllowLastPlanEditorApproval
	}
	if slices.Contains(masks, "title") {
		c.project.Title = project.Title
	}
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectApprovalClient) ListProject(context.Context, *api.ProjectFilter) ([]*v1pb.Project, error) {
	return []*v1pb.Project{proto.Clone(c.project).(*v1pb.Project)}, nil
}

func (*projectApprovalClient) ListDatabase(context.Context, string, *api.DatabaseFilter, bool) ([]*v1pb.Database, error) {
	return nil, nil
}
