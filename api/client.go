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
	GetEnvironment(ctx context.Context, environmentID string) (*EnvironmentMessage, error)
	// ListEnvironment finds all environments.
	ListEnvironment(ctx context.Context, showDeleted bool) (*ListEnvironmentMessage, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(ctx context.Context, environmentID string, patch *EnvironmentPatchMessage) (*EnvironmentMessage, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(ctx context.Context, environmentID string) error
	// UndeleteEnvironment undeletes the environment.
	UndeleteEnvironment(ctx context.Context, environmentID string) (*EnvironmentMessage, error)

	// Instance
	// ListInstance will return instances.
	ListInstance(ctx context.Context, find *InstanceFindMessage) (*ListInstanceMessage, error)
	// GetInstance gets the instance by id.
	GetInstance(ctx context.Context, find *InstanceFindMessage) (*InstanceMessage, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, instanceID string, instance *InstanceMessage) (*InstanceMessage, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, instanceID string, patch *InstancePatchMessage) (*InstanceMessage, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, instanceID string) error
	// UndeleteInstance undeletes the instance.
	UndeleteInstance(ctx context.Context, instanceID string) (*InstanceMessage, error)
	// SyncInstanceSchema will trigger the schema sync for an instance.
	SyncInstanceSchema(ctx context.Context, instanceUID string) error

	// Role
	// CreateRole creates the role in the instance.
	CreateRole(ctx context.Context, instanceID string, create *RoleUpsert) (*Role, error)
	// GetRole gets the role by instance id and role name.
	GetRole(ctx context.Context, instanceID, roleName string) (*Role, error)
	// ListRole lists the role in instance.
	ListRole(ctx context.Context, instanceID string) ([]*Role, error)
	// UpdateRole updates the role in instance.
	UpdateRole(ctx context.Context, instanceID, roleName string, patch *RoleUpsert) (*Role, error)
	// DeleteRole deletes the role in the instance.
	DeleteRole(ctx context.Context, instanceID, roleName string) error

	// Policy
	// ListPolicies lists policies in a specific resource.
	ListPolicies(ctx context.Context, find *PolicyFindMessage) (*ListPolicyMessage, error)
	// GetPolicy gets a policy in a specific resource.
	GetPolicy(ctx context.Context, find *PolicyFindMessage) (*PolicyMessage, error)
	// UpsertPolicy creates or updates the policy.
	UpsertPolicy(ctx context.Context, find *PolicyFindMessage, patch *PolicyPatchMessage) (*PolicyMessage, error)
	// DeletePolicy deletes the policy.
	DeletePolicy(ctx context.Context, find *PolicyFindMessage) error

	// Database
	// GetDatabase gets the database by instance resource id and the database name.
	GetDatabase(ctx context.Context, find *DatabaseFindMessage) (*DatabaseMessage, error)
	// ListDatabase list the databases.
	ListDatabase(ctx context.Context, find *DatabaseFindMessage) (*ListDatabaseMessage, error)

	// Project
	// GetProject gets the project by resource id.
	GetProject(ctx context.Context, projectID string, showDeleted bool) (*ProjectMessage, error)
	// ListProject list the projects,
	ListProject(ctx context.Context, showDeleted bool) (*ListProjectMessage, error)
	// CreateProject creates the project.
	CreateProject(ctx context.Context, projectID string, project *ProjectMessage) (*ProjectMessage, error)
	// UpdateProject updates the project.
	UpdateProject(ctx context.Context, projectID string, patch *ProjectPatchMessage) (*ProjectMessage, error)
	// DeleteProject deletes the project.
	DeleteProject(ctx context.Context, projectID string) error
	// UndeleteProject undeletes the project.
	UndeleteProject(ctx context.Context, projectID string) (*ProjectMessage, error)
}
