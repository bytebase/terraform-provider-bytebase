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

func TestAccServiceAccount(t *testing.T) {
	identifier := "test_sa"
	resourceName := fmt.Sprintf("bytebase_service_account.%s", identifier)

	parent := "workspaces/-"
	serviceAccountID := "test-sa"
	title := "Test Service Account"
	titleUpdated := "Updated Service Account"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			// resource create
			{
				Config: testAccCheckServiceAccountResourceConfig(identifier, parent, serviceAccountID, title),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "parent", parent),
					resource.TestCheckResourceAttr(resourceName, "service_account_id", serviceAccountID),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@service.bytebase.com", serviceAccountID)),
					resource.TestCheckResourceAttr(resourceName, "state", v1pb.State_ACTIVE.String()),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
			// resource update title
			{
				Config: testAccCheckServiceAccountResourceConfig(identifier, parent, serviceAccountID, titleUpdated),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", titleUpdated),
					resource.TestCheckResourceAttr(resourceName, "email", fmt.Sprintf("%s@service.bytebase.com", serviceAccountID)),
					resource.TestCheckResourceAttr(resourceName, "state", v1pb.State_ACTIVE.String()),
				),
			},
		},
	})
}

func TestAccServiceAccount_InvalidInput(t *testing.T) {
	identifier := "invalid_sa"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			// Empty title
			{
				Config:      testAccCheckServiceAccountResourceConfig(identifier, "workspaces/-", "test-sa", ""),
				ExpectError: regexp.MustCompile(`expected "title" to not be an empty string`),
			},
			// Empty service_account_id
			{
				Config:      testAccCheckServiceAccountResourceConfig(identifier, "workspaces/-", "", "Test SA"),
				ExpectError: regexp.MustCompile(`expected "service_account_id" to not be an empty string`),
			},
			// Invalid parent
			{
				Config:      testAccCheckServiceAccountResourceConfig(identifier, "invalid-parent", "test-sa", "Test SA"),
				ExpectError: regexp.MustCompile(`Resource id not match`),
			},
		},
	})
}

func TestAccServiceAccount_DataSource(t *testing.T) {
	resourceIdentifier := "test_sa_ds"
	dataSourceIdentifier := "test_sa_ds_read"
	resourceName := fmt.Sprintf("bytebase_service_account.%s", resourceIdentifier)
	dataSourceName := fmt.Sprintf("data.bytebase_service_account.%s", dataSourceIdentifier)

	parent := "workspaces/-"
	serviceAccountID := "test-sa-ds"
	title := "Test SA Data Source"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServiceAccountDataSourceConfig(resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					internal.TestCheckResourceExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "title", title),
					resource.TestCheckResourceAttr(dataSourceName, "email", fmt.Sprintf("%s@service.bytebase.com", serviceAccountID)),
					resource.TestCheckResourceAttr(dataSourceName, "state", v1pb.State_ACTIVE.String()),
				),
			},
		},
	})
}

func TestAccServiceAccount_DataSourceList(t *testing.T) {
	resourceIdentifier := "test_sa_list"
	dataSourceIdentifier := "test_sa_list_read"
	resourceName := fmt.Sprintf("bytebase_service_account.%s", resourceIdentifier)
	dataSourceName := fmt.Sprintf("data.bytebase_service_account_list.%s", dataSourceIdentifier)

	parent := "workspaces/-"
	serviceAccountID := "test-sa-list"
	title := "Test SA List"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServiceAccountDataSourceListConfig(resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier),
				Check: resource.ComposeTestCheckFunc(
					internal.TestCheckResourceExists(resourceName),
					internal.TestCheckResourceExists(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckServiceAccountResourceConfig(identifier, parent, serviceAccountID, title string) string {
	return fmt.Sprintf(`
resource "bytebase_service_account" "%s" {
	parent             = "%s"
	service_account_id = "%s"
	title              = "%s"
}
`, identifier, parent, serviceAccountID, title)
}

func testAccCheckServiceAccountDataSourceConfig(resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier string) string {
	return fmt.Sprintf(`
resource "bytebase_service_account" "%s" {
	parent             = "%s"
	service_account_id = "%s"
	title              = "%s"
}

data "bytebase_service_account" "%s" {
	name = bytebase_service_account.%s.name
}
`, resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier, resourceIdentifier)
}

func testAccCheckServiceAccountDataSourceListConfig(resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier string) string {
	return fmt.Sprintf(`
resource "bytebase_service_account" "%s" {
	parent             = "%s"
	service_account_id = "%s"
	title              = "%s"
}

data "bytebase_service_account_list" "%s" {
	parent = "%s"
	depends_on = [
		bytebase_service_account.%s
	]
}
`, resourceIdentifier, parent, serviceAccountID, title, dataSourceIdentifier, parent, resourceIdentifier)
}

func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	c, ok := testAccProvider.Meta().(api.Client)
	if !ok {
		return errors.Errorf("cannot get the api client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bytebase_service_account" {
			continue
		}

		if err := c.DeleteServiceAccount(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
	}

	return nil
}
