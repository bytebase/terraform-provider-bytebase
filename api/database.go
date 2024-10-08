package api

// DatabaseFindMessage is the API message for finding database.
type DatabaseFindMessage struct {
	InstanceID   string
	DatabaseName string
	Filter       *string
}

// DatabaseMessage is the API message for database.
type DatabaseMessage struct {
	// Format: instances/{unique resource id}/databases/{database name}
	Name string `json:"name"`
	// Format: projects/{unique resource id}
	Project string `json:"project"`
	// Format: environments/{unique resource id}
	Environment        string            `json:"environment"`
	SchemaVersion      string            `json:"schemaVersion"`
	SyncState          State             `json:"syncState"`
	SuccessfulSyncTime string            `json:"successfulSyncTime"`
	Labels             map[string]string `json:"labels"`
}

// ListDatabaseMessage is the API message for list database response.
type ListDatabaseMessage struct {
	Databases     []*DatabaseMessage `json:"databases"`
	NextPageToken string             `json:"nextPageToken"`
}

// DatabasePatchMessage is the API message to patch the database.
type DatabasePatchMessage struct {
	Name    string             `json:"name"`
	Project *string            `json:"project"`
	Labels  *map[string]string `json:"labels"`
}
