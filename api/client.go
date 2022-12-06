package api

import "context"

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// Auth
	// Login will login the user and get the response.
	Login() (*AuthResponse, error)

	// Environment
	// CreateEnvironment creates the environment.
	CreateEnvironment(ctx context.Context, create *EnvironmentCreate) (*Environment, error)
	// GetEnvironment gets the environment by id.
	GetEnvironment(ctx context.Context, environmentID int) (*Environment, error)
	// ListEnvironment finds all environments.
	ListEnvironment(ctx context.Context, find *EnvironmentFind) ([]*Environment, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(ctx context.Context, environmentID int, patch *EnvironmentPatch) (*Environment, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(ctx context.Context, environmentID int) error

	// Instance
	// ListInstance will return all instances.
	ListInstance(ctx context.Context, find *InstanceFind) ([]*Instance, error)
	// CreateInstance creates the instance.
	CreateInstance(ctx context.Context, create *InstanceCreate) (*Instance, error)
	// GetInstance gets the instance by id.
	GetInstance(ctx context.Context, instanceID int) (*Instance, error)
	// UpdateInstance updates the instance.
	UpdateInstance(ctx context.Context, instanceID int, patch *InstancePatch) (*Instance, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(ctx context.Context, instanceID int) error
}
