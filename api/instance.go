package api

// InstanceMessage is the API message for an instance.
type InstanceMessage struct {
	// Format: instances/{unique resource id}
	Name         string               `json:"name"`
	State        State                `json:"state,omitempty"`
	Title        string               `json:"title"`
	Engine       EngineType           `json:"engine"`
	ExternalLink string               `json:"externalLink"`
	DataSources  []*DataSourceMessage `json:"dataSources"`
	// Environment is the environment id with format environments/{resource id}
	Environment string `json:"environment"`
}

// InstancePatchMessage is the API message to patch the instance.
type InstancePatchMessage struct {
	Name         string               `json:"name"`
	Title        *string              `json:"title,omitempty"`
	ExternalLink *string              `json:"externalLink,omitempty"`
	DataSources  []*DataSourceMessage `json:"dataSources,omitempty"`
}

// InstanceFindMessage is the API message for finding instance.
type InstanceFindMessage struct {
	InstanceID  string
	ShowDeleted bool
}

// ListInstanceMessage is the API message for list instance response.
type ListInstanceMessage struct {
	Instances     []*InstanceMessage `json:"instances"`
	NextPageToken string             `json:"nextPageToken"`
}
