package provider

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

func TestAccAISettingSQLReviewToggleOnly(t *testing.T) {
	const address = "bytebase_setting.review"
	config := func(enabled bool) string {
		return fmt.Sprintf(`
resource "bytebase_setting" "review" {
  name = "settings/AI"
  ai {
    sql_review_enabled = %t
  }
}

data "bytebase_setting" "review" {
  name       = "settings/AI"
  depends_on = [bytebase_setting.review]
}
`, enabled)
	}
	client := &saasAISettingTestClient{setting: aiSettingResponse()}
	client.setting.GetValue().GetAi().SqlReviewEnabled = false
	p := NewProvider()
	p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, nil
	}

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"bytebase": p},
		Steps: []resource.TestStep{
			{
				Config: config(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "ai.0.sql_review_enabled", "true"),
					resource.TestCheckResourceAttr(address, "ai.0.enabled", "true"),
					resource.TestCheckResourceAttr(address, "ai.0.model", "model-a"),
					resource.TestCheckResourceAttr("data.bytebase_setting.review", "ai.0.sql_review_enabled", "true"),
				),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "ai.0.sql_review_enabled", "false"),
					resource.TestCheckResourceAttr(address, "ai.0.enabled", "true"),
					resource.TestCheckResourceAttr(address, "ai.0.model", "model-a"),
					resource.TestCheckResourceAttr("data.bytebase_setting.review", "ai.0.sql_review_enabled", "false"),
				),
			},
		},
	})
}

type saasAISettingTestClient struct {
	api.Client
	mu      sync.Mutex
	setting *v1pb.Setting
}

func (c *saasAISettingTestClient) GetSetting(_ context.Context, name string) (*v1pb.Setting, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if name != c.setting.GetName() {
		return nil, errors.Errorf("setting %q does not exist", name)
	}
	return proto.CloneOf(c.setting), nil
}

func (c *saasAISettingTestClient) UpsertSetting(_ context.Context, setting *v1pb.Setting, masks []string) (*v1pb.Setting, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ai := setting.GetValue().GetAi()
	if !slices.Equal(masks, []string{"value.ai.sql_review_enabled"}) || !proto.Equal(ai, &v1pb.AISetting{SqlReviewEnabled: ai.GetSqlReviewEnabled()}) {
		return nil, errors.Errorf("SaaS AI settings only allow writing the SQL review toggle; got masks %v", masks)
	}
	c.setting.GetValue().GetAi().SqlReviewEnabled = ai.GetSqlReviewEnabled()
	return proto.CloneOf(c.setting), nil
}

func TestAccReviewAIPolicyDeferredContent(t *testing.T) {
	const address = "bytebase_policy.review"
	p := NewProvider()
	client := &reviewPolicyClient{}
	p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, nil
	}

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"bytebase": p},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "config" {
  input = {
    parent  = "projects/review-project"
    content = "Require a rollback plan."
  }
}

resource "bytebase_policy" "review" {
  type   = "REVIEW_AI"
  parent = terraform_data.config.output.parent
  review_ai_policy {
    content = terraform_data.config.output.content
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "parent", "projects/review-project"),
					resource.TestCheckResourceAttr(address, "review_ai_policy.0.content", "Require a rollback plan."),
				),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
