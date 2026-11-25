package provider

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func iamBindingIdentity(binding *v1pb.Binding) string {
	condition := binding.GetCondition()
	kind := "raw"
	if _, supported := parseIAMConditionExpression(condition.GetExpression()); supported {
		kind = "structured"
	}
	identity, _ := json.Marshal([6]string{
		binding.GetRole(), condition.GetTitle(), condition.GetDescription(), condition.GetLocation(), kind, normalizeIAMConditionExpression(condition.GetExpression()),
	})
	return string(identity)
}

// Normalize only complete conditions expressible by the structured schema.
// Unknown CEL remains untouched rather than being partially interpreted.
func normalizeIAMConditionExpression(expression string) string {
	attributes, ok := parseIAMConditionExpression(expression)
	if !ok {
		return expression
	}
	clauses := make([]string, 0, len(attributes))
	for _, field := range iamConditionClauseFields {
		value, ok := attributes[field.attribute]
		if !ok {
			continue
		}
		if field.attribute == "expire_timestamp" {
			parsed, _ := time.Parse(time.RFC3339Nano, value.(string))
			value = parsed.UTC().Format(time.RFC3339Nano)
		}
		encoded, _ := json.Marshal(value)
		if field.attribute == "expire_timestamp" {
			encoded = []byte("timestamp(" + string(encoded) + ")")
		}
		clauses = append(clauses, field.expression+field.operator+string(encoded))
	}
	slices.Sort(clauses)
	return strings.Join(clauses, " && ")
}

var iamConditionClauseFields = []struct{ expression, attribute, operator string }{
	{"resource.database", "database", " == "},
	{"resource.schema_name", "schema", " == "},
	{"resource.table_name", "tables", " in "},
	{"resource.environment_id", "environment_ids", " in "},
	{"request.time", "expire_timestamp", " < "},
}

func parseIAMConditionExpression(expression string) (map[string]interface{}, bool) {
	attributes := make(map[string]interface{})
	if expression == "" {
		return attributes, true
	}
	remaining := strings.TrimSpace(expression)
	var fields []string
	for remaining != "" {
		matched := false
		for _, field := range iamConditionClauseFields {
			prefix := field.expression + field.operator
			if !strings.HasPrefix(remaining, prefix) {
				continue
			}
			if _, exists := attributes[field.attribute]; exists {
				return nil, false
			}
			remaining = strings.TrimSpace(strings.TrimPrefix(remaining, prefix))
			if field.attribute == "expire_timestamp" {
				if !strings.HasPrefix(remaining, "timestamp(") {
					return nil, false
				}
				remaining = strings.TrimPrefix(remaining, "timestamp(")
			}
			// Let the JSON decoder consume complete literals, including any
			// conjunction text or escaped quotes inside string values.
			decoder := json.NewDecoder(strings.NewReader(remaining))
			var literal json.RawMessage
			if decoder.Decode(&literal) != nil {
				return nil, false
			}
			remaining = strings.TrimSpace(remaining[decoder.InputOffset():])
			if field.operator == " in " {
				var elements []json.RawMessage
				if len(literal) == 0 || literal[0] != '[' || json.Unmarshal(literal, &elements) != nil {
					return nil, false
				}
				values := make([]string, 0, len(elements))
				for _, element := range elements {
					var value *string
					if json.Unmarshal(element, &value) != nil || value == nil {
						return nil, false
					}
					values = append(values, *value)
				}
				slices.Sort(values)
				attributes[field.attribute] = slices.Compact(values)
			} else {
				var value *string
				if json.Unmarshal(literal, &value) != nil || value == nil {
					return nil, false
				}
				if field.attribute == "expire_timestamp" {
					if _, err := time.Parse(time.RFC3339Nano, *value); err != nil || !strings.HasPrefix(remaining, ")") {
						return nil, false
					}
					remaining = strings.TrimSpace(strings.TrimPrefix(remaining, ")"))
				}
				attributes[field.attribute] = *value
			}
			fields = append(fields, field.attribute)
			matched = true
			break
		}
		if !matched {
			return nil, false
		}
		if remaining != "" {
			if !strings.HasPrefix(remaining, "&&") {
				return nil, false
			}
			remaining = strings.TrimSpace(strings.TrimPrefix(remaining, "&&"))
			if remaining == "" {
				return nil, false
			}
		}
	}
	// Bytebase only repairs edge conjunctions when stripping environment
	// predicates for non-DDL/DML permissions.
	for index, field := range fields {
		if field == "environment_ids" && index != 0 && index != len(fields)-1 {
			return nil, false
		}
	}
	return attributes, true
}

func validateIAMBindingIdentities(bindings map[string]*v1pb.Binding) error {
	seen := make(map[string]string)
	for _, key := range sortedIAMBindingKeys(bindings) {
		identity := iamBindingIdentity(bindings[key])
		if previous, ok := seen[identity]; ok {
			return errors.Errorf("bindings %q and %q have the same role and condition; combine their members into one binding", previous, key)
		}
		seen[identity] = key
	}
	return nil
}

// The API has no binding keys and permits duplicate identities. Merge their
// grants before matching keys so refresh never hides an out-of-band grant.
func reconcileIAMBindings(remote []*v1pb.Binding, prior map[string]*v1pb.Binding) map[string]*v1pb.Binding {
	merged := make(map[string]*v1pb.Binding)
	for _, binding := range remote {
		if binding == nil || len(binding.Members) == 0 {
			continue
		}
		identity := iamBindingIdentity(binding)
		if existing := merged[identity]; existing != nil {
			existing.Members = append(existing.Members, binding.Members...)
		} else {
			cloned := proto.Clone(binding).(*v1pb.Binding)
			if cloned.GetCondition().GetTitle() == "" && cloned.GetCondition().GetDescription() == "" && cloned.GetCondition().GetExpression() == "" && cloned.GetCondition().GetLocation() == "" {
				cloned.Condition = nil
			}
			merged[identity] = cloned
		}
	}
	for _, binding := range merged {
		slices.Sort(binding.Members)
		binding.Members = slices.Compact(binding.Members)
	}

	result := make(map[string]*v1pb.Binding, len(merged))
	reserved := make(map[string]bool, len(prior))
	for _, key := range sortedIAMBindingKeys(prior) {
		reserved[key] = true
		identity := iamBindingIdentity(prior[key])
		if binding := merged[identity]; binding != nil {
			result[key] = binding
			delete(merged, identity)
		}
	}

	identities := make([]string, 0, len(merged))
	for identity := range merged {
		identities = append(identities, identity)
	}
	slices.Sort(identities)
	for _, identity := range identities {
		binding := merged[identity]
		digest := sha256.Sum256([]byte(identity))
		base := fmt.Sprintf("%s#%x", binding.Role, digest[:6])
		key := base
		for suffix := 2; reserved[key]; suffix++ {
			key = fmt.Sprintf("%s-%d", base, suffix)
		}
		reserved[key] = true
		result[key] = binding
	}
	return result
}

func sortedIAMBindingKeys(bindings map[string]*v1pb.Binding) []string {
	keys := make([]string, 0, len(bindings))
	for key := range bindings {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
