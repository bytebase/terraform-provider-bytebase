package internal

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

var environmentMap map[string]*v1pb.Environment
var instanceMap map[string]*v1pb.Instance
var policyMap map[string]*api.PolicyMessage
var projectMap map[string]*v1pb.Project
var databaseMap map[string]*v1pb.Database

func init() {
	environmentMap = map[string]*v1pb.Environment{}
	instanceMap = map[string]*v1pb.Instance{}
	policyMap = map[string]*api.PolicyMessage{}
	projectMap = map[string]*v1pb.Project{}
	databaseMap = map[string]*v1pb.Database{}
}

type mockClient struct {
	environmentMap map[string]*v1pb.Environment
	instanceMap    map[string]*v1pb.Instance
	policyMap      map[string]*api.PolicyMessage
	projectMap     map[string]*v1pb.Project
	databaseMap    map[string]*v1pb.Database
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
func (*mockClient) Login() (*v1pb.LoginResponse, error) {
	return &v1pb.LoginResponse{}, nil
}

// CreateEnvironment creates the environment.
func (c *mockClient) CreateEnvironment(_ context.Context, environmentID string, create *v1pb.Environment) (*v1pb.Environment, error) {
	env := &v1pb.Environment{
		Name:  fmt.Sprintf("%s%s", EnvironmentNamePrefix, environmentID),
		Order: create.Order,
		Title: create.Title,
		State: v1pb.State_ACTIVE,
		Tier:  create.Tier,
	}

	if _, ok := c.environmentMap[env.Name]; ok {
		return nil, errors.Errorf("Environment %s already exists", env.Name)
	}

	c.environmentMap[env.Name] = env

	return env, nil
}

// GetEnvironment gets the environment by id.
func (c *mockClient) GetEnvironment(_ context.Context, environmentName string) (*v1pb.Environment, error) {
	env, ok := c.environmentMap[environmentName]
	if !ok {
		return nil, errors.Errorf("Cannot found environment %s", environmentName)
	}

	return env, nil
}

// ListEnvironment finds all environments.
func (c *mockClient) ListEnvironment(_ context.Context, showDeleted bool) (*v1pb.ListEnvironmentsResponse, error) {
	environments := make([]*v1pb.Environment, 0)
	for _, env := range c.environmentMap {
		if env.State == v1pb.State_DELETED && !showDeleted {
			continue
		}
		environments = append(environments, env)
	}

	return &v1pb.ListEnvironmentsResponse{
		Environments: environments,
	}, nil
}

// UpdateEnvironment updates the environment.
func (c *mockClient) UpdateEnvironment(ctx context.Context, patch *v1pb.Environment, updateMasks []string) (*v1pb.Environment, error) {
	env, err := c.GetEnvironment(ctx, patch.Name)
	if err != nil {
		return nil, err
	}

	if slices.Contains(updateMasks, "title") {
		env.Title = patch.Title
	}
	if slices.Contains(updateMasks, "order") {
		env.Order = patch.Order
	}
	if slices.Contains(updateMasks, "tier") {
		env.Tier = patch.Tier
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

	env.State = v1pb.State_DELETED
	c.environmentMap[env.Name] = env
	return nil
}

// UndeleteEnvironment undeletes the environment.
func (c *mockClient) UndeleteEnvironment(ctx context.Context, environmentName string) (*v1pb.Environment, error) {
	env, err := c.GetEnvironment(ctx, environmentName)
	if err != nil {
		return nil, err
	}

	env.State = v1pb.State_ACTIVE
	c.environmentMap[env.Name] = env
	return env, nil
}

// ListInstance will return instances in environment.
func (c *mockClient) ListInstance(_ context.Context, showDeleted bool) (*v1pb.ListInstancesResponse, error) {
	instances := make([]*v1pb.Instance, 0)
	for _, ins := range c.instanceMap {
		if ins.State == v1pb.State_DELETED && !showDeleted {
			continue
		}
		instances = append(instances, ins)
	}

	return &v1pb.ListInstancesResponse{
		Instances: instances,
	}, nil
}

// GetInstance gets the instance by id.
func (c *mockClient) GetInstance(_ context.Context, instanceName string) (*v1pb.Instance, error) {
	ins, ok := c.instanceMap[instanceName]
	if !ok {
		return nil, errors.Errorf("Cannot found instance %s", instanceName)
	}

	return ins, nil
}

// CreateInstance creates the instance.
func (c *mockClient) CreateInstance(_ context.Context, instanceID string, instance *v1pb.Instance) (*v1pb.Instance, error) {
	ins := &v1pb.Instance{
		Name:         fmt.Sprintf("%s%s", InstanceNamePrefix, instanceID),
		State:        v1pb.State_ACTIVE,
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

	database := &v1pb.Database{
		Name:      fmt.Sprintf("%s/%sdefault", ins.Name, DatabaseIDPrefix),
		SyncState: v1pb.State_ACTIVE,
		Labels: map[string]string{
			"bb.environment": envID,
		},
	}

	c.instanceMap[ins.Name] = ins
	c.databaseMap[database.Name] = database
	return ins, nil
}

// UpdateInstance updates the instance.
func (c *mockClient) UpdateInstance(ctx context.Context, patch *v1pb.Instance, updateMasks []string) (*v1pb.Instance, error) {
	ins, err := c.GetInstance(ctx, patch.Name)
	if err != nil {
		return nil, err
	}

	if slices.Contains(updateMasks, "title") {
		ins.Title = patch.Title
	}
	if slices.Contains(updateMasks, "external_link") {
		ins.ExternalLink = patch.ExternalLink
	}
	if slices.Contains(updateMasks, "data_sources") {
		ins.DataSources = patch.DataSources
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

	ins.State = v1pb.State_DELETED
	c.instanceMap[ins.Name] = ins

	return nil
}

// UndeleteInstance undeletes the instance.
func (c *mockClient) UndeleteInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error) {
	ins, err := c.GetInstance(ctx, instanceName)
	if err != nil {
		return nil, err
	}

	ins.State = v1pb.State_ACTIVE
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
func (c *mockClient) GetDatabase(_ context.Context, databaseName string) (*v1pb.Database, error) {
	db, ok := c.databaseMap[databaseName]
	if !ok {
		return nil, errors.Errorf("Cannot found database %s", databaseName)
	}

	return db, nil
}

// ListDatabase list the databases.
func (c *mockClient) ListDatabase(_ context.Context, instaceID, filter string) (*v1pb.ListDatabasesResponse, error) {
	projectID := "-"
	if strings.HasPrefix(filter, "project == ") {
		projectID = strings.Split(filter, "project == ")[1]
	}
	databases := make([]*v1pb.Database, 0)
	for _, db := range c.databaseMap {
		if projectID != "-" && fmt.Sprintf(`"%s"`, db.Project) != projectID {
			continue
		}
		if instaceID != "-" && !strings.HasPrefix(db.Name, fmt.Sprintf("%s%s", InstanceNamePrefix, instaceID)) {
			continue
		}
		databases = append(databases, db)
	}

	return &v1pb.ListDatabasesResponse{
		Databases: databases,
	}, nil
}

// UpdateDatabase patches the database.
func (c *mockClient) UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error) {
	db, err := c.GetDatabase(ctx, patch.Name)
	if err != nil {
		return nil, err
	}
	if slices.Contains(updateMasks, "project") {
		db.Project = patch.Project
	}
	if slices.Contains(updateMasks, "labels") {
		db.Labels = patch.Labels
	}
	c.databaseMap[db.Name] = db
	return db, nil
}

// GetProject gets the project by resource id.
func (c *mockClient) GetProject(_ context.Context, projectName string) (*v1pb.Project, error) {
	proj, ok := c.projectMap[projectName]
	if !ok {
		return nil, errors.Errorf("Cannot found project %s", projectName)
	}

	return proj, nil
}

// ListProject list the projects.
func (c *mockClient) ListProject(_ context.Context, showDeleted bool) (*v1pb.ListProjectsResponse, error) {
	projects := make([]*v1pb.Project, 0)
	for _, proj := range c.projectMap {
		if proj.State == v1pb.State_DELETED && !showDeleted {
			continue
		}
		projects = append(projects, proj)
	}

	return &v1pb.ListProjectsResponse{
		Projects: projects,
	}, nil
}

// CreateProject creates the project.
func (c *mockClient) CreateProject(_ context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error) {
	proj := &v1pb.Project{
		Name:     fmt.Sprintf("%s%s", ProjectNamePrefix, projectID),
		State:    v1pb.State_ACTIVE,
		Title:    project.Title,
		Key:      project.Key,
		Workflow: v1pb.Workflow_UI,
	}

	c.projectMap[proj.Name] = proj
	return proj, nil
}

// UpdateProject updates the project.
func (c *mockClient) UpdateProject(ctx context.Context, patch *v1pb.Project, updateMasks []string) (*v1pb.Project, error) {
	proj, err := c.GetProject(ctx, patch.Name)
	if err != nil {
		return nil, err
	}

	if slices.Contains(updateMasks, "title") {
		proj.Title = patch.Title
	}
	if slices.Contains(updateMasks, "key") {
		proj.Key = patch.Key
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

	proj.State = v1pb.State_DELETED
	c.projectMap[proj.Name] = proj

	return nil
}

// UndeleteProject undeletes the project.
func (c *mockClient) UndeleteProject(ctx context.Context, projectName string) (*v1pb.Project, error) {
	proj, err := c.GetProject(ctx, projectName)
	if err != nil {
		return nil, err
	}

	proj.State = v1pb.State_ACTIVE
	c.projectMap[proj.Name] = proj

	return proj, nil
}
