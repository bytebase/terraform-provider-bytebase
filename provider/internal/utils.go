package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const (
	// EnvironmentNamePrefix is the prefix for environment unique name.
	EnvironmentNamePrefix = "environments/"
	// InstanceNamePrefix is the prefix for instance unique name.
	InstanceNamePrefix = "instances/"
	// ProjectNamePrefix is the prefix for project unique name.
	ProjectNamePrefix = "projects/"
	// DatabaseIDPrefix is the prefix for database unique name.
	DatabaseIDPrefix = "databases/"
	// PolicyNamePrefix is the prefix for policy unique name.
	PolicyNamePrefix = "policies/"
	// SettingNamePrefix is the prefix for setting unique name.
	SettingNamePrefix = "settings/"
	// ResourceIDPattern is the pattern for resource id.
	ResourceIDPattern = "[a-z]([a-z0-9-]{0,61}[a-z0-9])?"
)

var (
	resourceIDRegex = regexp.MustCompile(fmt.Sprintf("^%s$", ResourceIDPattern))
)

// ResourceIDValidation is the resource id regexp validation.
var ResourceIDValidation = validation.StringMatch(resourceIDRegex, fmt.Sprintf("resource id must matches %v", resourceIDRegex))

// ResourceNameValidation validate the resource name with prefix.
func ResourceNameValidation(regexs ...*regexp.Regexp) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		v, ok := i.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Bad data type",
				Detail:        "expected type to be string",
				AttributePath: path,
			})
			return diags
		}
		for _, regex := range regexs {
			if ok := regex.MatchString(v); ok {
				return diags
			}
		}
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Resource id not match",
			Detail:        fmt.Sprintf("resource id must matches %s pattern", ResourceIDPattern),
			AttributePath: path,
		})
		return diags
	}
}

// GetPolicyParentAndType returns the policy parent and type by the name.
func GetPolicyParentAndType(name string) (string, api.PolicyType, error) {
	names := strings.Split(name, PolicyNamePrefix)
	if len(names) != 2 {
		return "", "", errors.Errorf("invalid policy name %s", name)
	}
	policyType := api.PolicyType(strings.ToUpper(names[1]))

	return strings.TrimSuffix(names[0], "/"), policyType, nil
}

// GetEnvironmentID will parse the environment resource id.
func GetEnvironmentID(name string) (string, error) {
	// the environment request should be environments/{environment-id}
	tokens, err := getNameParentTokens(name, EnvironmentNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetInstanceID will parse the environment resource id and instance resource id.
func GetInstanceID(name string) (string, error) {
	// the instance request should be instances/{instance-id}
	tokens, err := getNameParentTokens(name, InstanceNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetProjectID will parse the project resource id.
func GetProjectID(name string) (string, error) {
	tokens, err := getNameParentTokens(name, ProjectNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetInstanceDatabaseID will parse the instance resource id and database name.
func GetInstanceDatabaseID(name string) (string, string, error) {
	// the instance request should be instances/{instance-id}/databases/{database-id}
	tokens, err := getNameParentTokens(name, InstanceNamePrefix, DatabaseIDPrefix)
	if err != nil {
		return "", "", err
	}
	return tokens[0], tokens[1], nil
}

func getNameParentTokens(name string, tokenPrefixes ...string) ([]string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2*len(tokenPrefixes) {
		return nil, errors.Errorf("invalid name %q", name)
	}

	var tokens []string
	for i, tokenPrefix := range tokenPrefixes {
		if fmt.Sprintf("%s/", parts[2*i]) != tokenPrefix {
			return nil, errors.Errorf("invalid prefix %q in request %q", tokenPrefix, name)
		}
		if parts[2*i+1] == "" {
			return nil, errors.Errorf("invalid name %q with empty prefix %q", name, tokenPrefix)
		}
		tokens = append(tokens, parts[2*i+1])
	}
	return tokens, nil
}
