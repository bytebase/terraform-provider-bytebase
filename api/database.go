package api

// DatabaseFindMessage is the API message for finding database.
type DatabaseFindMessage struct {
	EnvironmentID string
	InstanceID    string
	DatabaseName  string
}

// DatabaseMessage is the API message for database.
type DatabaseMessage struct {
	Name          string `json:"name"`
	Project       string `json:"project"`
	SchemaVersion string `json:"schemaVersion"`
}
