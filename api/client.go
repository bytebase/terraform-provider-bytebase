package api

import (
	"context"

	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
	v1alpha1 "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// Auth
	// Login will login the user and get the response.
	Login() (*v1pb.LoginResponse, error)

	// Environment
	// CreateEnvironment creates the environment.
	CreateEnvironment(ctx context.Context, environmentID string, create *v1pb.Environment) (*v1pb.Environment, error)
	// GetEnvironment gets the environment by id.
	GetEnvironment(ctx context.Context, environmentName string) (*v1pb.Environment, error)
	// ListEnvironment finds all environments.
	ListEnvironment(ctx context.Context, showDeleted bool) (*v1pb.ListEnvironmentsResponse, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(ctx context.Context, patch *v1pb.Environment, updateMask []string) (*v1pb.Environment, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(ctx context.Context, environmentName string) error
	// UndeleteEnvironment undeletes the environment.
	UndeleteEnvironment(ctx context.Context, environmentName string) (*v1pb.Environment, error)

	// Instance
	// ListInstance will return instances.
	ListInstance(ctx context.Context, showDeleted bool) (*v1pb.ListInstancesResponse, error)
	// GetInstance gets the instance by id.
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
	ListDatabase(ctx context.Context, instanceID, filter string) (*v1pb.ListDatabasesResponse, error)
	// UpdateDatabase patches the database.
	UpdateDatabase(ctx context.Context, patch *v1pb.Database, updateMasks []string) (*v1pb.Database, error)

	// Project
	// GetProject gets the project by resource id.
	GetProject(ctx context.Context, projectName string) (*v1pb.Project, error)
	// ListProject list the projects,
	ListProject(ctx context.Context, showDeleted bool) (*v1pb.ListProjectsResponse, error)
	// CreateProject creates the project.
	CreateProject(ctx context.Context, projectID string, project *v1pb.Project) (*v1pb.Project, error)
	// UpdateProject updates the project.
	UpdateProject(ctx context.Context, patch *v1pb.Project, updateMask []string) (*v1pb.Project, error)
	// DeleteProject deletes the project.
	DeleteProject(ctx context.Context, projectName string) error
	// UndeleteProject undeletes the project.
	UndeleteProject(ctx context.Context, projectName string) (*v1pb.Project, error)

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

	// VCS Provider
	// ListVCSProvider will returns all vcs providers.
	ListVCSProvider(ctx context.Context) (*v1pb.ListVCSProvidersResponse, error)
	// GetVCSProvider gets the vcs by id.
	GetVCSProvider(ctx context.Context, name string) (*v1pb.VCSProvider, error)
	// CreateVCSProvider creates the vcs provider.
	CreateVCSProvider(ctx context.Context, vcsID string, vcs *v1pb.VCSProvider) (*v1pb.VCSProvider, error)
	// UpdateVCSProvider updates the vcs provider.
	UpdateVCSProvider(ctx context.Context, patch *v1pb.VCSProvider, updateMasks []string) (*v1pb.VCSConnector, error)
	// DeleteVCSProvider deletes the vcs provider.
	DeleteVCSProvider(ctx context.Context, name string) error

	// VCS Connector
	// ListVCSConnector will returns all vcs connector in a project.
	ListVCSConnector(ctx context.Context, projectName string) (*v1pb.ListVCSConnectorsResponse, error)
	// GetVCSConnector gets the vcs connector by id.
	GetVCSConnector(ctx context.Context, name string) (*v1pb.VCSConnector, error)
	// CreateVCSConnector creates the vcs connector in a project.
	CreateVCSConnector(ctx context.Context, projectName, connectorID string, connector *v1pb.VCSConnector) (*v1pb.VCSConnector, error)
	// UpdateVCSConnector updates the vcs connector.
	UpdateVCSConnector(ctx context.Context, patch *v1pb.VCSConnector, updateMasks []string) (*v1pb.VCSConnector, error)
	// DeleteVCSConnector deletes the vcs provider.
	DeleteVCSConnector(ctx context.Context, name string) error

	// User
	// ListUser list all users.
	ListUser(ctx context.Context, showDeleted bool) (*v1pb.ListUsersResponse, error)
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
}
