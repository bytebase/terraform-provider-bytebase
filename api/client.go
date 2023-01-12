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
	// ListInstance will return instances in environment.
	ListInstance(ctx context.Context, find *InstanceFindMessage) (*ListInstanceMessage, error)
	// GetInstance gets the instance by id.
	GetInstance(ctx context.Context, find *InstanceFindMessage) (*InstanceMessage, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, environmentID, instanceID string, instance *InstanceMessage) (*InstanceMessage, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, environmentID, instanceID string, patch *InstancePatchMessage) (*InstanceMessage, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, environmentID, instanceID string) error
	// UndeleteInstance undeletes the instance.
	UndeleteInstance(ctx context.Context, environmentID, instanceID string) (*InstanceMessage, error)

	// Role
	// CreateRole creates the role in the instance.
	CreateRole(ctx context.Context, environmentID, instanceID string, create *RoleUpsert) (*Role, error)
	// GetRole gets the role by instance id and role name.
	GetRole(ctx context.Context, environmentID, instanceID, roleName string) (*Role, error)
	// ListRole lists the role in instance.
	ListRole(ctx context.Context, environmentID, instanceID string) ([]*Role, error)
	// UpdateRole updates the role in instance.
	UpdateRole(ctx context.Context, environmentID, instanceID, roleName string, patch *RoleUpsert) (*Role, error)
	// DeleteRole deletes the role in the instance.
	DeleteRole(ctx context.Context, environmentID, instanceID, roleName string) error

	// Policy
	// ListPolicies lists policies in a specific resource.
	ListPolicies(ctx context.Context, find *PolicyFindMessage) (*ListPolicyMessage, error)
	// GetPolicy gets a policy in a specific resource.
	GetPolicy(ctx context.Context, find *PolicyFindMessage) (*PolicyMessage, error)
}
