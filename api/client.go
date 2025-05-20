package api

import (
	"context"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// InstanceFilter is the filter for list instances API.
type InstanceFilter struct {
	Query       string
	Environment string
	Project     string
	State       v1pb.State
	Engines     []v1pb.Engine
	Host        string
	Port        string
}

// ProjectFilter is the filter for list projects API.
type ProjectFilter struct {
	Query          string
	ExcludeDefault bool
	State          v1pb.State
}

// Label is the database label.
type Label struct {
	Key   string
	Value string
}

// DatabaseFilter is the filter for list databases API.
type DatabaseFilter struct {
	Query             string
	Environment       string
	Project           string
	Instance          string
	Engines           []v1pb.Engine
	Labels            []*Label
	ExcludeUnassigned bool
}

// UserFilter is the filter for list users API.
type UserFilter struct {
	Name      string
	Email     string
	Project   string
	UserTypes []v1pb.UserType
	State     v1pb.State
}

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// GetCaller returns the API caller.
	GetCaller() *v1pb.User

	// Instance
	// ListInstance will return instances.
	ListInstance(ctx context.Context, filter *InstanceFilter) ([]*v1pb.Instance, error)
	// GetInstance gets the instance by full name.
	GetInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, instanceID string, instance *v1pb.Instance) (*v1pb.Instance, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, patch *v1pb.Instance, updateMasks []string) (*v1pb.Instance, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, instanceName string) error
	// UndeleteInstance undeletes the instance.
	UndeleteInstance(ctx context.Context, instanceName string) (*v1pb.Instance, error)
	// SyncInstanceSchema will trigger the schema sync for an instance.
	SyncInstanceSchema(ctx context.Context, instanceName string) error

	// Policy
	// ListPolicies lists policies in a specific resource.
	ListPolicies(ctx context.Context, parent string) (*v1pb.ListPoliciesResponse, error)
	// GetPolicy gets a policy in a specific resource.
	GetPolicy(ctx context.Context, policyName string) (*v1pb.Policy, error)
	// UpsertPolicy creates or updates the policy.
	UpsertPolicy(ctx context.Context, patch *v1pb.Policy, updateMasks []string) (*v1pb.Policy, error)
	// DeletePolicy deletes the policy.
	DeletePolicy(ctx context.Context, policyName string) error

	// Database
	// GetDatabase gets the database by instance resource id and the database name.
	GetDatabase(ctx context.Context, databaseName string) (*v1pb.Database, error)
	// ListDatabase list the databases.
	ListDatabase(ctx context.Context, instanceID string, filter *DatabaseFilter, listAll bool) ([]*v1pb.Database, error)
	// UpdateDatabase patches the database.
	UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error)
	// BatchUpdateDatabases batch updates databases.
	BatchUpdateDatabases(ctx context.Context, request *v1pb.BatchUpdateDatabasesRequest) (*v1pb.BatchUpdateDatabasesResponse, error)
	// GetDatabaseCatalog gets the database catalog by the database full name.
	GetDatabaseCatalog(ctx context.Context, databaseName string) (*v1pb.DatabaseCatalog, error)
	// UpdateDatabaseCatalog patches the database catalog.
	UpdateDatabaseCatalog(ctx context.Context, patch *v1pb.DatabaseCatalog) (*v1pb.DatabaseCatalog, error)

	// Project
	// GetProject gets the project by project full name.
	GetProject(ctx context.Context, projectName string) (*v1pb.Project, error)
	// ListProject list all projects,
	ListProject(ctx context.Context, filter *ProjectFilter) ([]*v1pb.Project, error)
	// CreateProject creates the project.
	CreateProject(ctx context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error)
	// UpdateProject updates the project.
	UpdateProject(ctx context.Context, patch *v1pb.Project, updateMask []string) (*v1pb.Project, error)
	// DeleteProject deletes the project.
	DeleteProject(ctx context.Context, projectName string) error
	// UndeleteProject undeletes the project.
	UndeleteProject(ctx context.Context, projectName string) (*v1pb.Project, error)
	// GetProjectIAMPolicy gets the project IAM policy by project full name.
	GetProjectIAMPolicy(ctx context.Context, projectName string) (*v1pb.IamPolicy, error)
	// SetProjectIAMPolicy sets the project IAM policy.
	SetProjectIAMPolicy(ctx context.Context, projectName string, update *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error)

	// Setting
	// ListSettings lists all settings.
	ListSettings(ctx context.Context) (*v1pb.ListSettingsResponse, error)
	// GetSetting gets the setting by the name.
	GetSetting(ctx context.Context, settingName string) (*v1pb.Setting, error)
	// UpsertSetting updates or creates the setting.
	UpsertSetting(ctx context.Context, upsert *v1pb.Setting, updateMasks []string) (*v1pb.Setting, error)

	// Cel
	// ParseExpression parse the expression string.
	ParseExpression(ctx context.Context, expression string) (*v1alpha1.Expr, error)

	// User
	// ListUser list all users.
	ListUser(ctx context.Context, filter *UserFilter) ([]*v1pb.User, error)
	// CreateUser creates the user.
	CreateUser(ctx context.Context, user *v1pb.User) (*v1pb.User, error)
	// GetUser gets the user by name.
	GetUser(ctx context.Context, userName string) (*v1pb.User, error)
	// UpdateUser updates the user.
	UpdateUser(ctx context.Context, patch *v1pb.User, updateMasks []string) (*v1pb.User, error)
	// DeleteUser deletes the user by name.
	DeleteUser(ctx context.Context, userName string) error
	// UndeleteUser undeletes the user by name.
	UndeleteUser(ctx context.Context, userName string) (*v1pb.User, error)

	// Role
	// ListRole will returns all roles.
	ListRole(ctx context.Context) (*v1pb.ListRolesResponse, error)
	// DeleteRole deletes the role by name.
	DeleteRole(ctx context.Context, name string) error
	// CreateRole creates the role.
	CreateRole(ctx context.Context, roleID string, role *v1pb.Role) (*v1pb.Role, error)
	// GetRole gets the role by full name.
	GetRole(ctx context.Context, name string) (*v1pb.Role, error)
	// UpdateRole updates the role.
	UpdateRole(ctx context.Context, patch *v1pb.Role, updateMasks []string) (*v1pb.Role, error)

	// Group
	// ListGroup list all groups.
	ListGroup(ctx context.Context) (*v1pb.ListGroupsResponse, error)
	// CreateGroup creates the group.
	CreateGroup(ctx context.Context, email string, group *v1pb.Group) (*v1pb.Group, error)
	// GetGroup gets the group by name.
	GetGroup(ctx context.Context, name string) (*v1pb.Group, error)
	// UpdateGroup updates the group.
	UpdateGroup(ctx context.Context, patch *v1pb.Group, updateMasks []string) (*v1pb.Group, error)
	// DeleteGroup deletes the group by name.
	DeleteGroup(ctx context.Context, name string) error

	// Workspace
	// GetWorkspaceIAMPolicy gets the workspace IAM policy.
	GetWorkspaceIAMPolicy(ctx context.Context) (*v1pb.IamPolicy, error)
	// SetWorkspaceIAMPolicy sets the workspace IAM policy.
	SetWorkspaceIAMPolicy(ctx context.Context, setIamPolicyRequest *v1pb.SetIamPolicyRequest) (*v1pb.IamPolicy, error)

	// Review config
	// ListReviewConfig will return review configs.
	ListReviewConfig(ctx context.Context) (*v1pb.ListReviewConfigsResponse, error)
	// GetReviewConfig gets the review config by full name.
	GetReviewConfig(ctx context.Context, reviewName string) (*v1pb.ReviewConfig, error)
	// UpsertReviewConfig updates or creates the review config.
	UpsertReviewConfig(ctx context.Context, patch *v1pb.ReviewConfig, updateMasks []string) (*v1pb.ReviewConfig, error)
	// DeleteReviewConfig deletes the review config.
	DeleteReviewConfig(ctx context.Context, reviewName string) error

	// Risk
	// ListRisk lists the risk.
	ListRisk(ctx context.Context) ([]*v1pb.Risk, error)
	// GetRisk gets the risk by full name.
	GetRisk(ctx context.Context, name string) (*v1pb.Risk, error)
	// CreateRisk creates the risk.
	CreateRisk(ctx context.Context, risk *v1pb.Risk) (*v1pb.Risk, error)
	// UpdateRisk updates the risk.
	UpdateRisk(ctx context.Context, patch *v1pb.Risk, updateMasks []string) (*v1pb.Risk, error)
	// DeleteRisk deletes the risk by name.
	DeleteRisk(ctx context.Context, name string) error

	// ListDatabaseGroup list all database groups in a project.
	ListDatabaseGroup(ctx context.Context, project string) (*v1pb.ListDatabaseGroupsResponse, error)
	// CreateDatabaseGroup creates the database group.
	CreateDatabaseGroup(ctx context.Context, project, groupID string, group *v1pb.DatabaseGroup) (*v1pb.DatabaseGroup, error)
	// GetDatabaseGroup gets the database group by name.
	GetDatabaseGroup(ctx context.Context, name string, view v1pb.DatabaseGroupView) (*v1pb.DatabaseGroup, error)
	// UpdateDatabaseGroup updates the database group.
	UpdateDatabaseGroup(ctx context.Context, patch *v1pb.DatabaseGroup, updateMasks []string) (*v1pb.DatabaseGroup, error)
	// DeleteDatabaseGroup deletes the database group by name.
	DeleteDatabaseGroup(ctx context.Context, name string) error
}
