package api

// InstanceMessage is the API message for an instance.
type InstanceMessage struct {
	UID          string               `json:"uid"`
	Name         string               `json:"name"`
	State        State                `json:"state,omitempty"`
	Title        string               `json:"title"`
	Engine       string               `json:"engine"`
	ExternalLink string               `json:"externalLink"`
	DataSources  []*DataSourceMessage `json:"dataSources"`
}

// InstanceFindMessage is the API message for finding instance.
type InstanceFindMessage struct {
	EnvironmentID string
	InstanceID    string
	ShowDeleted   bool
}

// ListInstanceMessage is the API message for list instance response.
type ListInstanceMessage struct {
	Instances     []*InstanceMessage `json:"instances"`
	NextPageToken string             `json:"nextPageToken"`
}
