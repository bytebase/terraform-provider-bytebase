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
}
