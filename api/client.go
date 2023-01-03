package api

import "context"

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// Auth
	// Login will login the user and get the response.
	Login() (*AuthResponse, error)

	// Environment
	// CreateEnvironment creates the environment.
	CreateEnvironment(ctx context.Context, create *EnvironmentUpsert) (*Environment, error)
	// GetEnvironment gets the environment by id.
	GetEnvironment(ctx context.Context, environmentID int) (*Environment, error)
	// ListEnvironment finds all environments.
	ListEnvironment(ctx context.Context, find *EnvironmentFind) ([]*Environment, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(ctx context.Context, environmentID int, patch *EnvironmentUpsert) (*Environment, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(ctx context.Context, environmentID int) error

	// Instance
	// ListInstance will return instances in environment.
	ListInstance(ctx context.Context, find *InstanceFindMessage) (*ListInstanceMessage, error)
	// GetInstance gets the instance by id.
	GetInstance(ctx context.Context, find *InstanceFindMessage) (*InstanceMessage, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, environmentID, instanceID string, instance *InstanceMessage) (*InstanceMessage, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, environmentID, instanceID string, instance *InstanceMessage) (*InstanceMessage, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, environmentID, instanceID string) error

	// Role
	// CreateRole creates the role in the instance.
	CreateRole(ctx context.Context, instanceID int, create *RoleUpsert) (*Role, error)
	// GetRole gets the role by instance id and role name.
	GetRole(ctx context.Context, instanceID int, roleName string) (*Role, error)
	// ListRole lists the role in instance.
	ListRole(ctx context.Context, instanceID int) ([]*Role, error)
	// UpdateRole updates the role in instance.
	UpdateRole(ctx context.Context, instanceID int, roleName string, patch *RoleUpsert) (*Role, error)
	// DeleteRole deletes the role in the instance.
	DeleteRole(ctx context.Context, instanceID int, roleName string) error
}
