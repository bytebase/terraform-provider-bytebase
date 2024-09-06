package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"
)

var environmentMap map[string]*api.EnvironmentMessage
var instanceMap map[string]*api.InstanceMessage
var policyMap map[string]*api.PolicyMessage
var projectMap map[string]*api.ProjectMessage
var databaseMap map[string]*api.DatabaseMessage

func init() {
	environmentMap = map[string]*api.EnvironmentMessage{}
	instanceMap = map[string]*api.InstanceMessage{}
	policyMap = map[string]*api.PolicyMessage{}
	projectMap = map[string]*api.ProjectMessage{}
	databaseMap = map[string]*api.DatabaseMessage{}
}

type mockClient struct {
	environmentMap map[string]*api.EnvironmentMessage
	instanceMap    map[string]*api.InstanceMessage
	policyMap      map[string]*api.PolicyMessage
	projectMap     map[string]*api.ProjectMessage
	databaseMap    map[string]*api.DatabaseMessage
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		environmentMap: environmentMap,
		instanceMap:    instanceMap,
		policyMap:      policyMap,
		projectMap:     projectMap,
		databaseMap:    databaseMap,
	}, nil
}

// Login will login the user and get the response.
func (*mockClient) Login() (*api.AuthResponse, error) {
	return &api.AuthResponse{}, nil
}

// CreateEnvironment creates the environment.
func (c *mockClient) CreateEnvironment(_ context.Context, environmentID string, create *api.EnvironmentMessage) (*api.EnvironmentMessage, error) {
	env := &api.EnvironmentMessage{
		Name:  fmt.Sprintf("%s%s", EnvironmentNamePrefix, environmentID),
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
func (c *mockClient) GetEnvironment(_ context.Context, environmentName string) (*api.EnvironmentMessage, error) {
	env, ok := c.environmentMap[environmentName]
	if !ok {
		return nil, errors.Errorf("Cannot found environment %s", environmentName)
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
func (c *mockClient) UpdateEnvironment(ctx context.Context, patch *api.EnvironmentPatchMessage) (*api.EnvironmentMessage, error) {
	env, err := c.GetEnvironment(ctx, patch.Name)
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
func (c *mockClient) DeleteEnvironment(ctx context.Context, environmentName string) error {
	env, err := c.GetEnvironment(ctx, environmentName)
	if err != nil {
		return err
	}

	env.State = api.Deleted
	c.environmentMap[env.Name] = env
	return nil
}

// UndeleteEnvironment undeletes the environment.
func (c *mockClient) UndeleteEnvironment(ctx context.Context, environmentName string) (*api.EnvironmentMessage, error) {
	env, err := c.GetEnvironment(ctx, environmentName)
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
	for _, ins := range c.instanceMap {
		if ins.State == api.Deleted && !find.ShowDeleted {
			continue
		}
		instances = append(instances, ins)
	}

	return &api.ListInstanceMessage{
		Instances: instances,
	}, nil
}

// GetInstance gets the instance by id.
func (c *mockClient) GetInstance(_ context.Context, instanceName string) (*api.InstanceMessage, error) {
	ins, ok := c.instanceMap[instanceName]
	if !ok {
		return nil, errors.Errorf("Cannot found instance %s", instanceName)
	}

	return ins, nil
}

// CreateInstance creates the instance.
func (c *mockClient) CreateInstance(_ context.Context, instanceID string, instance *api.InstanceMessage) (*api.InstanceMessage, error) {
	ins := &api.InstanceMessage{
		Name:         fmt.Sprintf("%s%s", InstanceNamePrefix, instanceID),
		State:        api.Active,
		Title:        instance.Title,
		Engine:       instance.Engine,
		ExternalLink: instance.ExternalLink,
		DataSources:  instance.DataSources,
		Environment:  instance.Environment,
	}

	envID, err := GetEnvironmentID(ins.Environment)
	if err != nil {
		return nil, err
	}

	database := &api.DatabaseMessage{
		Name:      fmt.Sprintf("%s/%sdefault", ins.Name, DatabaseIDPrefix),
		SyncState: api.Active,
		Labels: map[string]string{
			"bb.environment": envID,
		},
	}

	c.instanceMap[ins.Name] = ins
	c.databaseMap[database.Name] = database
	return ins, nil
}

// UpdateInstance updates the instance.
func (c *mockClient) UpdateInstance(ctx context.Context, patch *api.InstancePatchMessage) (*api.InstanceMessage, error) {
	ins, err := c.GetInstance(ctx, patch.Name)
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
func (c *mockClient) DeleteInstance(ctx context.Context, instanceName string) error {
	ins, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return err
	}

	ins.State = api.Deleted
	c.instanceMap[ins.Name] = ins

	return nil
}

// UndeleteInstance undeletes the instance.
func (c *mockClient) UndeleteInstance(ctx context.Context, instanceName string) (*api.InstanceMessage, error) {
	ins, err := c.GetInstance(ctx, instanceName)
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

// ListPolicies lists policies in a specific resource.
func (c *mockClient) ListPolicies(_ context.Context, find *api.PolicyFindMessage) (*api.ListPolicyMessage, error) {
	policies := make([]*api.PolicyMessage, 0)
	for _, policy := range c.policyMap {
		if find.Parent == "" || strings.HasPrefix(policy.Name, find.Parent) {
			policies = append(policies, policy)
		}
	}

	return &api.ListPolicyMessage{
		Policies: policies,
	}, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *mockClient) GetPolicy(_ context.Context, policyName string) (*api.PolicyMessage, error) {
	policy, ok := c.policyMap[policyName]
	if !ok {
		return nil, errors.Errorf("Cannot found policy %s", policyName)
	}

	return policy, nil
}

// UpsertPolicy creates or updates the policy.
func (c *mockClient) UpsertPolicy(_ context.Context, patch *api.PolicyPatchMessage) (*api.PolicyMessage, error) {
	_, policyType, err := GetPolicyParentAndType(patch.Name)
	if err != nil {
		return nil, err
	}

	policy, existed := c.policyMap[patch.Name]

	if !existed {
		policy = &api.PolicyMessage{
			Name:    patch.Name,
			Type:    policyType,
			Enforce: true,
		}
	}

	switch policyType {
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
		return nil, errors.Errorf("invalid policy type %v", policyType)
	}

	if v := patch.InheritFromParent; v != nil {
		policy.InheritFromParent = *v
	}

	c.policyMap[policy.Name] = policy

	return policy, nil
}

// DeletePolicy deletes the policy.
func (c *mockClient) DeletePolicy(_ context.Context, policyName string) error {
	delete(c.policyMap, policyName)
	return nil
}

// GetDatabase gets the database by instance resource id and the database name.
func (c *mockClient) GetDatabase(_ context.Context, databaseName string) (*api.DatabaseMessage, error) {
	db, ok := c.databaseMap[databaseName]
	if !ok {
		return nil, errors.Errorf("Cannot found database %s", databaseName)
	}

	return db, nil
}

// ListDatabase list the databases.
func (c *mockClient) ListDatabase(_ context.Context, find *api.DatabaseFindMessage) (*api.ListDatabaseMessage, error) {
	projectID := "-"
	if v := find.Filter; v != nil && strings.HasPrefix(*v, "project == ") {
		projectID = strings.Split(*v, "project == ")[1]
	}
	databases := make([]*api.DatabaseMessage, 0)
	for _, db := range c.databaseMap {
		if projectID != "-" && fmt.Sprintf(`"%s"`, db.Project) != projectID {
			continue
		}
		if find.InstanceID != "-" && !strings.HasPrefix(db.Name, fmt.Sprintf("%s%s", InstanceNamePrefix, find.InstanceID)) {
			continue
		}
		databases = append(databases, db)
	}

	return &api.ListDatabaseMessage{
		Databases: databases,
	}, nil
}

// UpdateDatabase patches the database.
func (c *mockClient) UpdateDatabase(ctx context.Context, patch *api.DatabasePatchMessage) (*api.DatabaseMessage, error) {
	db, err := c.GetDatabase(ctx, patch.Name)
	if err != nil {
		return nil, err
	}
	if v := patch.Project; v != nil {
		db.Project = *v
	}
	if v := patch.Labels; v != nil {
		db.Labels = *v
	}
	c.databaseMap[db.Name] = db
	return db, nil
}

// GetProject gets the project by resource id.
func (c *mockClient) GetProject(_ context.Context, projectName string) (*api.ProjectMessage, error) {
	proj, ok := c.projectMap[projectName]
	if !ok {
		return nil, errors.Errorf("Cannot found project %s", projectName)
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
		Name:     fmt.Sprintf("%s%s", ProjectNamePrefix, projectID),
		State:    api.Active,
		Title:    project.Title,
		Key:      project.Key,
		Workflow: api.ProjectWorkflowUI,
	}

	c.projectMap[proj.Name] = proj
	return proj, nil
}

// UpdateProject updates the project.
func (c *mockClient) UpdateProject(ctx context.Context, patch *api.ProjectPatchMessage) (*api.ProjectMessage, error) {
	proj, err := c.GetProject(ctx, patch.Name)
	if err != nil {
		return nil, err
	}

	if v := patch.Title; v != nil {
		proj.Title = *v
	}
	if v := patch.Key; v != nil {
		proj.Key = *v
	}

	c.projectMap[proj.Name] = proj
	return proj, nil
}

// DeleteProject deletes the project.
func (c *mockClient) DeleteProject(ctx context.Context, projectName string) error {
	proj, err := c.GetProject(ctx, projectName)
	if err != nil {
		return err
	}

	proj.State = api.Deleted
	c.projectMap[proj.Name] = proj

	return nil
}

// UndeleteProject undeletes the project.
func (c *mockClient) UndeleteProject(ctx context.Context, projectName string) (*api.ProjectMessage, error) {
	proj, err := c.GetProject(ctx, projectName)
	if err != nil {
		return nil, err
	}

	proj.State = api.Active
	c.projectMap[proj.Name] = proj

	return proj, nil
}
