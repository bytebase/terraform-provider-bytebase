package api

// Client is the API message for Bytebase OpenAPI client.
type Client interface {
	// Auth
	// Login will login the user and get the response.
	Login() (*AuthResponse, error)

	// Environment
	// CreateEnvironment creates the environment.
	CreateEnvironment(create *EnvironmentCreate) (*Environment, error)
	// GetEnvironment gets the environment by id.
	GetEnvironment(environmentID int) (*Environment, error)
	// ListEnvironment finds all environments.
	ListEnvironment() ([]*Environment, error)
	// UpdateEnvironment updates the environment.
	UpdateEnvironment(environmentID int, patch *EnvironmentPatch) (*Environment, error)
	// DeleteEnvironment deletes the environment.
	DeleteEnvironment(environmentID int) error

	// Instance
	// ListInstance will return all instances.
	ListInstance() ([]*Instance, error)
	// CreateInstance creates the instance.
	CreateInstance(create *InstanceCreate) (*Instance, error)
	// GetInstance gets the instance by id.
	GetInstance(instanceID int) (*Instance, error)
	// UpdateInstance updates the instance.
	UpdateInstance(instanceID int, patch *InstancePatch) (*Instance, error)
	// DeleteInstance deletes the instance.
	DeleteInstance(instanceID int) error
}
