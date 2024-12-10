package internal

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var environmentMap map[string]*v1pb.Environment
var instanceMap map[string]*v1pb.Instance
var policyMap map[string]*v1pb.Policy
var projectMap map[string]*v1pb.Project
var databaseMap map[string]*v1pb.Database
var settingMap map[string]*v1pb.Setting

func init() {
	environmentMap = map[string]*v1pb.Environment{}
	instanceMap = map[string]*v1pb.Instance{}
	policyMap = map[string]*v1pb.Policy{}
	projectMap = map[string]*v1pb.Project{}
	databaseMap = map[string]*v1pb.Database{}
	settingMap = map[string]*v1pb.Setting{}
}

type mockClient struct {
	environmentMap map[string]*v1pb.Environment
	instanceMap    map[string]*v1pb.Instance
	policyMap      map[string]*v1pb.Policy
	projectMap     map[string]*v1pb.Project
	databaseMap    map[string]*v1pb.Database
	settingMap     map[string]*v1pb.Setting
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		environmentMap: environmentMap,
		instanceMap:    instanceMap,
		policyMap:      policyMap,
		projectMap:     projectMap,
		databaseMap:    databaseMap,
		settingMap:     settingMap,
	}, nil
}

// GetCaller returns the API caller.
func (*mockClient) GetCaller() *v1pb.User {
	return &v1pb.User{
		Name:  "users/mock@bytease.com",
		Email: "mock@bytease.com",
	}
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
func (c *mockClient) ListPolicies(_ context.Context, parent string) (*v1pb.ListPoliciesResponse, error) {
	policies := make([]*v1pb.Policy, 0)
	for _, policy := range c.policyMap {
		if parent == "" || strings.HasPrefix(policy.Name, parent) {
			policies = append(policies, policy)
		}
	}

	return &v1pb.ListPoliciesResponse{
		Policies: policies,
	}, nil
}

// GetPolicy gets a policy in a specific resource.
func (c *mockClient) GetPolicy(_ context.Context, policyName string) (*v1pb.Policy, error) {
	policy, ok := c.policyMap[policyName]
	if !ok {
		return nil, errors.Errorf("Cannot found policy %s", policyName)
	}

	return policy, nil
}

// UpsertPolicy creates or updates the policy.
func (c *mockClient) UpsertPolicy(_ context.Context, patch *v1pb.Policy, updateMasks []string) (*v1pb.Policy, error) {
	_, policyType, err := GetPolicyParentAndType(patch.Name)
	if err != nil {
		return nil, err
	}

	policy, existed := c.policyMap[patch.Name]

	if !existed {
		policy = &v1pb.Policy{
			Name:    patch.Name,
			Type:    policyType,
			Enforce: true,
		}
	}

	switch policyType {
	case v1pb.PolicyType_MASKING:
		if !existed {
			if patch.GetMaskingPolicy() == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.GetMaskingPolicy(); v != nil {
			policy.Policy = &v1pb.Policy_MaskingPolicy{
				MaskingPolicy: v,
			}
		}
	case v1pb.PolicyType_MASKING_EXCEPTION:
		if !existed {
			if patch.GetMaskingExceptionPolicy() == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.GetMaskingExceptionPolicy(); v != nil {
			policy.Policy = &v1pb.Policy_MaskingExceptionPolicy{
				MaskingExceptionPolicy: v,
			}
		}
	default:
		return nil, errors.Errorf("invalid policy type %v", policyType)
	}

	if slices.Contains(updateMasks, "inherit_from_parent") {
		policy.InheritFromParent = patch.InheritFromParent
	}
	if slices.Contains(updateMasks, "enforce") {
		policy.Enforce = patch.Enforce
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

// GetProjectIAMPolicy gets the project IAM policy by project full name.
func (*mockClient) GetProjectIAMPolicy(_ context.Context, _ string) (*v1pb.IamPolicy, error) {
	return &v1pb.IamPolicy{}, nil
}

// SetProjectIAMPolicy sets the project IAM policy.
func (*mockClient) SetProjectIAMPolicy(_ context.Context, _ string, _ *v1pb.IamPolicy) (*v1pb.IamPolicy, error) {
	return &v1pb.IamPolicy{}, nil
}

// ListSettings lists all settings.
func (c *mockClient) ListSettings(_ context.Context) (*v1pb.ListSettingsResponse, error) {
	settings := make([]*v1pb.Setting, 0)
	for _, setting := range c.settingMap {
		settings = append(settings, setting)
	}

	return &v1pb.ListSettingsResponse{
		Settings: settings,
	}, nil
}

// ListSettings lists all settings.
func (c *mockClient) GetSetting(_ context.Context, settingName string) (*v1pb.Setting, error) {
	setting, ok := c.settingMap[settingName]
	if !ok {
		return nil, errors.Errorf("Cannot found setting %s", settingName)
	}

	return setting, nil
}

// UpsertSetting updates or creates the setting.
func (c *mockClient) UpsertSetting(_ context.Context, upsert *v1pb.Setting, _ []string) (*v1pb.Setting, error) {
	setting, ok := c.settingMap[upsert.Name]
	if !ok {
		c.settingMap[upsert.Name] = upsert
	} else {
		setting.Value = upsert.Value
		c.settingMap[upsert.Name] = setting
	}
	return c.settingMap[upsert.Name], nil
}

// ParseExpression parse the expression string.
func (*mockClient) ParseExpression(_ context.Context, _ string) (*v1alpha1.Expr, error) {
	return nil, nil
}

// ListVCSProvider will returns all vcs providers.
func (*mockClient) ListVCSProvider(_ context.Context) (*v1pb.ListVCSProvidersResponse, error) {
	return nil, nil
}

// GetVCSProvider gets the vcs by id.
func (*mockClient) GetVCSProvider(_ context.Context, _ string) (*v1pb.VCSProvider, error) {
	return nil, nil
}

// CreateVCSProvider creates the vcs provider.
func (*mockClient) CreateVCSProvider(_ context.Context, _ string, _ *v1pb.VCSProvider) (*v1pb.VCSProvider, error) {
	return nil, nil
}

// UpdateVCSProvider updates the vcs provider.
func (*mockClient) UpdateVCSProvider(_ context.Context, _ *v1pb.VCSProvider, _ []string) (*v1pb.VCSConnector, error) {
	return nil, nil
}

// DeleteVCSProvider deletes the vcs provider.
func (*mockClient) DeleteVCSProvider(_ context.Context, _ string) error {
	return nil
}

// ListVCSConnector will returns all vcs connector in a project.
func (*mockClient) ListVCSConnector(_ context.Context, _ string) (*v1pb.ListVCSConnectorsResponse, error) {
	return nil, nil
}

// GetVCSConnector gets the vcs connector by id.
func (*mockClient) GetVCSConnector(_ context.Context, _ string) (*v1pb.VCSConnector, error) {
	return nil, nil
}

// CreateVCSConnector creates the vcs connector in a project.
func (*mockClient) CreateVCSConnector(_ context.Context, _, _ string, _ *v1pb.VCSConnector) (*v1pb.VCSConnector, error) {
	return nil, nil
}

// UpdateVCSConnector updates the vcs connector.
func (*mockClient) UpdateVCSConnector(_ context.Context, _ *v1pb.VCSConnector, _ []string) (*v1pb.VCSConnector, error) {
	return nil, nil
}

// DeleteVCSConnector deletes the vcs provider.
func (*mockClient) DeleteVCSConnector(_ context.Context, _ string) error {
	return nil
}

// ListUser list all users.
func (*mockClient) ListUser(_ context.Context, _ bool) (*v1pb.ListUsersResponse, error) {
	return nil, nil
}

// GetUser gets the user by name.
func (*mockClient) GetUser(_ context.Context, _ string) (*v1pb.User, error) {
	return nil, nil
}

// CreateUser creates the user.
func (*mockClient) CreateUser(_ context.Context, _ *v1pb.User) (*v1pb.User, error) {
	return nil, nil
}

// UpdateUser updates the user.
func (*mockClient) UpdateUser(_ context.Context, _ *v1pb.User, _ []string) (*v1pb.User, error) {
	return nil, nil
}

// DeleteUser deletes the user by name.
func (*mockClient) DeleteUser(_ context.Context, _ string) error {
	return nil
}

// UndeleteUser undeletes the user by name.
func (*mockClient) UndeleteUser(_ context.Context, _ string) (*v1pb.User, error) {
	return nil, nil
}
