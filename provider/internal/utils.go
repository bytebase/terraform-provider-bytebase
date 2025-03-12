package internal

import (
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
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
	// VCSProviderNamePrefix is the prefix for vcs provider unique name.
	VCSProviderNamePrefix = "vcsProviders/"
	// VCSConnectorNamePrefix is the prefix for vcs connector unique name.
	VCSConnectorNamePrefix = "vcsConnectors/"
	// UserNamePrefix is the prefix for user name.
	UserNamePrefix = "users/"
	// GroupNamePrefix is the prefix for group name.
	GroupNamePrefix = "groups/"
	// RoleNamePrefix is the prefix for role name.
	RoleNamePrefix = "roles/"
	// ReviewConfigNamePrefix is the prefix for the review config name.
	ReviewConfigNamePrefix = "reviewConfigs/"
	// DatabaseCatalogNameSuffix is the suffix for the database catalog name.
	DatabaseCatalogNameSuffix = "/catalog"
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

var EngineValidation = validation.StringInSlice([]string{
	v1pb.Engine_CLICKHOUSE.String(),
	v1pb.Engine_MYSQL.String(),
	v1pb.Engine_POSTGRES.String(),
	v1pb.Engine_SNOWFLAKE.String(),
	v1pb.Engine_SQLITE.String(),
	v1pb.Engine_TIDB.String(),
	v1pb.Engine_MONGODB.String(),
	v1pb.Engine_REDIS.String(),
	v1pb.Engine_ORACLE.String(),
	v1pb.Engine_SPANNER.String(),
	v1pb.Engine_MSSQL.String(),
	v1pb.Engine_REDSHIFT.String(),
	v1pb.Engine_MARIADB.String(),
	v1pb.Engine_OCEANBASE.String(),
	v1pb.Engine_DM.String(),
	v1pb.Engine_RISINGWAVE.String(),
	v1pb.Engine_OCEANBASE_ORACLE.String(),
	v1pb.Engine_STARROCKS.String(),
	v1pb.Engine_DORIS.String(),
	v1pb.Engine_HIVE.String(),
	v1pb.Engine_ELASTICSEARCH.String(),
}, false)

// GetPolicyParentAndType returns the policy parent and type by the name.
func GetPolicyParentAndType(name string) (string, v1pb.PolicyType, error) {
	names := strings.Split(name, PolicyNamePrefix)
	if len(names) != 2 {
		return "", v1pb.PolicyType_POLICY_TYPE_UNSPECIFIED, errors.Errorf("invalid policy name %s", name)
	}
	policyType := strings.ToUpper(names[1])
	pType, ok := v1pb.PolicyType_value[policyType]
	if !ok {
		return "", v1pb.PolicyType_POLICY_TYPE_UNSPECIFIED, errors.Errorf("invalid policy name %s", name)
	}

	return strings.TrimSuffix(names[0], "/"), v1pb.PolicyType(pType), nil
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

// GetVCSProviderID will parse the vcs provider resource id.
func GetVCSProviderID(name string) (string, error) {
	// the vcs provider name should be vcsProviders/{resource-id}
	tokens, err := getNameParentTokens(name, VCSProviderNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetVCSConnectorID will parse the vcs connector resource id.
func GetVCSConnectorID(name string) (string, string, error) {
	// the vcs connector name should be projects/{project}/vcsConnectors/{resource-id}
	tokens, err := getNameParentTokens(name, ProjectNamePrefix, VCSConnectorNamePrefix)
	if err != nil {
		return "", "", err
	}
	return tokens[0], tokens[1], nil
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

// GetRoleID will parse the role resource id.
func GetRoleID(name string) (string, error) {
	tokens, err := getNameParentTokens(name, RoleNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetGroupEmail will parse the email from group full name.
func GetGroupEmail(name string) (string, error) {
	tokens, err := getNameParentTokens(name, GroupNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetReviewConfigID will parse the id from review config full name.
func GetReviewConfigID(name string) (string, error) {
	tokens, err := getNameParentTokens(name, ReviewConfigNamePrefix)
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

// ValidateMemberBinding checks the member binding format.
func ValidateMemberBinding(member string) error {
	if !strings.HasPrefix(member, "user:") && !strings.HasPrefix(member, "group:") {
		return errors.Errorf("invalid member format")
	}
	return nil
}

// ToHashcodeInt returns int by string.
func ToHashcodeInt(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
