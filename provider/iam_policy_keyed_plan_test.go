package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const keyedIAMAddress = "bytebase_iam_policy.test"

// These tests run the real Terraform CLI and provider RPC server. Only the API
// client is replaced, so the plan assertions cover Terraform's nested map diff.
func TestIAMPolicyKeyedTerraformPlan(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	config := func(members, schemaName string) string {
		return keyedIAMConfig("projects/test-project", fmt.Sprintf(`
  readers = {
    role = "roles/reader"
    members = [%s]
    condition = {
      database = "instances/test/databases/test"
      schema = %q
      tables = ["orders", "users"]
    }
  }
  writers = {
    role = "roles/writer"
    members = ["user:writer@example.com"]
  }
`, members, schemaName))
	}
	alice := `"user:alice@example.com"`
	bob := `"user:bob@example.com"`
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{
			{Config: config(alice, "public")},
			{
				Config:  config(alice+", "+bob, "public"),
				LogPlan: true,
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMBindingChange("readers", "members",
					[]string{"user:alice@example.com"}, []string{"user:alice@example.com", "user:bob@example.com"})),
			},
			{
				Config:           config(bob+", "+alice, "public"),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "no-op")),
			},
			{
				Config: config(bob, "public"),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMBindingChange("readers", "members",
					[]string{"user:alice@example.com", "user:bob@example.com"}, []string{"user:bob@example.com"})),
			},
			{
				Config:           config(bob, "analytics"),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMBindingChange("readers", "condition.schema", "public", "analytics")),
			},
			{
				Config: keyedIAMConfig("projects/test-project", `
  writers = {
    members = ["user:writer@example.com"]
    role = "roles/writer"
  }
  readers = {
    members = ["user:bob@example.com"]
    role = "roles/reader"
    condition = {
      tables = ["users", "orders"]
      schema = "analytics"
      database = "instances/test/databases/test"
    }
  }
`),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "no-op")),
			},
			{
				Config: strings.Replace(config(bob, "analytics"), "readers = {", "renamed = {", 1),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMPlanCheck(func(change *tfjson.Change) error {
					if !change.Actions.Update() {
						return errors.Errorf("key rename must update in place, got %v", change.Actions)
					}
					before, after := keyedIAMPlanBindings(change.Before), keyedIAMPlanBindings(change.After)
					if len(before) != 2 || len(after) != 2 || before["renamed"] != nil || after["readers"] != nil ||
						!reflect.DeepEqual(before["readers"], after["renamed"]) || !reflect.DeepEqual(before["writers"], after["writers"]) {
						return errors.Errorf("key rename changed grants: before=%v after=%v", before, after)
					}
					return nil
				})),
			},
			{
				Config:           strings.Replace(config(bob, "analytics"), "readers = {", "renamed = {", 1),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "no-op")),
			},
		},
	})
	client.assertNoConflicts(t)
}

func TestIAMPolicyKeyedTerraformLegacyUpgrade(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	client.checkEtag = false // The old SDK resource did not send etags.
	var legacyExpression string
	legacy := `
resource "bytebase_iam_policy" "test" {
  parent = "projects/test-project"
  iam_policy {
    binding {
      role = "roles/reader"
      members = ["user:alice@example.com", "user:bob@example.com"]
      condition {
        database = "instances/test/databases/test"
        schema = "public"
        tables = ["orders", "users"]
      }
    }
  }
}
`
	keyedIAMRun(t, keyedIAMTestCase{
		Steps: []keyedIAMTestStep{
			{
				ProtoV6ProviderFactories: keyedIAMFactories(client, true),
				Config:                   legacy,
				Check: func(_ *tfjson.State) error {
					client.mu.Lock()
					defer client.mu.Unlock()
					legacyExpression = client.policy.Bindings[0].GetCondition().GetExpression()
					return nil
				},
			},
			{
				PreConfig: func() {
					client.mu.Lock()
					client.checkEtag = true
					client.mu.Unlock()
				},
				ProtoV6ProviderFactories: keyedIAMFactories(client, false),
				Config:                   legacy,
				ConfigPlanChecks:         keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "no-op")),
			},
			{
				ProtoV6ProviderFactories: keyedIAMFactories(client, false),
				Config: keyedIAMConfig("projects/test-project", `
  readers = {
    role = "roles/reader"
    members = ["user:bob@example.com", "user:alice@example.com"]
    condition = {
      database = "instances/test/databases/test"
      schema = "public"
      tables = ["users", "orders"]
      environment_ids = []
    }
  }
`),
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "update")),
				Check: func(state *tfjson.State) error {
					if err := keyedIAMResourceAttr(keyedIAMAddress, "bindings.readers.members.#", "2")(state); err != nil {
						return err
					}
					client.mu.Lock()
					defer client.mu.Unlock()
					current := client.policy.Bindings[0].GetCondition().GetExpression()
					if legacyExpression == "" || normalizeIAMConditionExpression(legacyExpression) != normalizeIAMConditionExpression(current) {
						return errors.Errorf("migration changed condition semantics: before=%q after=%q", legacyExpression, current)
					}
					return nil
				},
			},
		},
	})
	client.assertNoConflicts(t)
}

func TestIAMPolicyKeyedTerraformDrift(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	config := keyedIAMConfig("projects/test-project", `
  readers = {
    role = "roles/reader"
    members = ["user:alice@example.com"]
  }
`)
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{
			{Config: config},
			{
				PreConfig: func() {
					client.mu.Lock()
					defer client.mu.Unlock()
					client.policy.Bindings[0].Members = append(client.policy.Bindings[0].Members, "user:remote@example.com")
					client.policy.Bindings = append(client.policy.Bindings, &v1pb.Binding{
						Role: "roles/remote", Members: []string{"user:added@example.com"},
					})
					client.nextEtag()
				},
				Config: config,
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMPlanCheck(func(change *tfjson.Change) error {
					if !change.Actions.Update() {
						return errors.Errorf("drift should update in place, got %v", change.Actions)
					}
					before, after := keyedIAMPlanBindings(change.Before), keyedIAMPlanBindings(change.After)
					if len(before) != 2 || len(after) != 1 {
						return errors.Errorf("remote binding must remain visible as drift: before=%v after=%v", before, after)
					}
					readers, ok := before["readers"].(map[string]interface{})
					if !ok || !reflect.DeepEqual(keyedIAMSortedMembers(readers["members"]), []string{"user:alice@example.com", "user:remote@example.com"}) {
						return errors.Errorf("remote member drift lost the readers key or grant: %v", before)
					}
					for key, value := range before {
						if key != "readers" && value.(map[string]interface{})["role"] != "roles/remote" {
							return errors.Errorf("unexpected generated binding: %s=%v", key, value)
						}
					}
					return nil
				})),
				Check: func(_ *tfjson.State) error {
					client.mu.Lock()
					defer client.mu.Unlock()
					if len(client.policy.Bindings) != 1 || !reflect.DeepEqual(client.policy.Bindings[0].Members, []string{"user:alice@example.com"}) {
						return errors.Errorf("authoritative apply did not correct drift: %v", client.policy)
					}
					return nil
				},
			},
		},
	})
	client.assertNoConflicts(t)
}

func TestIAMPolicyKeyedTerraformWorkspaceImport(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{
			{
				Config: keyedIAMConfig("", `
  owners = {
    role = "roles/owner"
    members = ["user:alice@example.com"]
  }
`),
				Check: keyedIAMResourceAttr(keyedIAMAddress, "parent", "workspaces/test-workspace"),
			},
			{
				ResourceName:  keyedIAMAddress,
				ImportState:   true,
				ImportStateID: "workspaces/test-workspace",
				ImportStateCheck: func(state *tfjson.State) error {
					values := keyedIAMStateValues(state)
					if values["id"] != "workspaces/test-workspace" || values["parent"] != "workspaces/test-workspace" {
						return errors.Errorf("unexpected imported policy: %v", values)
					}
					bindings, ok := values["bindings"].(map[string]interface{})
					if !ok || len(bindings) != 1 {
						return errors.Errorf("import should expose keyed bindings: %v", values)
					}
					for _, value := range bindings {
						binding := value.(map[string]interface{})
						if binding["role"] == "roles/owner" && reflect.DeepEqual(keyedIAMSortedMembers(binding["members"]), []string{"user:alice@example.com"}) {
							return nil
						}
					}
					return errors.Errorf("imported binding is missing grants: %v", bindings)
				},
			},
		},
	})
	client.assertNoConflicts(t)
}

func TestIAMPolicyKeyedTerraformDuplicateConfig(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{{
			Config: keyedIAMConfig("projects/test-project", `
  first = {
    role = "roles/reader"
    members = ["user:alice@example.com"]
  }
  second = {
    role = "roles/reader"
    members = ["user:bob@example.com"]
    condition = {}
  }
`),
			ExpectError: regexp.MustCompile(`(?i)duplicate.*(binding|identity)|same role.*condition`),
		}},
	})
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.writes != 0 {
		t.Fatalf("invalid duplicate configuration performed %d writes", client.writes)
	}
}

func TestIAMPolicyKeyedTerraformUnknownMembers(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	config := keyedIAMConfig("projects/test-project", `
  readers = {
    role = "roles/reader"
    members = ["user:${bytebase_test_value.member.id}@example.com"]
  }
`) + `
resource "bytebase_test_value" "member" {}
`
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{
			{
				Config: config,
				Check:  keyedIAMResourceAttr(keyedIAMAddress, "bindings.readers.members.#", "1"),
			},
			{
				Config:           config,
				ConfigPlanChecks: keyedIAMPreApply(keyedIAMExpectAction(keyedIAMAddress, "no-op")),
			},
		},
	})
	client.assertNoConflicts(t)
}

func TestIAMPolicyKeyedTerraformEtagConflict(t *testing.T) {
	keyedIAMTerraform(t)
	client := newKeyedIAMTestClient()
	config := func(member string) string {
		return keyedIAMConfig("projects/test-project", fmt.Sprintf(`
  readers = {
    role = "roles/reader"
    members = [%q]
  }
`, member))
	}
	keyedIAMRun(t, keyedIAMTestCase{
		ProtoV6ProviderFactories: keyedIAMFactories(client, false),
		Steps: []keyedIAMTestStep{
			{Config: config("user:alice@example.com")},
			{
				PreConfig: func() {
					client.mu.Lock()
					client.conflictNext = true
					client.mu.Unlock()
				},
				Config:      config("user:bob@example.com"),
				ExpectError: regexp.MustCompile(`etag conflict`),
			},
		},
	})
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.conflicts != 1 || client.writes != 2 {
		t.Fatalf("conflict must surface without retry: conflicts=%d writes=%d", client.conflicts, client.writes)
	}
	if !reflect.DeepEqual(client.policy.Bindings[0].Members, []string{"user:alice@example.com"}) {
		t.Fatalf("conflicting write changed grants: %v", client.policy)
	}
}

func keyedIAMTerraform(t *testing.T) {
	t.Helper()
	t.Setenv("TF_LOG_SDK", "ERROR")
	t.Setenv("TF_LOG_SDK_PROTO", "ERROR")
	t.Setenv("TF_LOG_SDK_FRAMEWORK", "ERROR")
	t.Setenv("TF_LOG_SDK_MUX", "ERROR")
	t.Setenv("TF_LOG_SDK_HELPER_SCHEMA", "ERROR")
	if os.Getenv("TF_ACC_TERRAFORM_PATH") != "" {
		return
	}
	path, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("Terraform CLI is not installed")
	}
	t.Setenv("TF_ACC_TERRAFORM_PATH", path)
}

type keyedIAMTestCase struct {
	ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	Steps                    []keyedIAMTestStep
}

type keyedIAMTestStep struct {
	ProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	PreConfig                func()
	Config                   string
	LogPlan                  bool
	ConfigPlanChecks         keyedIAMConfigPlanChecks
	Check                    func(*tfjson.State) error
	ExpectError              *regexp.Regexp
	ResourceName             string
	ImportState              bool
	ImportStateID            string
	ImportStateCheck         func(*tfjson.State) error
}

type keyedIAMConfigPlanChecks struct {
	PreApply []keyedIAMPlanCheck
}

// A small direct CLI harness avoids mixing the SDK and plugin-testing resource
// helpers, which both register the same global test flags.
func keyedIAMRun(t *testing.T, testCase keyedIAMTestCase) {
	t.Helper()
	dir := t.TempDir()
	tf, err := tfexec.NewTerraform(dir, os.Getenv("TF_ACC_TERRAFORM_PATH"))
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.SetDisablePluginTLS(true); err != nil {
		t.Fatal(err)
	}
	tf.SetStdout(io.Discard)
	tf.SetStderr(io.Discard)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	for index, step := range testCase.Steps {
		func() {
			if step.PreConfig != nil {
				step.PreConfig()
			}
			factories := step.ProtoV6ProviderFactories
			if factories == nil {
				factories = testCase.ProtoV6ProviderFactories
			}
			server, err := factories["bytebase"]()
			if err != nil {
				t.Fatalf("step %d: create provider: %v", index+1, err)
			}
			reattach, stop := keyedIAMServe(t, server)
			defer stop()
			if step.Config != "" {
				config := `terraform {
  required_providers {
    bytebase = { source = "bytebase/bytebase" }
  }
}
provider "bytebase" {}
` + step.Config
				if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(config), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := tf.Init(ctx, tfexec.Reattach(reattach)); err != nil {
				t.Fatalf("step %d: terraform init: %v", index+1, err)
			}
			if step.ImportState {
				if err := os.Remove(filepath.Join(dir, "terraform.tfstate")); err != nil {
					t.Fatal(err)
				}
				if err := tf.Import(ctx, step.ResourceName, step.ImportStateID, tfexec.Reattach(reattach)); err != nil {
					t.Fatalf("step %d: terraform import: %v", index+1, err)
				}
			} else {
				_, planErr := tf.Plan(ctx, tfexec.Reattach(reattach), tfexec.Out("test.tfplan"))
				if planErr != nil {
					if step.ExpectError != nil && step.ExpectError.MatchString(planErr.Error()) {
						return
					}
					t.Fatalf("step %d: terraform plan: %v", index+1, planErr)
				}
				plan, err := tf.ShowPlanFile(ctx, "test.tfplan", tfexec.Reattach(reattach))
				if err != nil {
					t.Fatalf("step %d: terraform show -json: %v", index+1, err)
				}
				if step.LogPlan {
					raw, err := tf.ShowPlanFileRaw(ctx, "test.tfplan", tfexec.Reattach(reattach))
					if err != nil {
						t.Fatalf("step %d: human-readable Terraform plan: %v", index+1, err)
					}
					t.Logf("step %d human-readable Terraform plan:\n%s", index+1, raw)
				}
				for _, check := range step.ConfigPlanChecks.PreApply {
					var change *tfjson.Change
					for _, resource := range plan.ResourceChanges {
						if resource.Address == keyedIAMAddress {
							change = resource.Change
							break
						}
					}
					if change == nil {
						t.Fatalf("step %d: IAM resource missing from Terraform plan", index+1)
					}
					if err := check(change); err != nil {
						raw, _ := tf.ShowPlanFileRaw(ctx, "test.tfplan", tfexec.Reattach(reattach))
						t.Fatalf("step %d: %v\n%s", index+1, err, raw)
					}
				}
				applyErr := tf.Apply(ctx, tfexec.Reattach(reattach), tfexec.DirOrPlan("test.tfplan"))
				if step.ExpectError != nil {
					if applyErr == nil || !step.ExpectError.MatchString(applyErr.Error()) {
						t.Fatalf("step %d: expected %s, got %v", index+1, step.ExpectError, applyErr)
					}
					return
				}
				if applyErr != nil {
					t.Fatalf("step %d: terraform apply: %v", index+1, applyErr)
				}
			}
			state, err := tf.Show(ctx, tfexec.Reattach(reattach))
			if err != nil {
				t.Fatalf("step %d: terraform show: %v", index+1, err)
			}
			for _, check := range []func(*tfjson.State) error{step.Check, step.ImportStateCheck} {
				if check != nil {
					if err := check(state); err != nil {
						t.Fatalf("step %d: %v", index+1, err)
					}
				}
			}
		}()
	}
}

func keyedIAMServe(t *testing.T, server tfprotov6.ProviderServer) (tfexec.ReattachInfo, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	configCh, closeCh := make(chan *plugin.ReattachConfig, 1), make(chan struct{})
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- tf6server.Serve("registry.terraform.io/bytebase/bytebase", func() tfprotov6.ProviderServer { return server },
			tf6server.WithDebug(ctx, configCh, closeCh), tf6server.WithGoPluginLogger(hclog.NewNullLogger()), tf6server.WithoutLogStderrOverride())
	}()
	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			select {
			case <-closeCh:
			case <-time.After(5 * time.Second):
				t.Error("provider RPC server did not stop")
			}
		})
	}
	t.Cleanup(stop)
	select {
	case config := <-configCh:
		return tfexec.ReattachInfo{"registry.terraform.io/bytebase/bytebase": {
			Protocol: string(config.Protocol), ProtocolVersion: config.ProtocolVersion, Pid: config.Pid, Test: true,
			Addr: tfexec.ReattachConfigAddr{Network: config.Addr.Network(), String: config.Addr.String()},
		}}, stop
	case err := <-serveErr:
		t.Fatalf("serve provider: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("provider RPC server did not start")
	}
	return nil, stop
}

func keyedIAMExpectAction(_ string, expected string) keyedIAMPlanCheck {
	return func(change *tfjson.Change) error {
		if len(change.Actions) != 1 || string(change.Actions[0]) != expected {
			return errors.Errorf("expected %s action, got %v", expected, change.Actions)
		}
		return nil
	}
}

func keyedIAMStateValues(state *tfjson.State) map[string]interface{} {
	for _, resource := range state.Values.RootModule.Resources {
		if resource.Address == keyedIAMAddress {
			return resource.AttributeValues
		}
	}
	return nil
}

func keyedIAMResourceAttr(_ string, path, expected string) func(*tfjson.State) error {
	return func(state *tfjson.State) error {
		var value interface{} = keyedIAMStateValues(state)
		for _, part := range strings.Split(path, ".") {
			if part == "#" {
				value = fmt.Sprint(len(value.([]interface{})))
			} else {
				value = value.(map[string]interface{})[part]
			}
		}
		if value != expected {
			return errors.Errorf("%s: expected %s, got %v", path, expected, value)
		}
		return nil
	}
}

func keyedIAMConfig(parent, bindings string) string {
	parentAttribute := ""
	if parent != "" {
		parentAttribute = fmt.Sprintf("parent = %q", parent)
	}
	return fmt.Sprintf(`
resource "bytebase_iam_policy" "test" {
  %s
  bindings = {
    %s
  }
}
`, parentAttribute, bindings)
}

func keyedIAMFactories(client *keyedIAMTestClient, legacy bool) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bytebase": func() (tfprotov6.ProviderServer, error) {
			sdk := NewProvider()
			// Supply a computed reference without another provider or Terraform >1.0.
			sdk.ResourcesMap["bytebase_test_value"] = &schema.Resource{
				CreateContext: func(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
					data.SetId("alice")
					return nil
				},
				ReadContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
				DeleteContext: func(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
					data.SetId("")
					return nil
				},
				Schema: map[string]*schema.Schema{},
			}
			sdk.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
				return client, nil
			}
			if legacy {
				return tf5to6server.UpgradeServer(context.Background(), sdk.GRPCProvider)
			}
			return NewProtocol6Server(context.Background(), sdk)
		},
	}
}

func keyedIAMPreApply(check keyedIAMPlanCheck) keyedIAMConfigPlanChecks {
	return keyedIAMConfigPlanChecks{PreApply: []keyedIAMPlanCheck{check}}
}

type keyedIAMPlanCheck func(*tfjson.Change) error

func keyedIAMBindingChange(key, field string, expectedBefore, expectedAfter interface{}) keyedIAMPlanCheck {
	return keyedIAMPlanCheck(func(change *tfjson.Change) error {
		if !change.Actions.Update() {
			return errors.Errorf("binding edit must update in place, got %v", change.Actions)
		}
		before, after := keyedIAMPlanBindings(change.Before), keyedIAMPlanBindings(change.After)
		if len(before) != len(after) || len(before) != 2 {
			return errors.Errorf("binding edit changed map keys: before=%v after=%v", before, after)
		}
		for bindingKey, beforeValue := range before {
			afterValue, ok := after[bindingKey]
			if !ok {
				return errors.Errorf("binding edit lost key %q", bindingKey)
			}
			if bindingKey != key {
				if !reflect.DeepEqual(beforeValue, afterValue) {
					return errors.Errorf("binding edit changed unrelated binding %q: before=%v after=%v", bindingKey, beforeValue, afterValue)
				}
				continue
			}
			oldBinding, newBinding := beforeValue.(map[string]interface{}), afterValue.(map[string]interface{})
			for attribute, oldValue := range oldBinding {
				newValue := newBinding[attribute]
				if attribute == field {
					if field == "members" {
						oldValue, newValue = keyedIAMSortedMembers(oldValue), keyedIAMSortedMembers(newValue)
					}
					if !reflect.DeepEqual(oldValue, expectedBefore) || !reflect.DeepEqual(newValue, expectedAfter) {
						return errors.Errorf("unexpected %s change: before=%v after=%v", field, oldValue, newValue)
					}
					continue
				}
				if attribute == "condition" && strings.HasPrefix(field, "condition.") {
					oldCondition, newCondition := oldValue.(map[string]interface{}), newValue.(map[string]interface{})
					for conditionField, oldConditionValue := range oldCondition {
						newConditionValue := newCondition[conditionField]
						if "condition."+conditionField == field {
							if oldConditionValue != expectedBefore || newConditionValue != expectedAfter {
								return errors.Errorf("unexpected %s change: before=%v after=%v", field, oldConditionValue, newConditionValue)
							}
						} else if !reflect.DeepEqual(oldConditionValue, newConditionValue) {
							return errors.Errorf("condition edit changed unrelated field %q", conditionField)
						}
					}
					continue
				}
				if !reflect.DeepEqual(oldValue, newValue) {
					return errors.Errorf("binding edit changed unrelated field %q: before=%v after=%v", attribute, oldValue, newValue)
				}
			}
		}
		return nil
	})
}

func keyedIAMPlanBindings(value interface{}) map[string]interface{} {
	return value.(map[string]interface{})["bindings"].(map[string]interface{})
}

func keyedIAMSortedMembers(value interface{}) []string {
	var members []string
	for _, member := range value.([]interface{}) {
		members = append(members, member.(string))
	}
	slices.Sort(members)
	return members
}

type keyedIAMTestClient struct {
	api.Client
	mu           sync.Mutex
	policy       *v1pb.IamPolicy
	revision     int
	writes       int
	conflicts    int
	checkEtag    bool
	conflictNext bool
}

func newKeyedIAMTestClient() *keyedIAMTestClient {
	return &keyedIAMTestClient{policy: &v1pb.IamPolicy{Etag: "1"}, revision: 1, checkEtag: true}
}

func (*keyedIAMTestClient) GetWorkspaceName() string { return "workspaces/test-workspace" }

func (*keyedIAMTestClient) GetRole(_ context.Context, name string) (*v1pb.Role, error) {
	return &v1pb.Role{Name: name, Permissions: []string{"bb.sql.select", "bb.sql.ddl"}}, nil
}

func (client *keyedIAMTestClient) GetProjectIAMPolicy(context.Context, string) (*v1pb.IamPolicy, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	return proto.Clone(client.policy).(*v1pb.IamPolicy), nil
}

func (client *keyedIAMTestClient) GetWorkspaceIAMPolicy(ctx context.Context) (*v1pb.IamPolicy, error) {
	return client.GetProjectIAMPolicy(ctx, "")
}

func (client *keyedIAMTestClient) SetProjectIAMPolicy(_ context.Context, parent string, request *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.writes++
	if request.Resource != parent {
		return nil, errors.Errorf("policy request parent %q does not match %q", request.Resource, parent)
	}
	if client.conflictNext {
		client.conflictNext = false
		client.nextEtag()
	}
	if client.checkEtag && request.Policy.Etag != client.policy.Etag {
		client.conflicts++
		return nil, errors.Errorf("etag conflict: sent %q, current %q", request.Policy.Etag, client.policy.Etag)
	}
	client.policy = proto.Clone(request.Policy).(*v1pb.IamPolicy)
	client.nextEtag()
	return proto.Clone(client.policy).(*v1pb.IamPolicy), nil
}

func (client *keyedIAMTestClient) SetWorkspaceIAMPolicy(ctx context.Context, request *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	return client.SetProjectIAMPolicy(ctx, "workspaces/test-workspace", request)
}

// nextEtag is called while holding mu.
func (client *keyedIAMTestClient) nextEtag() {
	client.revision++
	client.policy.Etag = fmt.Sprint(client.revision)
}

func (client *keyedIAMTestClient) assertNoConflicts(t *testing.T) {
	t.Helper()
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.conflicts != 0 {
		t.Fatalf("provider failed to preserve the latest etag: %d conflicts", client.conflicts)
	}
}
