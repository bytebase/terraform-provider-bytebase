package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

const (
	// EnvironmentNamePrefix is the prefix for environment unique name.
	EnvironmentNamePrefix  = "environments/"
	instanceNamePrefix     = "instances/"
	instanceRoleNamePrefix = "roles/"
	projectNamePrefix      = "projects/"
	databaseIDPrefix       = "databases/"
	policyNamePrefix       = "policies/"
)

var (
	resourceIDRegex = regexp.MustCompile("^([0-9a-z]+-?)+[0-9a-z]+$")
)

// ResourceIDValidation is the resource id regexp validation.
var ResourceIDValidation = validation.StringMatch(resourceIDRegex, fmt.Sprintf("resource id must matches %v", resourceIDRegex))

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
	tokens, err := getNameParentTokens(name, instanceNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetInstanceRoleID will parse the instance resource id and the role name.
func GetInstanceRoleID(name string) (string, string, error) {
	// the instance request should be instances/{instance-id}/roles/{role-name}
	tokens, err := getNameParentTokens(name, instanceNamePrefix, instanceRoleNamePrefix)
	if err != nil {
		return "", "", err
	}
	return tokens[0], tokens[1], nil
}

// GetPolicyFindMessageByName will generate the policy find by the name.
func GetPolicyFindMessageByName(name string) (*api.PolicyFindMessage, error) {
	tokens := strings.Split(name, policyNamePrefix)
	if len(tokens) != 2 {
		return nil, errors.Errorf("invalid policy name %s", name)
	}

	parent := tokens[0]
	policyType := api.PolicyType(strings.ToUpper(tokens[1]))
	find := &api.PolicyFindMessage{
		Type: &policyType,
	}

	if strings.HasSuffix(parent, "/") {
		parent = parent[:(len(parent) - 1)]
	}

	if parent == "" {
		return find, nil
	}

	if strings.HasPrefix(parent, projectNamePrefix) {
		// project policy request name should be projects/{project id}
		projectID, err := GetProjectID(parent)
		if err != nil {
			return nil, err
		}
		find.ProjectID = &projectID
		return find, nil
	}

	if strings.HasPrefix(parent, EnvironmentNamePrefix) {
		// environment policy request name should be environments/{environment id}
		environmentID, err := GetEnvironmentID(parent)
		if err != nil {
			return nil, err
		}
		find.EnvironmentID = &environmentID
		return find, nil
	}

	if strings.HasPrefix(parent, instanceNamePrefix) {
		sections := strings.Split(parent, "/")
		if len(sections) == 2 {
			// instance policy request name should be instances/{instance id}
			instanceID, err := GetInstanceID(parent)
			if err != nil {
				return nil, err
			}
			find.InstanceID = &instanceID
			return find, nil
		} else if len(sections) == 4 {
			// database policy request name should be instances/{instance id}/databases/{db name}
			instanceID, databaseName, err := getInstanceDatabaseID(parent)
			if err != nil {
				return nil, err
			}
			find.InstanceID = &instanceID
			find.DatabaseName = &databaseName
			return find, nil
		}
	}

	return nil, errors.Errorf("invalid policy name %s", name)
}

// GetProjectID will parse the project resource id.
func GetProjectID(name string) (string, error) {
	tokens, err := getNameParentTokens(name, projectNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

func getInstanceDatabaseID(name string) (string, string, error) {
	// the instance request should be instances/{instance-id}/databases/{database-id}
	tokens, err := getNameParentTokens(name, instanceNamePrefix, databaseIDPrefix)
	if err != nil {
		return "", "", err
	}
	return tokens[0], tokens[1], nil
}

func getNameParentTokens(name string, tokenPrefixes ...string) ([]string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2*len(tokenPrefixes) {
		return nil, errors.Errorf("invalid request %q", name)
	}

	var tokens []string
	for i, tokenPrefix := range tokenPrefixes {
		if fmt.Sprintf("%s/", parts[2*i]) != tokenPrefix {
			return nil, errors.Errorf("invalid prefix %q in request %q", tokenPrefix, name)
		}
		if parts[2*i+1] == "" {
			return nil, errors.Errorf("invalid request %q with empty prefix %q", name, tokenPrefix)
		}
		tokens = append(tokens, parts[2*i+1])
	}
	return tokens, nil
}
