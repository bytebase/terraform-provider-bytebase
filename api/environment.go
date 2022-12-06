package api

// Environment is the API message for an environment.
type Environment struct {
	ID int `json:"id"`

	// Domain specific fields
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// EnvironmentFind is the API message for finding environment.
type EnvironmentFind struct {
	// Domain specific fields
	Name string `url:"name,omitempty"`
}

// EnvironmentCreate is the API message for creating an environment.
type EnvironmentCreate struct {
	// Domain specific fields
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// EnvironmentPatch is the API message for patching an environment.
type EnvironmentPatch struct {
	// Domain specific fields
	Name  *string `json:"name,omitempty"`
	Order *int    `json:"order,omitempty"`
}
