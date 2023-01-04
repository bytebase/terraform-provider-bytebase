package internal

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

var environmentMap map[int]*api.Environment
var instanceMap map[string]*api.InstanceMessage
var roleMap map[string]*api.Role

func init() {
	environmentMap = map[int]*api.Environment{}
	instanceMap = map[string]*api.InstanceMessage{}
	roleMap = map[string]*api.Role{}
}

type mockClient struct {
	environmentMap map[int]*api.Environment
	instanceMap    map[string]*api.InstanceMessage
	roleMap        map[string]*api.Role
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
		ID:                     rand.Intn(1000),
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

// ListInstance will return instances in environment.
func (c *mockClient) ListInstance(_ context.Context, find *api.InstanceFindMessage) (*api.ListInstanceMessage, error) {
	instances := make([]*api.InstanceMessage, 0)
	for _, instance := range c.instanceMap {
		envID, _, err := GetEnvironmentInstanceID(instance.Name)
		if err != nil {
			return nil, err
		}
		if instance.State == api.Deleted && !find.ShowDeleted {
			continue
		}
		if find.EnvironmentID == "-" || find.EnvironmentID == envID {
			instances = append(instances, instance)
		}
	}

	return &api.ListInstanceMessage{
		Instances: instances,
	}, nil
}

// GetInstance gets the instance by id.
func (c *mockClient) GetInstance(_ context.Context, find *api.InstanceFindMessage) (*api.InstanceMessage, error) {
	name := fmt.Sprintf("environments/%s/instances/%s", find.EnvironmentID, find.InstanceID)
	ins, ok := c.instanceMap[name]
	if !ok {
		return nil, errors.Errorf("Cannot found instance %s", name)
	}

	return ins, nil
}

// CreateInstance creates the instance.
func (c *mockClient) CreateInstance(_ context.Context, environmentID, instanceID string, instance *api.InstanceMessage) (*api.InstanceMessage, error) {
	ins := &api.InstanceMessage{
		Name:         fmt.Sprintf("environments/%s/instances/%s", environmentID, instanceID),
		State:        api.Active,
		Title:        instance.Title,
		Engine:       instance.Engine,
		ExternalLink: instance.ExternalLink,
		DataSources:  instance.DataSources,
	}

	c.instanceMap[ins.Name] = ins
	return ins, nil
}

// UpdateInstance updates the instance.
func (c *mockClient) UpdateInstance(ctx context.Context, environmentID, instanceID string, patch *api.InstancePatchMessage) (*api.InstanceMessage, error) {
	ins, err := c.GetInstance(ctx, &api.InstanceFindMessage{
		InstanceID:    instanceID,
		EnvironmentID: environmentID,
	})
	if err != nil {
		return nil, err
	}

	if v := patch.Title; v != nil {
		ins.Title = *v
	}
	if v := patch.ExternalLink; v != nil {
		ins.ExternalLink = *v
	}
	if v := patch.DataSources; v != nil {
		ins.DataSources = v
	}

	c.instanceMap[ins.Name] = ins
	return ins, nil
}

// DeleteInstance deletes the instance.
func (c *mockClient) DeleteInstance(ctx context.Context, environmentID, instanceID string) error {
	ins, err := c.GetInstance(ctx, &api.InstanceFindMessage{
		EnvironmentID: environmentID,
		InstanceID:    instanceID,
	})
	if err != nil {
		return err
	}

	ins.State = api.Deleted
	c.instanceMap[ins.Name] = ins

	return nil
}

// UndeleteInstance undeletes the instance.
func (c *mockClient) UndeleteInstance(ctx context.Context, environmentID, instanceID string) (*api.InstanceMessage, error) {
	ins, err := c.GetInstance(ctx, &api.InstanceFindMessage{
		EnvironmentID: environmentID,
		InstanceID:    instanceID,
	})
	if err != nil {
		return nil, err
	}

	ins.State = api.Active
	c.instanceMap[ins.Name] = ins

	return ins, nil
}

// CreateRole creates the role in the instance.
func (c *mockClient) CreateRole(_ context.Context, environmentID, instanceID string, create *api.RoleUpsert) (*api.Role, error) {
	id := getRoleMapID(environmentID, instanceID, create.Title)

	if _, ok := c.roleMap[id]; ok {
		return nil, errors.Errorf("Role %s already exists", create.Title)
	}

	role := &api.Role{
		Name:            id,
		Title:           create.Title,
		ConnectionLimit: -1,
		Attribute:       &api.RoleAttribute{},
	}
	if v := create.ConnectionLimit; v != nil {
		role.ConnectionLimit = *v
	}
	if v := create.ValidUntil; v != nil {
		role.ValidUntil = v
	}
	if v := create.Attribute; v != nil {
		role.Attribute = v
	}

	c.roleMap[id] = role
	return role, nil
}

// GetRole gets the role by instance id and role name.
func (c *mockClient) GetRole(_ context.Context, environmentID, instanceID, roleName string) (*api.Role, error) {
	id := getRoleMapID(environmentID, instanceID, roleName)
	role, ok := c.roleMap[id]
	if !ok {
		return nil, errors.Errorf("Cannot found role with ID %s", id)
	}

	return role, nil
}

func (c *mockClient) ListRole(_ context.Context, environmentID, instanceID string) ([]*api.Role, error) {
	res := []*api.Role{}
	regex := regexp.MustCompile(fmt.Sprintf("^environments/%s/instances/%s/roles/", environmentID, instanceID))
	for key, role := range c.roleMap {
		if regex.MatchString(key) {
			res = append(res, role)
		}
	}

	return res, nil
}

// UpdateRole updates the role in instance.
func (c *mockClient) UpdateRole(ctx context.Context, environmentID, instanceID, roleName string, patch *api.RoleUpsert) (*api.Role, error) {
	role, err := c.GetRole(ctx, environmentID, instanceID, roleName)
	if err != nil {
		return nil, err
	}

	newRole := &api.Role{
		Name:            getRoleMapID(environmentID, instanceID, patch.Title),
		Title:           patch.Title,
		ConnectionLimit: role.ConnectionLimit,
		ValidUntil:      role.ValidUntil,
		Attribute:       role.Attribute,
	}
	if err := c.DeleteRole(ctx, environmentID, instanceID, roleName); err != nil {
		return nil, err
	}

	if v := patch.ConnectionLimit; v != nil {
		newRole.ConnectionLimit = *v
	}
	if v := patch.ValidUntil; v != nil {
		newRole.ValidUntil = v
	}
	if v := patch.Attribute; v != nil {
		newRole.Attribute = v
	}

	c.roleMap[newRole.Name] = newRole
	return role, nil
}

// DeleteRole deletes the role in the instance.
func (c *mockClient) DeleteRole(_ context.Context, environmentID, instanceID, roleName string) error {
	delete(c.roleMap, getRoleMapID(environmentID, instanceID, roleName))
	return nil
}

func getRoleMapID(environmentID, instanceID, roleName string) string {
	return fmt.Sprintf("environments/%s/instances/%s/roles/%s", environmentID, instanceID, roleName)
}
