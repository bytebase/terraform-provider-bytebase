package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

var environmentMap map[string]*api.EnvironmentMessage
var instanceMap map[string]*api.InstanceMessage
var policyMap map[string]*api.PolicyMessage
var roleMap map[string]*api.Role
var projectMap map[string]*api.ProjectMessage

func init() {
	environmentMap = map[string]*api.EnvironmentMessage{}
	instanceMap = map[string]*api.InstanceMessage{}
	policyMap = map[string]*api.PolicyMessage{}
	roleMap = map[string]*api.Role{}
	projectMap = map[string]*api.ProjectMessage{}
}

type mockClient struct {
	environmentMap map[string]*api.EnvironmentMessage
	instanceMap    map[string]*api.InstanceMessage
	policyMap      map[string]*api.PolicyMessage
	projectMap     map[string]*api.ProjectMessage
	roleMap        map[string]*api.Role
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		environmentMap: environmentMap,
		instanceMap:    instanceMap,
		policyMap:      policyMap,
		roleMap:        roleMap,
		projectMap:     projectMap,
	}, nil
}

// Login will login the user and get the response.
func (*mockClient) Login() (*api.AuthResponse, error) {
	return &api.AuthResponse{}, nil
}

// CreateEnvironment creates the environment.
func (c *mockClient) CreateEnvironment(_ context.Context, environmentID string, create *api.EnvironmentMessage) (*api.EnvironmentMessage, error) {
	env := &api.EnvironmentMessage{
		Name:  fmt.Sprintf("environments/%s", environmentID),
		Order: create.Order,
		Title: create.Title,
		State: api.Active,
		Tier:  create.Tier,
	}

	if _, ok := c.environmentMap[env.Name]; ok {
		return nil, errors.Errorf("Environment %s already exists", env.Name)
	}

	c.environmentMap[env.Name] = env

	return env, nil
}

// GetEnvironment gets the environment by id.
func (c *mockClient) GetEnvironment(_ context.Context, environmentID string) (*api.EnvironmentMessage, error) {
	env, ok := c.environmentMap[fmt.Sprintf("environments/%s", environmentID)]
	if !ok {
		return nil, errors.Errorf("Cannot found environment with ID %s", environmentID)
	}

	return env, nil
}

// ListEnvironment finds all environments.
func (c *mockClient) ListEnvironment(_ context.Context, showDeleted bool) (*api.ListEnvironmentMessage, error) {
	environments := make([]*api.EnvironmentMessage, 0)
	for _, env := range c.environmentMap {
		if env.State == api.Deleted && !showDeleted {
			continue
		}
		environments = append(environments, env)
	}

	return &api.ListEnvironmentMessage{
		Environments: environments,
	}, nil
}

// UpdateEnvironment updates the environment.
func (c *mockClient) UpdateEnvironment(ctx context.Context, environmentID string, patch *api.EnvironmentPatchMessage) (*api.EnvironmentMessage, error) {
	env, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, err
	}

	if v := patch.Title; v != nil {
		env.Title = *v
	}
	if v := patch.Order; v != nil {
		env.Order = *v
	}
	if v := patch.Tier; v != nil {
		env.Tier = *v
	}

	c.environmentMap[env.Name] = env

	return env, nil
}

// DeleteEnvironment deletes the environment.
func (c *mockClient) DeleteEnvironment(ctx context.Context, environmentID string) error {
	env, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		return err
	}

	env.State = api.Deleted
	c.environmentMap[env.Name] = env
	return nil
}

// UndeleteEnvironment undeletes the environment.
func (c *mockClient) UndeleteEnvironment(ctx context.Context, environmentID string) (*api.EnvironmentMessage, error) {
	env, err := c.GetEnvironment(ctx, environmentID)
	if err != nil {
		return nil, err
	}

	env.State = api.Active
	c.environmentMap[env.Name] = env
	return env, nil
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
		ShowDeleted:   false,
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
		ShowDeleted:   true,
	})
	if err != nil {
		return nil, err
	}

	ins.State = api.Active
	c.instanceMap[ins.Name] = ins

	return ins, nil
}

// SyncInstanceSchema will trigger the schema sync for an instance.
func (*mockClient) SyncInstanceSchema(_ context.Context, _ string) error {
	return nil
}

// CreateRole creates the role in the instance.
func (c *mockClient) CreateRole(_ context.Context, environmentID, instanceID string, create *api.RoleUpsert) (*api.Role, error) {
	id := getRoleMapID(environmentID, instanceID, create.RoleName)

	if _, ok := c.roleMap[id]; ok {
		return nil, errors.Errorf("Role %s already exists", create.RoleName)
	}

	role := &api.Role{
		Name:            id,
		RoleName:        create.RoleName,
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
		Name:            getRoleMapID(environmentID, instanceID, patch.RoleName),
		RoleName:        patch.RoleName,
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

// ListPolicies lists policies in a specific resource.
func (c *mockClient) ListPolicies(_ context.Context, find *api.PolicyFindMessage) (*api.ListPolicyMessage, error) {
	policies := make([]*api.PolicyMessage, 0)
	name := getPolicyRequestName(find)
	for _, policy := range c.policyMap {
		if policy.State == api.Deleted && !find.ShowDeleted {
			continue
		}
		if name == "policies" || policy.Name == name {
			policies = append(policies, policy)
		}
	}

	return &api.ListPolicyMessage{
		Policies: policies,
	}, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *mockClient) GetPolicy(_ context.Context, find *api.PolicyFindMessage) (*api.PolicyMessage, error) {
	name := getPolicyRequestName(find)
	policy, ok := c.policyMap[name]
	if !ok {
		return nil, errors.Errorf("Cannot found policy %s", name)
	}

	return policy, nil
}

// UpsertPolicy creates or updates the policy.
func (c *mockClient) UpsertPolicy(_ context.Context, find *api.PolicyFindMessage, patch *api.PolicyPatchMessage) (*api.PolicyMessage, error) {
	name := getPolicyRequestName(find)
	policy, existed := c.policyMap[name]

	if !existed {
		policy = &api.PolicyMessage{
			Name:  name,
			State: api.Active,
		}
	}

	switch patch.Type {
	case api.PolicyTypeAccessControl:
		if !existed {
			if patch.AccessControlPolicy == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.AccessControlPolicy; v != nil {
			policy.AccessControlPolicy = v
		}
	case api.PolicyTypeBackupPlan:
		if !existed {
			if patch.BackupPlanPolicy == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.BackupPlanPolicy; v != nil {
			policy.BackupPlanPolicy = v
		}
	case api.PolicyTypeDeploymentApproval:
		if !existed {
			if patch.DeploymentApprovalPolicy == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.DeploymentApprovalPolicy; v != nil {
			policy.DeploymentApprovalPolicy = v
		}
	case api.PolicyTypeSQLReview:
		if !existed {
			if patch.SQLReviewPolicy == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.SQLReviewPolicy; v != nil {
			policy.SQLReviewPolicy = v
		}
	case api.PolicyTypeSensitiveData:
		if !existed {
			if patch.SensitiveDataPolicy == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.SensitiveDataPolicy; v != nil {
			policy.SensitiveDataPolicy = v
		}
	default:
		return nil, errors.Errorf("invalid policy type %v", patch.Type)
	}

	if v := patch.InheritFromParent; v != nil {
		policy.InheritFromParent = *v
	}

	c.policyMap[name] = policy

	return policy, nil
}

// DeletePolicy deletes the policy.
func (c *mockClient) DeletePolicy(_ context.Context, find *api.PolicyFindMessage) error {
	name := getPolicyRequestName(find)
	delete(c.roleMap, name)
	return nil
}

func getPolicyRequestName(find *api.PolicyFindMessage) string {
	paths := []string{}
	if v := find.ProjectID; v != nil {
		paths = append(paths, fmt.Sprintf("projects/%s", *v))
	}
	if v := find.EnvironmentID; v != nil {
		paths = append(paths, fmt.Sprintf("environments/%s", *v))
	}
	if v := find.InstanceID; v != nil {
		paths = append(paths, fmt.Sprintf("instances/%s", *v))
	}
	if v := find.DatabaseName; v != nil {
		paths = append(paths, fmt.Sprintf("databases/%s", *v))
	}

	paths = append(paths, "policies")

	name := strings.Join(paths, "/")
	if v := find.Type; v != nil {
		name = fmt.Sprintf("%s/%s", name, *v)
	}

	return name
}

// GetDatabase gets the database by environment resource id, instance resource id and the database name.
func (*mockClient) GetDatabase(_ context.Context, _ *api.DatabaseFindMessage) (*api.DatabaseMessage, error) {
	return nil, errors.Errorf("GetDatabase is not implemented")
}

// GetProject gets the project by resource id.
func (c *mockClient) GetProject(_ context.Context, projectID string, _ bool) (*api.ProjectMessage, error) {
	name := fmt.Sprintf("projects/%s", projectID)
	proj, ok := c.projectMap[name]
	if !ok {
		return nil, errors.Errorf("Cannot found project %s", name)
	}

	return proj, nil
}

// ListProject list the projects.
func (c *mockClient) ListProject(_ context.Context, showDeleted bool) (*api.ListProjectMessage, error) {
	projects := make([]*api.ProjectMessage, 0)
	for _, proj := range c.projectMap {
		if proj.State == api.Deleted && !showDeleted {
			continue
		}
		projects = append(projects, proj)
	}

	return &api.ListProjectMessage{
		Projects: projects,
	}, nil
}

// CreateProject creates the project.
func (c *mockClient) CreateProject(_ context.Context, projectID string, project *api.ProjectMessage) (*api.ProjectMessage, error) {
	proj := &api.ProjectMessage{
		Name:           fmt.Sprintf("projects/%s", projectID),
		State:          api.Active,
		Title:          project.Title,
		Key:            project.Key,
		Workflow:       project.Workflow,
		Visibility:     project.Visibility,
		TenantMode:     project.TenantMode,
		DBNameTemplate: project.DBNameTemplate,
		SchemaVersion:  project.SchemaVersion,
		SchemaChange:   project.SchemaChange,
	}

	c.projectMap[proj.Name] = proj
	return proj, nil
}

// UpdateProject updates the project.
func (c *mockClient) UpdateProject(ctx context.Context, projectID string, patch *api.ProjectPatchMessage) (*api.ProjectMessage, error) {
	proj, err := c.GetProject(ctx, projectID, false)
	if err != nil {
		return nil, err
	}

	if v := patch.Title; v != nil {
		proj.Title = *v
	}
	if v := patch.Key; v != nil {
		proj.Key = *v
	}
	if v := patch.Workflow; v != nil {
		proj.Workflow = *v
	}
	if v := patch.TenantMode; v != nil {
		proj.TenantMode = *v
	}
	if v := patch.DBNameTemplate; v != nil {
		proj.DBNameTemplate = *v
	}
	if v := patch.SchemaChange; v != nil {
		proj.SchemaChange = *v
	}

	c.projectMap[proj.Name] = proj
	return proj, nil
}

// DeleteProject deletes the project.
func (c *mockClient) DeleteProject(ctx context.Context, projectID string) error {
	proj, err := c.GetProject(ctx, projectID, false)
	if err != nil {
		return err
	}

	proj.State = api.Deleted
	c.projectMap[proj.Name] = proj

	return nil
}

// UndeleteProject undeletes the project.
func (c *mockClient) UndeleteProject(ctx context.Context, projectID string) (*api.ProjectMessage, error) {
	proj, err := c.GetProject(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	proj.State = api.Active
	c.projectMap[proj.Name] = proj

	return proj, nil
}
