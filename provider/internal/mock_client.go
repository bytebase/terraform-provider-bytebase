package internal

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bytebase/terraform-provider-bytebase/api"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var instanceMap map[string]*v1pb.Instance
var policyMap map[string]*v1pb.Policy
var projectMap map[string]*v1pb.Project
var projectIAMMap map[string]*v1pb.IamPolicy
var databaseMap map[string]*v1pb.Database
var databaseCatalogMap map[string]*v1pb.DatabaseCatalog
var settingMap map[string]*v1pb.Setting
var userMap map[string]*v1pb.User
var roleMap map[string]*v1pb.Role
var groupMap map[string]*v1pb.Group
var reviewConfigMap map[string]*v1pb.ReviewConfig
var riskMap map[string]*v1pb.Risk
var databaseGroupMap map[string]*v1pb.DatabaseGroup
var workspaceIAMPolicy *v1pb.IamPolicy

func init() {
	instanceMap = map[string]*v1pb.Instance{}
	policyMap = map[string]*v1pb.Policy{}
	projectMap = map[string]*v1pb.Project{}
	projectIAMMap = map[string]*v1pb.IamPolicy{}
	databaseMap = map[string]*v1pb.Database{}
	databaseCatalogMap = map[string]*v1pb.DatabaseCatalog{}
	settingMap = map[string]*v1pb.Setting{}
	userMap = map[string]*v1pb.User{}
	roleMap = map[string]*v1pb.Role{}
	groupMap = map[string]*v1pb.Group{}
	reviewConfigMap = map[string]*v1pb.ReviewConfig{}
	riskMap = map[string]*v1pb.Risk{}
	databaseGroupMap = map[string]*v1pb.DatabaseGroup{}
	workspaceIAMPolicy = &v1pb.IamPolicy{}

	// Initialize environment setting with an empty list
	settingMap[fmt.Sprintf("%s%s", SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String())] = &v1pb.Setting{
		Name: fmt.Sprintf("%s%s", SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()),
		Value: &v1pb.Value{
			Value: &v1pb.Value_EnvironmentSetting{
				EnvironmentSetting: &v1pb.EnvironmentSetting{
					Environments: []*v1pb.EnvironmentSetting_Environment{},
				},
			},
		},
	}
}

type mockClient struct {
	instanceMap        map[string]*v1pb.Instance
	policyMap          map[string]*v1pb.Policy
	projectMap         map[string]*v1pb.Project
	projectIAMMap      map[string]*v1pb.IamPolicy
	databaseMap        map[string]*v1pb.Database
	databaseCatalogMap map[string]*v1pb.DatabaseCatalog
	settingMap         map[string]*v1pb.Setting
	userMap            map[string]*v1pb.User
	roleMap            map[string]*v1pb.Role
	groupMap           map[string]*v1pb.Group
	reviewConfigMap    map[string]*v1pb.ReviewConfig
	riskMap            map[string]*v1pb.Risk
	databaseGroupMap   map[string]*v1pb.DatabaseGroup
}

// newMockClient returns the new Bytebase API mock client.
func newMockClient(_, _, _ string) (api.Client, error) {
	return &mockClient{
		instanceMap:        instanceMap,
		policyMap:          policyMap,
		projectMap:         projectMap,
		projectIAMMap:      projectIAMMap,
		databaseMap:        databaseMap,
		databaseCatalogMap: databaseCatalogMap,
		settingMap:         settingMap,
		userMap:            userMap,
		roleMap:            roleMap,
		groupMap:           groupMap,
		reviewConfigMap:    reviewConfigMap,
		riskMap:            riskMap,
		databaseGroupMap:   databaseGroupMap,
	}, nil
}

// ListInstance will return instances in environment.
func (c *mockClient) ListInstance(_ context.Context, filter *api.InstanceFilter) ([]*v1pb.Instance, error) {
	instances := make([]*v1pb.Instance, 0)
	for _, ins := range c.instanceMap {
		if ins.State == v1pb.State_DELETED && filter.State != v1pb.State_DELETED {
			continue
		}
		instances = append(instances, ins)
	}

	return instances, nil
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

	var envID string
	var err error
	if ins.Environment != nil {
		envID, err = GetEnvironmentID(*ins.Environment)
	} else {
		err = errors.New("instance environment is nil")
	}
	if err != nil {
		return nil, err
	}

	// Create default database
	defaultDb := &v1pb.Database{
		Name:  fmt.Sprintf("%s/%sdefault", ins.Name, DatabaseIDPrefix),
		State: v1pb.State_ACTIVE,
		Labels: map[string]string{
			"bb.environment": envID,
		},
	}

	// Also create test databases that will be used in tests
	testDb := &v1pb.Database{
		Name:  fmt.Sprintf("%s/%stest-database", ins.Name, DatabaseIDPrefix),
		State: v1pb.State_ACTIVE,
		Labels: map[string]string{
			"bb.environment": envID,
		},
	}

	testDbLabels := &v1pb.Database{
		Name:  fmt.Sprintf("%s/%stest-database-labels", ins.Name, DatabaseIDPrefix),
		State: v1pb.State_ACTIVE,
		Labels: map[string]string{
			"bb.environment": envID,
		},
	}

	c.instanceMap[ins.Name] = ins
	c.databaseMap[defaultDb.Name] = defaultDb
	c.databaseMap[testDb.Name] = testDb
	c.databaseMap[testDbLabels.Name] = testDbLabels

	// Also create empty catalogs for the databases
	c.databaseCatalogMap[defaultDb.Name] = &v1pb.DatabaseCatalog{
		Name: defaultDb.Name,
	}
	c.databaseCatalogMap[testDb.Name] = &v1pb.DatabaseCatalog{
		Name: testDb.Name,
	}
	c.databaseCatalogMap[testDbLabels.Name] = &v1pb.DatabaseCatalog{
		Name: testDbLabels.Name,
	}
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
	if slices.Contains(updateMasks, "sync_interval") {
		ins.SyncInterval = patch.SyncInterval
	}
	if slices.Contains(updateMasks, "maximum_connections") {
		ins.MaximumConnections = patch.MaximumConnections
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
	case v1pb.PolicyType_MASKING_RULE:
		if !existed {
			if patch.GetMaskingRulePolicy() == nil {
				return nil, errors.Errorf("payload is required to create the policy")
			}
		}
		if v := patch.GetMaskingRulePolicy(); v != nil {
			policy.Policy = &v1pb.Policy_MaskingRulePolicy{
				MaskingRulePolicy: v,
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
func (c *mockClient) ListDatabase(_ context.Context, instaceID string, filter *api.DatabaseFilter, _ bool) ([]*v1pb.Database, error) {
	projectID := "-"
	if filter.Project != "" {
		projectID = filter.Project
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

	return databases, nil
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

// BatchUpdateDatabases batch updates databases.
func (c *mockClient) BatchUpdateDatabases(ctx context.Context, request *v1pb.BatchUpdateDatabasesRequest) (*v1pb.BatchUpdateDatabasesResponse, error) {
	for _, req := range request.Requests {
		db, err := c.GetDatabase(ctx, req.Database.Name)
		if err != nil {
			return nil, err
		}
		if slices.Contains(req.UpdateMask.Paths, "project") {
			db.Project = req.Database.Project
		}
		c.databaseMap[db.Name] = db
	}

	return &v1pb.BatchUpdateDatabasesResponse{}, nil
}

// GetDatabaseCatalog gets the database catalog by the database full name.
func (c *mockClient) GetDatabaseCatalog(_ context.Context, databaseName string) (*v1pb.DatabaseCatalog, error) {
	db, ok := c.databaseCatalogMap[databaseName]
	if !ok {
		return nil, errors.Errorf("Cannot found database catalog %s", databaseName)
	}

	return db, nil
}

// UpdateDatabaseCatalog patches the database catalog.
func (c *mockClient) UpdateDatabaseCatalog(_ context.Context, patch *v1pb.DatabaseCatalog) (*v1pb.DatabaseCatalog, error) {
	c.databaseCatalogMap[patch.Name] = patch
	return patch, nil
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
func (c *mockClient) ListProject(_ context.Context, filter *api.ProjectFilter) ([]*v1pb.Project, error) {
	projects := make([]*v1pb.Project, 0)
	for _, proj := range c.projectMap {
		if proj.State == v1pb.State_DELETED && filter.State != v1pb.State_DELETED {
			continue
		}
		projects = append(projects, proj)
	}

	return projects, nil
}

// CreateProject creates the project.
func (c *mockClient) CreateProject(_ context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error) {
	proj := &v1pb.Project{
		Name:  fmt.Sprintf("%s%s", ProjectNamePrefix, projectID),
		State: v1pb.State_ACTIVE,
		Title: project.Title,
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
func (c *mockClient) GetProjectIAMPolicy(_ context.Context, projectName string) (*v1pb.IamPolicy, error) {
	iamPolicy, ok := c.projectIAMMap[projectName]
	if !ok {
		return &v1pb.IamPolicy{}, nil
	}
	return iamPolicy, nil
}

// SetProjectIAMPolicy sets the project IAM policy.
func (c *mockClient) SetProjectIAMPolicy(_ context.Context, projectName string, update *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	c.projectIAMMap[projectName] = update.Policy
	return c.projectIAMMap[projectName], nil
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
func (*mockClient) ParseExpression(_ context.Context, expression string) (*v1alpha1.Expr, error) {
	// For mock client, we parse the expression and return a proper structure
	// The real client would parse the expression, but for testing we create a mock response based on the input

	// Parse OR conditions (||)
	if strings.Contains(expression, " || ") {
		conditions := strings.Split(expression, " || ")
		args := make([]*v1alpha1.Expr, 0, len(conditions))

		for i, condition := range conditions {
			// Parse each condition (source == "X" && level == Y)
			parsed := parseCondition(condition, int64(i*10+1))
			if parsed != nil {
				args = append(args, parsed)
			}
		}

		// If we have multiple conditions, wrap them in OR
		if len(args) > 1 {
			return &v1alpha1.Expr{
				Id: 1,
				ExprKind: &v1alpha1.Expr_CallExpr{
					CallExpr: &v1alpha1.Expr_Call{
						Function: "_||_",
						Args:     args,
					},
				},
			}, nil
		} else if len(args) == 1 {
			return args[0], nil
		}
	}

	// Single condition
	return parseCondition(expression, 1), nil
}

func parseCondition(condition string, baseID int64) *v1alpha1.Expr {
	// Parse AND conditions (&&)
	parts := strings.Split(condition, " && ")
	if len(parts) != 2 {
		// Return a simple default expression
		return &v1alpha1.Expr{
			Id: baseID,
			ExprKind: &v1alpha1.Expr_IdentExpr{
				IdentExpr: &v1alpha1.Expr_Ident{
					Name: condition,
				},
			},
		}
	}

	var sourceValue string
	var levelValue int64

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "source ==") {
			// Extract source value
			sourceValue = strings.Trim(strings.TrimPrefix(part, "source =="), ` "`)
		} else if strings.HasPrefix(part, "level ==") {
			// Extract level value
			levelStr := strings.TrimSpace(strings.TrimPrefix(part, "level =="))
			var err error
			levelValue, err = strconv.ParseInt(levelStr, 10, 64)
			if err != nil {
				// Default to 0 if parsing fails
				levelValue = 0
			}
		}
	}

	return &v1alpha1.Expr{
		Id: baseID,
		ExprKind: &v1alpha1.Expr_CallExpr{
			CallExpr: &v1alpha1.Expr_Call{
				Function: "_&&_",
				Args: []*v1alpha1.Expr{
					{
						Id: baseID + 1,
						ExprKind: &v1alpha1.Expr_CallExpr{
							CallExpr: &v1alpha1.Expr_Call{
								Function: "_==_",
								Args: []*v1alpha1.Expr{
									{
										Id: baseID + 2,
										ExprKind: &v1alpha1.Expr_IdentExpr{
											IdentExpr: &v1alpha1.Expr_Ident{
												Name: "source",
											},
										},
									},
									{
										Id: baseID + 3,
										ExprKind: &v1alpha1.Expr_ConstExpr{
											ConstExpr: &v1alpha1.Constant{
												ConstantKind: &v1alpha1.Constant_StringValue{
													StringValue: sourceValue,
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Id: baseID + 4,
						ExprKind: &v1alpha1.Expr_CallExpr{
							CallExpr: &v1alpha1.Expr_Call{
								Function: "_==_",
								Args: []*v1alpha1.Expr{
									{
										Id: baseID + 5,
										ExprKind: &v1alpha1.Expr_IdentExpr{
											IdentExpr: &v1alpha1.Expr_Ident{
												Name: "level",
											},
										},
									},
									{
										Id: baseID + 6,
										ExprKind: &v1alpha1.Expr_ConstExpr{
											ConstExpr: &v1alpha1.Constant{
												ConstantKind: &v1alpha1.Constant_Int64Value{
													Int64Value: levelValue,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ListUser list all users.
func (c *mockClient) ListUser(_ context.Context, filter *api.UserFilter) ([]*v1pb.User, error) {
	users := make([]*v1pb.User, 0)
	for _, user := range c.userMap {
		if user.State == v1pb.State_DELETED && filter.State != v1pb.State_DELETED {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser gets the user by name.
func (c *mockClient) GetUser(_ context.Context, userName string) (*v1pb.User, error) {
	user, ok := c.userMap[userName]
	if !ok {
		return nil, errors.Errorf("Cannot found user %s", userName)
	}

	return user, nil
}

// CreateUser creates the user.
func (c *mockClient) CreateUser(_ context.Context, user *v1pb.User) (*v1pb.User, error) {
	// For service accounts, generate a service key
	if user.UserType == v1pb.UserType_SERVICE_ACCOUNT && user.ServiceKey == "" {
		user.ServiceKey = fmt.Sprintf("bbs_%s_mock_service_key", strings.ReplaceAll(user.Email, "@", "_"))
	}
	c.userMap[user.Name] = user
	return c.userMap[user.Name], nil
}

// UpdateUser updates the user.
func (c *mockClient) UpdateUser(ctx context.Context, user *v1pb.User, updateMasks []string) (*v1pb.User, error) {
	existed, err := c.GetUser(ctx, user.Name)
	if err != nil {
		return nil, err
	}
	if slices.Contains(updateMasks, "email") {
		existed.Email = user.Email
		existed.Name = fmt.Sprintf("%s%s", UserNamePrefix, user.Email)
	}
	if slices.Contains(updateMasks, "title") {
		existed.Title = user.Title
	}
	if slices.Contains(updateMasks, "password") {
		existed.Password = user.Password
	}
	if slices.Contains(updateMasks, "phone") {
		existed.Phone = user.Phone
	}
	c.userMap[user.Name] = existed
	return c.userMap[user.Name], nil
}

// DeleteUser deletes the user by name.
func (c *mockClient) DeleteUser(ctx context.Context, userName string) error {
	user, err := c.GetUser(ctx, userName)
	if err != nil {
		return err
	}

	user.State = v1pb.State_DELETED
	c.userMap[user.Name] = user

	return nil
}

// UndeleteUser undeletes the user by name.
func (c *mockClient) UndeleteUser(ctx context.Context, userName string) (*v1pb.User, error) {
	user, err := c.GetUser(ctx, userName)
	if err != nil {
		return nil, err
	}

	user.State = v1pb.State_ACTIVE
	c.userMap[user.Name] = user

	return c.userMap[user.Name], nil
}

// ListGroup list all groups.
func (c *mockClient) ListGroup(_ context.Context, _ *api.GroupFilter) ([]*v1pb.Group, error) {
	groups := make([]*v1pb.Group, 0)
	for _, group := range c.groupMap {
		groups = append(groups, group)
	}

	return groups, nil
}

// GetGroup gets the group by name.
func (c *mockClient) GetGroup(_ context.Context, name string) (*v1pb.Group, error) {
	group, ok := c.groupMap[name]
	if !ok {
		return nil, errors.Errorf("Cannot found group %s", name)
	}

	return group, nil
}

// CreateGroup creates the group.
func (c *mockClient) CreateGroup(_ context.Context, email string, group *v1pb.Group) (*v1pb.Group, error) {
	groupName := fmt.Sprintf("%s%s", GroupNamePrefix, email)
	group.Name = groupName
	c.groupMap[groupName] = group
	return c.groupMap[groupName], nil
}

// UpdateGroup updates the group.
func (c *mockClient) UpdateGroup(ctx context.Context, group *v1pb.Group, updateMasks []string) (*v1pb.Group, error) {
	existed, err := c.GetGroup(ctx, group.Name)
	if err != nil {
		return nil, err
	}
	if slices.Contains(updateMasks, "description") {
		existed.Description = group.Description
	}
	if slices.Contains(updateMasks, "title") {
		existed.Title = group.Title
	}
	if slices.Contains(updateMasks, "members") {
		existed.Members = group.Members
	}
	c.groupMap[existed.Name] = existed
	return existed, nil
}

// DeleteGroup deletes the group by name.
func (c *mockClient) DeleteGroup(_ context.Context, name string) error {
	delete(c.groupMap, name)
	return nil
}

// GetWorkspaceIAMPolicy gets the workspace IAM policy.
func (*mockClient) GetWorkspaceIAMPolicy(_ context.Context) (*v1pb.IamPolicy, error) {
	return workspaceIAMPolicy, nil
}

// SetWorkspaceIAMPolicy sets the workspace IAM policy.
func (*mockClient) SetWorkspaceIAMPolicy(_ context.Context, update *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error) {
	if v := update.Policy; v != nil {
		workspaceIAMPolicy = v
	}
	return workspaceIAMPolicy, nil
}

// ListRole will returns all roles.
func (c *mockClient) ListRole(_ context.Context) (*v1pb.ListRolesResponse, error) {
	roles := make([]*v1pb.Role, 0)
	for _, role := range c.roleMap {
		roles = append(roles, role)
	}

	return &v1pb.ListRolesResponse{
		Roles: roles,
	}, nil
}

// GetRole gets the role by full name.
func (c *mockClient) GetRole(_ context.Context, roleName string) (*v1pb.Role, error) {
	role, ok := c.roleMap[roleName]
	if !ok {
		return nil, errors.Errorf("Cannot found role %s", roleName)
	}

	return role, nil
}

// CreateRole creates the role.
func (c *mockClient) CreateRole(_ context.Context, roleID string, role *v1pb.Role) (*v1pb.Role, error) {
	roleName := fmt.Sprintf("%s%s", RoleNamePrefix, roleID)
	role.Name = roleName
	c.roleMap[roleName] = role
	return role, nil
}

// UpdateRole updates the role.
func (c *mockClient) UpdateRole(ctx context.Context, role *v1pb.Role, updateMasks []string) (*v1pb.Role, error) {
	existed, err := c.GetRole(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	if slices.Contains(updateMasks, "title") {
		existed.Title = role.Title
	}
	if slices.Contains(updateMasks, "description") {
		existed.Description = role.Description
	}
	if slices.Contains(updateMasks, "permissions") {
		existed.Permissions = role.Permissions
	}
	c.roleMap[existed.Name] = existed
	return c.roleMap[existed.Name], nil
}

// DeleteRole deletes the role by name.
func (c *mockClient) DeleteRole(_ context.Context, roleName string) error {
	delete(c.roleMap, roleName)
	return nil
}

// ListReviewConfig will return review configs.
func (c *mockClient) ListReviewConfig(_ context.Context) (*v1pb.ListReviewConfigsResponse, error) {
	configs := make([]*v1pb.ReviewConfig, 0)
	for _, config := range c.reviewConfigMap {
		configs = append(configs, config)
	}
	return &v1pb.ListReviewConfigsResponse{
		ReviewConfigs: configs,
	}, nil
}

// GetReviewConfig gets the review config by full name.
func (c *mockClient) GetReviewConfig(_ context.Context, reviewConfigName string) (*v1pb.ReviewConfig, error) {
	config, ok := c.reviewConfigMap[reviewConfigName]
	if !ok {
		return nil, errors.Errorf("Cannot found review config %s", reviewConfigName)
	}
	return config, nil
}

// UpsertReviewConfig updates or creates the review config.
func (c *mockClient) UpsertReviewConfig(_ context.Context, reviewConfig *v1pb.ReviewConfig, updateMasks []string) (*v1pb.ReviewConfig, error) {
	existed, ok := c.reviewConfigMap[reviewConfig.Name]
	if !ok {
		// Create new review config
		c.reviewConfigMap[reviewConfig.Name] = reviewConfig
		return reviewConfig, nil
	}

	// Update existing review config
	if slices.Contains(updateMasks, "title") {
		existed.Title = reviewConfig.Title
	}
	if slices.Contains(updateMasks, "enabled") {
		existed.Enabled = reviewConfig.Enabled
	}
	if slices.Contains(updateMasks, "rules") {
		existed.Rules = reviewConfig.Rules
	}
	if slices.Contains(updateMasks, "resources") {
		existed.Resources = reviewConfig.Resources
	}
	// Creator is likely a read-only field, skip it

	c.reviewConfigMap[reviewConfig.Name] = existed
	return existed, nil
}

// DeleteReviewConfig deletes the review config.
func (c *mockClient) DeleteReviewConfig(_ context.Context, reviewConfigName string) error {
	delete(c.reviewConfigMap, reviewConfigName)
	return nil
}

// ListRisk lists the risk.
func (c *mockClient) ListRisk(_ context.Context) ([]*v1pb.Risk, error) {
	risks := make([]*v1pb.Risk, 0)
	for _, risk := range c.riskMap {
		risks = append(risks, risk)
	}
	return risks, nil
}

// GetRisk gets the risk by full name.
func (c *mockClient) GetRisk(_ context.Context, riskName string) (*v1pb.Risk, error) {
	risk, ok := c.riskMap[riskName]
	if !ok {
		return nil, errors.Errorf("Cannot found risk %s", riskName)
	}
	return risk, nil
}

// CreateRisk creates the risk.
func (c *mockClient) CreateRisk(_ context.Context, risk *v1pb.Risk) (*v1pb.Risk, error) {
	// Generate a unique name for the risk if not set
	if risk.Name == "" {
		risk.Name = fmt.Sprintf("risks/%d", len(c.riskMap)+1)
	}
	if _, exists := c.riskMap[risk.Name]; exists {
		return nil, errors.Errorf("risk %s already exists", risk.Name)
	}
	c.riskMap[risk.Name] = risk
	return risk, nil
}

// UpdateRisk updates the risk.
func (c *mockClient) UpdateRisk(_ context.Context, risk *v1pb.Risk, updateMasks []string) (*v1pb.Risk, error) {
	existed, ok := c.riskMap[risk.Name]
	if !ok {
		return nil, errors.Errorf("Cannot found risk %s", risk.Name)
	}

	if slices.Contains(updateMasks, "title") {
		existed.Title = risk.Title
	}
	if slices.Contains(updateMasks, "level") {
		existed.Level = risk.Level
	}
	if slices.Contains(updateMasks, "condition") {
		existed.Condition = risk.Condition
	}
	if slices.Contains(updateMasks, "source") {
		existed.Source = risk.Source
	}

	c.riskMap[risk.Name] = existed
	return existed, nil
}

// DeleteRisk deletes the risk by name.
func (c *mockClient) DeleteRisk(_ context.Context, riskName string) error {
	delete(c.riskMap, riskName)
	return nil
}

// FindEnvironment finds an environment by name in the environment settings
func FindEnvironment(ctx context.Context, client api.Client, name string) (*v1pb.EnvironmentSetting_Environment, int, []*v1pb.EnvironmentSetting_Environment, error) {
	environmentSetting, err := client.GetSetting(ctx, fmt.Sprintf("%s%s", SettingNamePrefix, v1pb.Setting_ENVIRONMENT.String()))
	if err != nil {
		return nil, 0, nil, errors.Wrapf(err, "failed to get environment setting")
	}

	enironmentList := environmentSetting.GetValue().GetEnvironmentSetting().GetEnvironments()
	if enironmentList == nil {
		enironmentList = []*v1pb.EnvironmentSetting_Environment{}
	}

	for index, env := range enironmentList {
		if env.Name == name {
			return env, index, enironmentList, nil
		}
	}
	return nil, 0, enironmentList, errors.Errorf("cannot found the environment %v", name)
}

// ListDatabaseGroup list all database groups in a project.
func (c *mockClient) ListDatabaseGroup(_ context.Context, projectName string) (*v1pb.ListDatabaseGroupsResponse, error) {
	groups := make([]*v1pb.DatabaseGroup, 0)
	for name, group := range c.databaseGroupMap {
		// Only return groups that belong to the specified project
		if strings.HasPrefix(name, projectName) {
			groups = append(groups, group)
		}
	}
	return &v1pb.ListDatabaseGroupsResponse{
		DatabaseGroups: groups,
	}, nil
}

// GetDatabaseGroup gets the database group by name.
func (c *mockClient) GetDatabaseGroup(_ context.Context, groupName string, _ v1pb.DatabaseGroupView) (*v1pb.DatabaseGroup, error) {
	group, ok := c.databaseGroupMap[groupName]
	if !ok {
		return nil, errors.Errorf("Cannot found database group %s", groupName)
	}
	return group, nil
}

// CreateDatabaseGroup creates the database group.
func (c *mockClient) CreateDatabaseGroup(_ context.Context, projectID, groupID string, group *v1pb.DatabaseGroup) (*v1pb.DatabaseGroup, error) {
	groupName := fmt.Sprintf("%s%s/databaseGroups/%s", ProjectNamePrefix, projectID, groupID)
	group.Name = groupName
	if _, exists := c.databaseGroupMap[groupName]; exists {
		return nil, errors.Errorf("database group %s already exists", groupName)
	}
	c.databaseGroupMap[groupName] = group
	return group, nil
}

// UpdateDatabaseGroup updates the database group.
func (c *mockClient) UpdateDatabaseGroup(_ context.Context, group *v1pb.DatabaseGroup, updateMasks []string) (*v1pb.DatabaseGroup, error) {
	existed, ok := c.databaseGroupMap[group.Name]
	if !ok {
		return nil, errors.Errorf("Cannot found database group %s", group.Name)
	}

	// DatabasePlaceholder might be named differently or not exist
	// Skip it for now as it's not a critical field
	if slices.Contains(updateMasks, "database_expr") {
		existed.DatabaseExpr = group.DatabaseExpr
	}
	if slices.Contains(updateMasks, "matched_databases") {
		existed.MatchedDatabases = group.MatchedDatabases
	}
	if slices.Contains(updateMasks, "unmatched_databases") {
		existed.UnmatchedDatabases = group.UnmatchedDatabases
	}

	c.databaseGroupMap[group.Name] = existed
	return existed, nil
}

// DeleteDatabaseGroup deletes the database group by name.
func (c *mockClient) DeleteDatabaseGroup(_ context.Context, groupName string) error {
	delete(c.databaseGroupMap, groupName)
	return nil
}

// CreateProjectWebhook creates the webhook in the project.
func (*mockClient) CreateProjectWebhook(_ context.Context, _ string, _ *v1pb.Webhook) (*v1pb.Webhook, error) {
	return &v1pb.Webhook{}, nil
}

// UpdateProjectWebhook updates the webhook.
func (*mockClient) UpdateProjectWebhook(_ context.Context, _ *v1pb.Webhook, _ []string) (*v1pb.Webhook, error) {
	return &v1pb.Webhook{}, nil
}

// DeleteProjectWebhook deletes the webhook.
func (*mockClient) DeleteProjectWebhook(_ context.Context, _ string) error {
	return nil
}
