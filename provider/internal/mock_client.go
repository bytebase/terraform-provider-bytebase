package internal

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

var environmentMap map[int]*api.Environment
var instanceMap map[int]*api.Instance
var roleMap map[int]*api.PGRole

func init() {
	environmentMap = map[int]*api.Environment{}
	instanceMap = map[int]*api.Instance{}
	roleMap = map[int]*api.PGRole{}
}

type mockClient struct {
	environmentMap map[int]*api.Environment
	instanceMap    map[int]*api.Instance
	roleMap        map[int]*api.PGRole
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		environmentMap: environmentMap,
		instanceMap:    instanceMap,
		roleMap:        roleMap,
	}, nil
}

// Login will login the user and get the response.
func (*mockClient) Login() (*api.AuthResponse, error) {
	return &api.AuthResponse{}, nil
}

// CreateEnvironment creates the environment.
func (c *mockClient) CreateEnvironment(_ context.Context, create *api.EnvironmentUpsert) (*api.Environment, error) {
	env := &api.Environment{
		ID:                     len(c.environmentMap) + 1,
		Name:                   *create.Name,
		Order:                  *create.Order,
		PipelineApprovalPolicy: create.PipelineApprovalPolicy,
		EnvironmentTierPolicy:  create.EnvironmentTierPolicy,
		BackupPlanPolicy:       create.BackupPlanPolicy,
	}

	if existed := c.findEnvironmentByName(env.Name); existed != nil {
		return nil, errors.Errorf("Environment %s already exists", *create.Name)
	}
	c.environmentMap[env.ID] = env

	return env, nil
}

// GetEnvironment gets the environment by id.
func (c *mockClient) GetEnvironment(_ context.Context, environmentID int) (*api.Environment, error) {
	env, ok := c.environmentMap[environmentID]
	if !ok {
		return nil, errors.Errorf("Cannot found environment with ID %d", environmentID)
	}

	return env, nil
}

// ListEnvironment finds all environments.
func (c *mockClient) ListEnvironment(_ context.Context, find *api.EnvironmentFind) ([]*api.Environment, error) {
	environments := make([]*api.Environment, 0)
	for _, env := range c.environmentMap {
		if find.Name == "" || env.Name == find.Name {
			environments = append(environments, env)
		}
	}

	return environments, nil
}

// UpdateEnvironment updates the environment.
func (c *mockClient) UpdateEnvironment(ctx context.Context, environmentID int, patch *api.EnvironmentUpsert) (*api.Environment, error) {
	env, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, err
	}

	if v := patch.Name; v != nil {
		if existed := c.findEnvironmentByName(*v); existed != nil {
			return nil, errors.Errorf("Environment %s already exists", env.Name)
		}
		env.Name = *v
	}
	if v := patch.Order; v != nil {
		env.Order = *v
	}
	if v := patch.PipelineApprovalPolicy; v != nil {
		env.PipelineApprovalPolicy = v
	}
	if v := patch.EnvironmentTierPolicy; v != nil {
		env.EnvironmentTierPolicy = v
	}
	if v := patch.BackupPlanPolicy; v != nil {
		env.BackupPlanPolicy = v
	}

	delete(c.environmentMap, env.ID)
	c.environmentMap[env.ID] = env

	return env, nil
}

// DeleteEnvironment deletes the environment.
func (c *mockClient) DeleteEnvironment(_ context.Context, environmentID int) error {
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
func (c *mockClient) ListInstance(_ context.Context, find *api.InstanceFind) ([]*api.Instance, error) {
	instances := make([]*api.Instance, 0)
	for _, instance := range c.instanceMap {
		if find.Name == "" || instance.Name == find.Name {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// CreateInstance creates the instance.
func (c *mockClient) CreateInstance(_ context.Context, create *api.InstanceCreate) (*api.Instance, error) {
	dataSourceList := []*api.DataSource{}
	for _, dataSource := range create.DataSourceList {
		dataSourceList = append(dataSourceList, &api.DataSource{
			Name:         dataSource.Name,
			Type:         dataSource.Type,
			Username:     dataSource.Username,
			HostOverride: dataSource.HostOverride,
			PortOverride: dataSource.PortOverride,
		})
	}

	ins := &api.Instance{
		ID:             len(c.instanceMap) + 1,
		Environment:    create.Environment,
		Name:           create.Name,
		Engine:         create.Engine,
		ExternalLink:   create.ExternalLink,
		Host:           create.Host,
		Port:           create.Port,
		Database:       create.Database,
		DataSourceList: dataSourceList,
	}

	c.instanceMap[ins.ID] = ins
	return ins, nil
}

// GetInstance gets the instance by id.
func (c *mockClient) GetInstance(_ context.Context, instanceID int) (*api.Instance, error) {
	ins, ok := c.instanceMap[instanceID]
	if !ok {
		return nil, errors.Errorf("Cannot found instance with ID %d", instanceID)
	}

	return ins, nil
}

// UpdateInstance updates the instance.
func (c *mockClient) UpdateInstance(ctx context.Context, instanceID int, patch *api.InstancePatch) (*api.Instance, error) {
	ins, err := c.GetInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	if v := patch.Name; v != nil {
		ins.Name = *v
	}
	if v := patch.ExternalLink; v != nil {
		ins.ExternalLink = *v
	}
	if v := patch.Host; v != nil {
		ins.Host = *v
	}
	if v := patch.Port; v != nil {
		ins.Port = *v
	}

	c.instanceMap[instanceID] = ins
	return ins, nil
}

// DeleteInstance deletes the instance.
func (c *mockClient) DeleteInstance(_ context.Context, instanceID int) error {
	delete(c.instanceMap, instanceID)
	return nil
}

// CreatePGRole creates the role in the instance.
func (c *mockClient) CreatePGRole(ctx context.Context, instanceID int, create *api.PGRoleUpsert) (*api.PGRole, error) {
	return nil, errors.Errorf("not implement yet")
}

// GetPGRole gets the role by instance id and role name.
func (c *mockClient) GetPGRole(ctx context.Context, instanceID int, roleName string) (*api.PGRole, error) {
	return nil, errors.Errorf("not implement yet")
}

// UpdatePGRole updates the role in instance.
func (c *mockClient) UpdatePGRole(ctx context.Context, instanceID int, patch *api.PGRoleUpsert) (*api.PGRole, error) {
	return nil, errors.Errorf("not implement yet")
}

// DeletePGRole deletes the role in the instance.
func (c *mockClient) DeletePGRole(ctx context.Context, instanceID int, roleName string) error {
	return errors.Errorf("not implement yet")
}
