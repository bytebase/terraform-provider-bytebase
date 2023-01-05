package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

const (
	environmentNamePrefix  = "environments/"
	instanceNamePrefix     = "instances/"
	instanceRoleNamePrefix = "roles/"
)

var (
	resourceIDRegex = regexp.MustCompile("^([0-9a-z]+-?)+[0-9a-z]+$")
)

// ResourceIDValidation is the resource id regexp validation.
var ResourceIDValidation = validation.StringMatch(resourceIDRegex, fmt.Sprintf("resource id must matches %v", resourceIDRegex))

// GetEnvironmentID will parse the environment resource id.
func GetEnvironmentID(name string) (string, error) {
	// the environment request should be environments/{environment-id}
	tokens, err := getNameParentTokens(name, environmentNamePrefix)
	if err != nil {
		return "", err
	}
	return tokens[0], nil
}

// GetEnvironmentInstanceID will parse the environment resource id and instance resource id.
func GetEnvironmentInstanceID(name string) (string, string, error) {
	// the instance request should be environments/{environment-id}/instances/{instance-id}
	tokens, err := getNameParentTokens(name, environmentNamePrefix, instanceNamePrefix)
	if err != nil {
		return "", "", err
	}
	return tokens[0], tokens[1], nil
}

// GetEnvironmentInstanceRoleID will parse the environment resource id, instance resource id and the role name.
func GetEnvironmentInstanceRoleID(name string) (string, string, string, error) {
	// the instance request should be environments/{environment-id}/instances/{instance-id}/roles/{role-name}
	tokens, err := getNameParentTokens(name, environmentNamePrefix, instanceNamePrefix, instanceRoleNamePrefix)
	if err != nil {
		return "", "", "", err
	}
	return tokens[0], tokens[1], tokens[2], nil
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
