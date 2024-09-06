package api

import "context"

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// Auth
	// Login will login the user and get the response.
	Login() (*AuthResponse, error)

	// Environment
	// CreateEnvironment creates the environment.
	CreateEnvironment(ctx context.Context, environmentID string, create *EnvironmentMessage) (*EnvironmentMessage, error)
	// GetEnvironment gets the environment by id.
	GetEnvironment(ctx context.Context, environmentName string) (*EnvironmentMessage, error)
	// ListEnvironment finds all environments.
	ListEnvironment(ctx context.Context, showDeleted bool) (*ListEnvironmentMessage, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(ctx context.Context, patch *EnvironmentPatchMessage) (*EnvironmentMessage, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(ctx context.Context, environmentName string) error
	// UndeleteEnvironment undeletes the environment.
	UndeleteEnvironment(ctx context.Context, environmentName string) (*EnvironmentMessage, error)

	// Instance
	// ListInstance will return instances.
	ListInstance(ctx context.Context, find *InstanceFindMessage) (*ListInstanceMessage, error)
	// GetInstance gets the instance by id.
	GetInstance(ctx context.Context, instanceName string) (*InstanceMessage, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, instanceID string, instance *InstanceMessage) (*InstanceMessage, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, patch *InstancePatchMessage) (*InstanceMessage, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, instanceName string) error
	// UndeleteInstance undeletes the instance.
	UndeleteInstance(ctx context.Context, instanceName string) (*InstanceMessage, error)
	// SyncInstanceSchema will trigger the schema sync for an instance.
	SyncInstanceSchema(ctx context.Context, instanceName string) error

	// Policy
	// ListPolicies lists policies in a specific resource.
	ListPolicies(ctx context.Context, find *PolicyFindMessage) (*ListPolicyMessage, error)
	// GetPolicy gets a policy in a specific resource.
	GetPolicy(ctx context.Context, policyName string) (*PolicyMessage, error)
	// UpsertPolicy creates or updates the policy.
	UpsertPolicy(ctx context.Context, patch *PolicyPatchMessage) (*PolicyMessage, error)
	// DeletePolicy deletes the policy.
	DeletePolicy(ctx context.Context, policyName string) error

	// Database
	// GetDatabase gets the database by instance resource id and the database name.
	GetDatabase(ctx context.Context, databaseName string) (*DatabaseMessage, error)
	// ListDatabase list the databases.
	ListDatabase(ctx context.Context, find *DatabaseFindMessage) (*ListDatabaseMessage, error)
	// UpdateDatabase patches the database.
	UpdateDatabase(ctx context.Context, patch *DatabasePatchMessage) (*DatabaseMessage, error)

	// Project
	// GetProject gets the project by resource id.
	GetProject(ctx context.Context, projectName string) (*ProjectMessage, error)
	// ListProject list the projects,
	ListProject(ctx context.Context, showDeleted bool) (*ListProjectMessage, error)
	// CreateProject creates the project.
	CreateProject(ctx context.Context, projectID string, project *ProjectMessage) (*ProjectMessage, error)
	// UpdateProject updates the project.
	UpdateProject(ctx context.Context, patch *ProjectPatchMessage) (*ProjectMessage, error)
	// DeleteProject deletes the project.
	DeleteProject(ctx context.Context, projectName string) error
	// UndeleteProject undeletes the project.
	UndeleteProject(ctx context.Context, projectName string) (*ProjectMessage, error)
}
