package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const (
	projectExportURL    = "https://acme.sharepoint.com/sites/Finance/Shared%20Documents/Exports"
	projectExportCopied = " HTTPS://ACME.sharepoint.com/:f:/r/sites/Finance/Shared%20Documents//Exports/?csf=1&web=1#fragment "
	projectExportAttr   = "export_setting.0.sharepoint.0.url"
)

func TestProjectSharePointDestinationCreateAndUpdate(t *testing.T) {
	client := &projectExportClient{}
	state, diags := applyProjectExportConfig(t, nil, client, "Project", true, cty.StringVal(projectExportCopied))
	if diags.HasError() {
		t.Fatalf("create failed: %v", diags)
	}
	if client.createHadExport || !slices.Equal(client.updateMasks, []string{"export_setting"}) {
		t.Fatalf("expected one masked update after create; create had destination = %v, update mask = %v", client.createHadExport, client.updateMasks)
	}
	if got := state.Attributes[projectExportAttr]; got != projectExportURL {
		t.Fatalf("created destination = %q, want canonical %q", got, projectExportURL)
	}
	assertProjectExportNoDiff(t, state, client, true, cty.StringVal(projectExportCopied))

	libraryView := "https://ACME.sharepoint.com/sites/Finance/Shared%20Documents/Forms/AllItems.aspx?id=%2Fsites%2FFinance%2FShared%20Documents%2FQuarterly"
	canonical := "https://acme.sharepoint.com/sites/Finance/Shared%20Documents/Quarterly"
	client.updateMasks = nil
	state, diags = applyProjectExportConfig(t, state, client, "Project", true, cty.StringVal(libraryView))
	if diags.HasError() {
		t.Fatalf("update failed: %v", diags)
	}
	if got := client.project.GetExportSetting().GetSharepoint().GetUrl(); !slices.Equal(client.updateMasks, []string{"export_setting"}) || got != canonical {
		t.Fatalf("updating destination got mask %v and URL %q, want export_setting and %q", client.updateMasks, got, canonical)
	}
	assertProjectExportNoDiff(t, state, client, true, cty.StringVal(libraryView))

	client.updateMasks = nil
	state, diags = applyProjectExportConfig(t, state, client, "Project", true, cty.NullVal(cty.String))
	if diags.HasError() {
		t.Fatalf("clear failed: %v", diags)
	}
	if got := client.project.GetExportSetting().GetSharepoint().GetUrl(); !slices.Equal(client.updateMasks, []string{"export_setting"}) || got != "" {
		t.Fatalf("clearing destination got mask %v and URL %q, want export_setting and empty URL", client.updateMasks, got)
	}
	assertProjectExportNoDiff(t, state, client, true, cty.NullVal(cty.String))
}

func TestProjectSharePointDestinationOmissionPreservesExisting(t *testing.T) {
	client := &projectExportClient{project: existingProjectWithExport()}
	state, diags := applyProjectExportConfig(t, nil, client, "Adopted project", false, cty.NullVal(cty.String))
	if diags.HasError() {
		t.Fatalf("adopting existing project failed: %v", diags)
	}
	if got := state.Attributes[projectExportAttr]; !slices.Equal(client.updateMasks, []string{"title"}) || got != projectExportURL {
		t.Fatalf("adopting existing project got mask %v and destination %q, want title and %q", client.updateMasks, got, projectExportURL)
	}

	state, diags = applyProjectExportConfig(t, state, client, "Project", true, cty.StringVal(projectExportURL))
	if diags.HasError() {
		t.Fatalf("setting retained destination failed: %v", diags)
	}

	client.updateMasks = nil
	state, diags = applyProjectExportConfig(t, state, client, "Updated project", false, cty.NullVal(cty.String))
	if diags.HasError() {
		t.Fatalf("updating project with omission failed: %v", diags)
	}
	if got := state.Attributes[projectExportAttr]; !slices.Equal(client.updateMasks, []string{"title"}) || got != projectExportURL {
		t.Fatalf("removing the export block while updating title got mask %v and destination %q, want title and %q", client.updateMasks, got, projectExportURL)
	}
}

func TestProjectSharePointDestinationPostCreateErrorRetainsID(t *testing.T) {
	client := &projectExportClient{updateErr: errors.New("workspace credentials unavailable")}
	state, diags := applyProjectExportConfig(t, nil, client, "Project", true, cty.StringVal(projectExportURL))
	if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "workspace credentials unavailable") {
		t.Fatalf("apply should report failed destination update, got %v", diags)
	}
	if state == nil || state.ID != "projects/project-export" {
		t.Fatalf("apply should have retained project ID after failed destination update; got %v", state)
	}
}

func TestProjectSharePointDestinationFailedAdoptionDoesNotClaimProject(t *testing.T) {
	client := &projectExportClient{
		project:   existingProjectWithExport(),
		updateErr: errors.New("workspace credentials unavailable"),
	}
	state, diags := applyProjectExportConfig(t, nil, client, "Project", true, cty.StringVal("https://acme.sharepoint.com/OtherExports"))
	if !diags.HasError() {
		t.Fatalf("failed adoption must report the update error, got %v", diags)
	}
	if state != nil && state.ID != "" {
		t.Fatalf("failed adoption must not claim the existing project: %v", state)
	}
}

func TestProjectSharePointDestinationRead(t *testing.T) {
	client := &projectExportClient{project: existingProjectWithExport()}
	for _, test := range []struct {
		name     string
		resource *schema.Resource
		field    string
	}{
		{"imported resource", resourceProjct(), projectExportAttr},
		{"data source", dataSourceProject(), projectExportAttr},
		{"list data source", dataSourceProjectList(), "projects.0." + projectExportAttr},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := map[string]interface{}{}
			if test.resource.Schema["resource_id"] != nil {
				config["resource_id"] = "project-export"
			}
			data := schema.TestResourceDataRaw(t, test.resource.Schema, config)
			data.SetId(client.project.Name)
			if diags := test.resource.ReadWithoutTimeout(context.Background(), data, client); diags.HasError() {
				t.Fatalf("read failed: %v", diags)
			}
			if got := data.State().Attributes[test.field]; got != projectExportURL {
				t.Errorf("read destination = %q, want %q", got, projectExportURL)
			}
		})
	}
}

func TestProjectSharePointDestinationPlanValidation(t *testing.T) {
	for _, url := range []string{
		"",
		"http://acme.sharepoint.com/sites/Finance/Exports",
		"https://acme.sharepoint.com:443/sites/Finance/Exports",
		"https://user:password@acme.sharepoint.com/sites/Finance/Exports",
		"https://acme.sharepoint.com.evil.test/sites/Finance/Exports",
		"https://sharepoint.com/sites/Finance/Exports",
		"https://other.acme.sharepoint.com/sites/Finance/Exports",
		"https://acme-my.sharepoint.com/personal/user/Documents",
		"https://acme.sharepoint.com/:f:/s/Finance/link-token",
		"https://acme.sharepoint.com/sites/Finance/",
		"https://acme.sharepoint.com/sites/Finance/%2e%2e/Exports",
		"https://acme.sharepoint.com/sites/Finance/_layouts/15/list.aspx",
	} {
		t.Run(url, func(t *testing.T) {
			r := resourceProjct()
			// Check that the error comes from destination validation, not an unknown block.
			if r.Schema["export_setting"] == nil {
				t.Fatal("project schema is missing export_setting")
			}
			config := newProjectExportConfig(t, r, "Project", true, cty.StringVal(url))
			diags := r.Validate(config)
			if !diags.HasError() || !strings.Contains(fmt.Sprint(diags), "SharePoint") {
				t.Fatalf("plan should reject invalid SharePoint URL; got %v", diags)
			}
		})
	}
	r := resourceProjct()
	config := newProjectExportConfig(t, r, "Project", true, cty.UnknownVal(cty.String))
	if diags := r.Validate(config); diags.HasError() {
		t.Fatalf("planning should defer validation of a URL unknown until apply, got %v", diags)
	}
	if _, err := r.Diff(context.Background(), nil, config, nil); err != nil {
		t.Fatalf("diff with unknown URL failed: %v", err)
	}
}

func TestAccProjectSharePointDestinationConverges(t *testing.T) {
	const resourceName = "bytebase_project.export"
	c := &projectExportClient{}
	p := NewProvider()
	p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return c, nil
	}
	config := func(title, exportBlock string) string {
		return fmt.Sprintf(`
resource "bytebase_project" "export" {
  resource_id = "project-export"
  title = %q
  %s
}
`, title, exportBlock)
	}
	withDestination := config("SharePoint exports", `
export_setting {
  sharepoint {
    url = terraform_data.copied_url.output
  }
}
`) + fmt.Sprintf(`
resource "terraform_data" "copied_url" {
  input = %q
}
`, projectExportCopied)
	omitted := config("Omit destination", "")
	cleared := config("Omit destination", "export_setting {}")

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"bytebase": p},
		Steps: []resource.TestStep{
			{
				Config: withDestination,
				Check:  resource.TestCheckResourceAttr(resourceName, projectExportAttr, projectExportURL),
			},
			{
				Config:   withDestination,
				PlanOnly: true,
			},
			{
				Config: omitted,
				Check:  resource.TestCheckResourceAttr(resourceName, projectExportAttr, projectExportURL),
			},
			{
				Config:   omitted,
				PlanOnly: true,
			},
			{
				Config: cleared,
				Check:  resource.TestCheckNoResourceAttr(resourceName, projectExportAttr),
			},
			{
				Config:   cleared,
				PlanOnly: true,
			},
		},
	})
}

func assertProjectExportNoDiff(t *testing.T, state *terraform.InstanceState, client *projectExportClient, configured bool, url cty.Value) {
	t.Helper()
	r := resourceProjct()
	data, diags := r.RefreshWithoutUpgrade(context.Background(), state, client)
	if diags.HasError() {
		t.Fatalf("refresh failed: %v", diags)
	}
	config := newProjectExportConfig(t, r, "Project", configured, url)
	diff, err := r.Diff(context.Background(), data, config, client)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diff != nil {
		for name, change := range diff.Attributes {
			if strings.HasPrefix(name, "export_setting.") {
				t.Errorf("export destination did not converge: %s = %#v", name, change)
			}
		}
	}
}

func newProjectExportConfig(t *testing.T, r *schema.Resource, title string, configured bool, url cty.Value) *terraform.ResourceConfig {
	t.Helper()
	types := r.CoreConfigSchema().ImpliedType().AttributeTypes()
	exportType, ok := types["export_setting"]
	if !ok {
		t.Fatal("project schema is missing export_setting")
	}
	attrs := map[string]cty.Value{}
	for name, typ := range types {
		attrs[name] = cty.NullVal(typ)
	}
	attrs["resource_id"] = cty.StringVal("project-export")
	attrs["title"] = cty.StringVal(title)
	attrs["export_setting"] = cty.ListValEmpty(exportType.ElementType())
	if configured {
		sharepointType := exportType.ElementType().AttributeType("sharepoint")
		sharepoint := cty.ListValEmpty(sharepointType.ElementType())
		if !url.IsNull() {
			sharepoint = cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"url": url})})
		}
		attrs["export_setting"] = cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"sharepoint": sharepoint})})
	}
	value := cty.ObjectVal(attrs)
	config := terraform.NewResourceConfigShimmed(value, r.CoreConfigSchema())
	config.CtyValue = value
	return config
}

func applyProjectExportConfig(t *testing.T, state *terraform.InstanceState, client *projectExportClient, title string, configured bool, url cty.Value) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()
	r := resourceProjct()
	config := newProjectExportConfig(t, r, title, configured, url)
	diff, err := r.Diff(context.Background(), state, config, client)
	if err != nil {
		t.Fatalf("diff failed: %v", err)
	}
	if diff == nil {
		return state, nil
	}
	diff.RawConfig = config.CtyValue
	return r.Apply(context.Background(), state, diff, client)
}

func existingProjectWithExport() *v1pb.Project {
	return &v1pb.Project{
		Name:  "projects/project-export",
		Title: "Project",
		State: v1pb.State_ACTIVE,
		ExportSetting: &v1pb.Project_ExportSetting{
			Sharepoint: &v1pb.SharePointDestination{Url: projectExportURL},
		},
	}
}

type projectExportClient struct {
	api.Client
	project         *v1pb.Project
	createHadExport bool
	updateMasks     []string
	updateErr       error
}

func (c *projectExportClient) GetProject(context.Context, string) (*v1pb.Project, error) {
	if c.project == nil {
		return nil, errors.New("project not found")
	}
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectExportClient) CreateProject(_ context.Context, _ string, project *v1pb.Project) (*v1pb.Project, error) {
	c.createHadExport = project.ExportSetting != nil
	c.project = proto.Clone(project).(*v1pb.Project)
	c.project.ExportSetting = nil // Bytebase persists this only through UpdateProject.
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectExportClient) UpdateProject(_ context.Context, project *v1pb.Project, masks []string) (*v1pb.Project, error) {
	c.updateMasks = append(c.updateMasks, masks...)
	if c.updateErr != nil {
		return nil, c.updateErr
	}
	if slices.Contains(masks, "export_setting") {
		if project.GetExportSetting().GetSharepoint().GetUrl() == "" {
			c.project.ExportSetting = nil
		} else {
			c.project.ExportSetting = proto.Clone(project.ExportSetting).(*v1pb.Project_ExportSetting)
		}
	}
	if slices.Contains(masks, "title") {
		c.project.Title = project.Title
	}
	return proto.Clone(c.project).(*v1pb.Project), nil
}

func (c *projectExportClient) ListProject(context.Context, *api.ProjectFilter) ([]*v1pb.Project, error) {
	return []*v1pb.Project{proto.Clone(c.project).(*v1pb.Project)}, nil
}

func (c *projectExportClient) DeleteProject(context.Context, string) error {
	c.project = nil
	return nil
}

func (*projectExportClient) ListDatabase(context.Context, string, *api.DatabaseFilter, bool) ([]*v1pb.Database, error) {
	return nil, nil
}
