package provider

import (
	"reflect"
	"strings"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"google.golang.org/genproto/googleapis/type/expr"
)

func TestIAMBindingIdentity(t *testing.T) {
	plain := &v1pb.Binding{Role: "roles/reader", Members: []string{"user:a@example.com"}}
	empty := &v1pb.Binding{Role: plain.Role, Condition: &expr.Expr{}, Members: []string{"user:b@example.com"}}
	if iamBindingIdentity(plain) != iamBindingIdentity(empty) {
		t.Fatal("membership and missing versus empty conditions must not change identity")
	}
	for _, condition := range []*expr.Expr{
		{Expression: "true"}, {Title: "title"}, {Description: "description"}, {Location: "source"},
	} {
		if iamBindingIdentity(plain) == iamBindingIdentity(&v1pb.Binding{Role: plain.Role, Condition: condition}) {
			t.Fatalf("condition must participate in identity: %v", condition)
		}
	}
	left := &v1pb.Binding{Role: "roles/reader", Condition: &expr.Expr{Title: "a|b", Description: "c"}}
	right := &v1pb.Binding{Role: "roles/reader", Condition: &expr.Expr{Title: "a", Description: "b|c"}}
	if iamBindingIdentity(left) == iamBindingIdentity(right) {
		t.Fatal("condition identity must not have delimiter collisions")
	}
}

func TestIAMBindingIdentityNormalizesSupportedConditions(t *testing.T) {
	left := &v1pb.Binding{Role: "roles/reader", Condition: &expr.Expr{Expression: `resource.database == "instances/prod/databases/app" && resource.table_name in ["orders","users"] && request.time < timestamp("2027-01-01T00:00:00Z")`}}
	right := &v1pb.Binding{Role: left.Role, Condition: &expr.Expr{Expression: `request.time < timestamp("2027-01-01T08:00:00+08:00") && resource.table_name in ["users", "orders", "users"] && resource.database == "instances/prod/databases/app"`}}
	if iamBindingIdentity(left) != iamBindingIdentity(right) {
		t.Fatal("set, conjunction ordering, or equivalent timestamps changed identity")
	}
	for _, expression := range []string{
		`resource.table_name in ["b","a"] || true`,
		`resource.database == "a" && resource.database == "b"`,
		`resource.table_name in ["b","a"] && unknown == true`,
		`resource.database == "a" && resource.environment_id in ["test"] && resource.table_name in ["b","a"]`,
	} {
		if normalizeIAMConditionExpression(expression) != expression {
			t.Fatalf("unsupported expression was rewritten: %s", expression)
		}
	}
}

func TestParseIAMConditionExpressionQuotedAndNullValues(t *testing.T) {
	parsed, ok := parseIAMConditionExpression(`resource.database == "a && b" && resource.table_name in ["A && B", "quoted\"table"]`)
	if !ok || parsed["database"] != "a && b" || !reflect.DeepEqual(parsed["tables"], []string{"A && B", `quoted"table`}) {
		t.Fatalf("quoted strings did not roundtrip: %v, %t", parsed, ok)
	}
	for _, expression := range []string{
		`resource.database == null`, `resource.table_name in [null]`,
		`resource.environment_id in [null, "prod"]`, `resource.table_name in [1]`,
	} {
		if _, ok := parseIAMConditionExpression(expression); ok {
			t.Fatalf("non-string CEL literal accepted: %s", expression)
		}
	}
}

func TestIAMBindingIdentitySeparatesUnsupportedConditions(t *testing.T) {
	valid := &v1pb.Binding{Role: "roles/reader", Condition: &expr.Expr{Expression: `resource.database == "a" && resource.schema_name == "b" && resource.environment_id in ["prod"]`}}
	unsupported := &v1pb.Binding{Role: valid.Role, Condition: &expr.Expr{Expression: normalizeIAMConditionExpression(valid.Condition.Expression)}}
	if iamBindingIdentity(valid) == iamBindingIdentity(unsupported) {
		t.Fatal("supported normalization collided with an unsupported environment-in-middle expression")
	}
	location := &v1pb.Binding{Role: valid.Role, Members: []string{"user:a@example.com"}, Condition: &expr.Expr{Location: "source"}}
	for _, binding := range reconcileIAMBindings([]*v1pb.Binding{location}, nil) {
		if binding.GetCondition().GetLocation() != "source" {
			t.Fatal("remote condition metadata was silently discarded")
		}
	}
}

func TestReconcileIAMBindingsPreservesKeysAndDrift(t *testing.T) {
	prior := map[string]*v1pb.Binding{
		"readers": {Role: "roles/reader", Members: []string{"user:a@example.com"}},
		"gone":    {Role: "roles/gone", Members: []string{"user:old@example.com"}},
	}
	remote := []*v1pb.Binding{
		{Role: "roles/reader", Condition: &expr.Expr{}, Members: []string{"user:b@example.com", "user:a@example.com"}},
		{Role: "roles/admin", Members: []string{"user:admin@example.com"}},
	}
	got := reconcileIAMBindings(remote, prior)
	if len(got) != 2 || got["readers"] == nil || got["gone"] != nil {
		t.Fatalf("lost keys or hidden remote binding: %v", got)
	}
	if !reflect.DeepEqual(got["readers"].Members, []string{"user:a@example.com", "user:b@example.com"}) {
		t.Fatalf("remote member drift was lost: %v", got["readers"].Members)
	}
	if len(prior["readers"].Members) != 1 || remote[0].Members[0] != "user:b@example.com" {
		t.Fatal("reconciliation mutated inputs")
	}
	reordered := reconcileIAMBindings([]*v1pb.Binding{remote[1], remote[0]}, got)
	if !reflect.DeepEqual(got, reordered) {
		t.Fatal("server ordering changed binding keys")
	}
}

func TestReconcileIAMBindingsCoalescesDuplicateRemoteGrants(t *testing.T) {
	remote := []*v1pb.Binding{
		{Role: "roles/reader", Members: []string{"user:a@example.com"}},
		{Role: "roles/reader", Condition: &expr.Expr{}, Members: []string{"user:b@example.com", "user:a@example.com"}},
	}
	prior := map[string]*v1pb.Binding{"readers": remote[0]}
	got := reconcileIAMBindings(remote, prior)
	if len(got) != 1 || got["readers"] == nil || !reflect.DeepEqual(got["readers"].Members, []string{"user:a@example.com", "user:b@example.com"}) {
		t.Fatalf("duplicate binding hid grants: %v", got)
	}
}

func TestReconcileIAMBindingsGeneratedKeysAvoidCollisions(t *testing.T) {
	remote := &v1pb.Binding{Role: "roles/reader", Members: []string{"user:a@example.com"}}
	initial := reconcileIAMBindings([]*v1pb.Binding{remote}, nil)
	var generated string
	for key := range initial {
		generated = key
	}
	prior := map[string]*v1pb.Binding{
		generated: {Role: "roles/different", Members: []string{"user:b@example.com"}},
	}
	got := reconcileIAMBindings([]*v1pb.Binding{remote}, prior)
	if len(got) != 1 || got[generated] != nil {
		t.Fatalf("new server binding stole an existing user key: %v", got)
	}
	if !reflect.DeepEqual(got, reconcileIAMBindings([]*v1pb.Binding{remote}, got)) {
		t.Fatal("generated collision key did not survive refresh")
	}
}

func TestValidateIAMBindingIdentities(t *testing.T) {
	bindings := map[string]*v1pb.Binding{
		"first":  {Role: "roles/reader", Members: []string{"user:a@example.com"}},
		"second": {Role: "roles/reader", Condition: &expr.Expr{}, Members: []string{"user:b@example.com"}},
	}
	err := validateIAMBindingIdentities(bindings)
	if err == nil || !strings.Contains(err.Error(), "first") || !strings.Contains(err.Error(), "second") {
		t.Fatalf("expected duplicate keys in error, got %v", err)
	}
	bindings["second"].Condition.Expression = "true"
	if err := validateIAMBindingIdentities(bindings); err != nil {
		t.Fatal(err)
	}
}
