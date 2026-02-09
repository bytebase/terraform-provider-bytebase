package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/bytebase/terraform-provider-bytebase/provider/internal"
)

func TestAccWorkloadIdentity(t *testing.T) {
	identifier := "test_wi"
	resourceName := fmt.Sprintf("bytebase_workload_identity.%s", identifier)

	parent := "workspaces/-"
	workloadIdentityID := "test-wi"
	title := "Test Workload Identity"
	titleUpdated := "Updated Workload Identity"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkloadIdentityDestroy,
		Steps: []resource.TestStep{
			// resource create with config
			{
				Config: testAccCheckWorkloadIdentityResourceConfig(identifier, parent, workloadIdentityID, title, v1pb.WorkloadIdentityConfig_GITHUB.String(), "repo:owner/repo:ref:refs/heads/main"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "parent", parent),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_id", workloadIdentityID),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@workload.bytebase.com", workloadIdentityID)),
					resource.TestCheckResourceAttr(resourceName, "state", v1pb.State_ACTIVE.String()),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.0.provider_type", v1pb.WorkloadIdentityConfig_GITHUB.String()),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.0.subject_pattern", "repo:owner/repo:ref:refs/heads/main"),
				),
			},
			// resource update title and config
			{
				Config: testAccCheckWorkloadIdentityResourceConfig(identifier, parent, workloadIdentityID, titleUpdated, v1pb.WorkloadIdentityConfig_GITLAB.String(), "project_path:group/project:ref_type:branch:ref:main"),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@workload.bytebase.com", workloadIdentityID)),
					resource.TestCheckResourceAttr(resourceName, "state", v1pb.State_ACTIVE.String()),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.0.provider_type", v1pb.WorkloadIdentityConfig_GITLAB.String()),
					resource.TestCheckResourceAttr(resourceName, "workload_identity_config.0.subject_pattern", "project_path:group/project:ref_type:branch:ref:main"),
				),
			},
		},
	})
}

func TestAccWorkloadIdentity_InvalidInput(t *testing.T) {
	identifier := "invalid_wi"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkloadIdentityDestroy,
		Steps: []resource.TestStep{
			// Empty title
			{
				Config:      testAccCheckWorkloadIdentityResourceConfigSimple(identifier, "workspaces/-", "test-wi", ""),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// Empty workload_identity_id
			{
				Config:      testAccCheckWorkloadIdentityResourceConfigSimple(identifier, "workspaces/-", "", "Test WI"),
				ExpectError: regexp.MustCompile(`expected "workload_identity_id" to not be an empty string`),
			},
			// Invalid parent
			{
				Config:      testAccCheckWorkloadIdentityResourceConfigSimple(identifier, "invalid-parent", "test-wi", "Test WI"),
				ExpectError: regexp.MustCompile(`Resource id not match`),
			},
		},
	})
}

func TestAccWorkloadIdentity_DataSource(t *testing.T) {
	resourceIdentifier := "test_wi_ds"
	dataSourceIdentifier := "test_wi_ds_read"
	resourceName := fmt.Sprintf("bytebase_workload_identity.%s", resourceIdentifier)
	dataSourceName := fmt.Sprintf("data.bytebase_workload_identity.%s", dataSourceIdentifier)

	parent := "workspaces/-"
	workloadIdentityID := "test-wi-ds"
	title := "Test WI Data Source"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkloadIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWorkloadIdentityDataSourceConfig(resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					internal.TestCheckResourceExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "title", title),
					resource.TestCheckResourceAttr(dataSourceName, "email", fmt.Sprintf("%s@workload.bytebase.com", workloadIdentityID)),
					resource.TestCheckResourceAttr(dataSourceName, "state", v1pb.State_ACTIVE.String()),
				),
			},
		},
	})
}

func TestAccWorkloadIdentity_DataSourceList(t *testing.T) {
	resourceIdentifier := "test_wi_list"
	dataSourceIdentifier := "test_wi_list_read"
	resourceName := fmt.Sprintf("bytebase_workload_identity.%s", resourceIdentifier)
	dataSourceName := fmt.Sprintf("data.bytebase_workload_identity_list.%s", dataSourceIdentifier)

	parent := "workspaces/-"
	workloadIdentityID := "test-wi-list"
	title := "Test WI List"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkloadIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWorkloadIdentityDataSourceListConfig(resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					internal.TestCheckResourceExists(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckWorkloadIdentityResourceConfig(identifier, parent, workloadIdentityID, title, providerType, subjectPattern string) string {
	return fmt.Sprintf(`
resource "bytebase_workload_identity" "%s" {
	parent               = "%s"
	workload_identity_id = "%s"
	title                = "%s"

	workload_identity_config {
		provider_type   = "%s"
		subject_pattern = "%s"
	}
}
`, identifier, parent, workloadIdentityID, title, providerType, subjectPattern)
}

func testAccCheckWorkloadIdentityResourceConfigSimple(identifier, parent, workloadIdentityID, title string) string {
	return fmt.Sprintf(`
resource "bytebase_workload_identity" "%s" {
	parent               = "%s"
	workload_identity_id = "%s"
	title                = "%s"
}
`, identifier, parent, workloadIdentityID, title)
}

func testAccCheckWorkloadIdentityDataSourceConfig(resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier string) string {
	return fmt.Sprintf(`
resource "bytebase_workload_identity" "%s" {
	parent               = "%s"
	workload_identity_id = "%s"
	title                = "%s"
}

data "bytebase_workload_identity" "%s" {
	name = bytebase_workload_identity.%s.name
}
`, resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier, resourceIdentifier)
}

func testAccCheckWorkloadIdentityDataSourceListConfig(resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier string) string {
	return fmt.Sprintf(`
resource "bytebase_workload_identity" "%s" {
	parent               = "%s"
	workload_identity_id = "%s"
	title                = "%s"
}

data "bytebase_workload_identity_list" "%s" {
	parent = "%s"
	depends_on = [
		bytebase_workload_identity.%s
	]
}
`, resourceIdentifier, parent, workloadIdentityID, title, dataSourceIdentifier, parent, resourceIdentifier)
}

func testAccCheckWorkloadIdentityDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_workload_identity" {
			continue
		}

		if err := c.DeleteWorkloadIdentity(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
