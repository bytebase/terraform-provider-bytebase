package internal

import (
	"github.com/bytebase/terraform-provider-bytebase/api"
	"github.com/pkg/errors"
)

var environmentMap map[int]*api.Environment
var instanceMap map[int]*api.Instance

func init() {
	environmentMap = map[int]*api.Environment{}
	instanceMap = map[int]*api.Instance{}
}

type mockClient struct {
	environmentMap map[int]*api.Environment
	instanceMap    map[int]*api.Instance
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		environmentMap: environmentMap,
		instanceMap:    instanceMap,
	}, nil
}

// Login will login the user and get the response.
func (*mockClient) Login() (*api.AuthResponse, error) {
	return &api.AuthResponse{}, nil
}

// CreateEnvironment creates the environment.
func (c *mockClient) CreateEnvironment(create *api.EnvironmentCreate) (*api.Environment, error) {
	order := len(c.environmentMap) + 1
	if v := create.Order; v != nil {
		order = *v
	}

	env := &api.Environment{
		ID:    len(c.environmentMap) + 1,
		Name:  create.Name,
		Order: order,
	}

	if existed := c.findEnvironmentByName(env.Name); existed != nil {
		return nil, errors.Errorf("Environment %s already exists", create.Name)
	}
	c.environmentMap[env.ID] = env

	return env, nil
}

// GetEnvironment gets the environment by id.
func (c *mockClient) GetEnvironment(environmentID int) (*api.Environment, error) {
	env, ok := c.environmentMap[environmentID]
	if !ok {
		return nil, errors.Errorf("Cannot found environment with ID %d", environmentID)
	}

	return env, nil
}

// ListEnvironment finds all environments.
func (*mockClient) ListEnvironment() ([]*api.Environment, error) {
	return nil, errors.Errorf("ListEnvironment is not implemented yet")
}

// UpdateEnvironment updates the environment.
func (c *mockClient) UpdateEnvironment(environmentID int, patch *api.EnvironmentPatch) (*api.Environment, error) {
	env, err := c.GetEnvironment(environmentID)
	if err != nil {
		return nil, err
	}

	if v := patch.Name; v != nil {
		env.Name = *v
	}
	if v := patch.Order; v != nil {
		env.Order = *v
	}

	if existed := c.findEnvironmentByName(env.Name); existed != nil {
		return nil, errors.Errorf("Environment %s already exists", env.Name)
	}

	delete(c.environmentMap, env.ID)
	c.environmentMap[env.ID] = env

	return env, nil
}

// DeleteEnvironment deletes the environment.
func (c *mockClient) DeleteEnvironment(environmentID int) error {
	delete(c.environmentMap, environmentID)
	return nil
}

func (c *mockClient) findEnvironmentByName(envName string) *api.Environment {
	for _, env := range c.environmentMap {
		if env.Name == envName {
			return env
		}
	}
	return nil
}

// ListInstance will return all instances.
func (*mockClient) ListInstance() ([]*api.Instance, error) {
	return nil, errors.Errorf("ListInstance is not implemented yet")
}

// CreateInstance creates the instance.
func (*mockClient) CreateInstance(_ *api.InstanceCreate) (*api.Instance, error) {
	return nil, errors.Errorf("CreateInstance is not implemented yet")
}

// GetInstance gets the instance by id.
func (*mockClient) GetInstance(_ int) (*api.Instance, error) {
	return nil, errors.Errorf("GetInstance is not implemented yet")
}

// UpdateInstance updates the instance.
func (*mockClient) UpdateInstance(_ int, _ *api.InstancePatch) (*api.Instance, error) {
	return nil, errors.Errorf("UpdateInstance is not implemented yet")
}

// DeleteInstance deletes the instance.
func (*mockClient) DeleteInstance(_ int) error {
	return errors.Errorf("DeleteInstance is not implemented yet")
}
