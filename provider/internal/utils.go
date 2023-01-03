package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

const (
	environmentNamePrefix = "environments/"
	instanceNamePrefix    = "instances/"
)

var (
	resourceIDRegex = regexp.MustCompile("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$")
)

// ResourceIDValidation is the resource id regexp validation.
var ResourceIDValidation = validation.StringMatch(resourceIDRegex, fmt.Sprintf("resource id must matches %v", resourceIDRegex))

// GetEnvironmentInstanceID will parse the environment resource id and instance resource id.
func GetEnvironmentInstanceID(name string) (string, string, error) {
	// the instance request should be environments/{environment-id}/instances/{instance-id}
	tokens, err := getNameParentTokens(name, environmentNamePrefix, instanceNamePrefix)
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
