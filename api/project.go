package api

// ProjectMessage is the API message for project.
type ProjectMessage struct {
	Name           string `json:"name"`
	Title          string `json:"title"`
	Key            string `json:"key"`
	Workflow       string `json:"workflow"`
	Visibility     string `json:"visibility"`
	TenantMode     string `json:"tenantMode"`
	DBNameTemplate string `json:"dbNameTemplate"`
	SchemaVersion  string `json:"schemaVersion"`
	SchemaChange   string `json:"schemaChange"`
	State          State  `json:"state,omitempty"`
}

// ListProjectMessage is the API message for list project response.
type ListProjectMessage struct {
	Projects      []*ProjectMessage `json:"projects"`
	NextPageToken string            `json:"nextPageToken"`
}

// ProjectPatchMessage is the API message to patch the project.
type ProjectPatchMessage struct {
	Title          *string `json:"title,omitempty"`
	Key            *string `json:"key,omitempty"`
	Workflow       *string `json:"workflow"`
	TenantMode     *string `json:"tenantMode"`
	DBNameTemplate *string `json:"dbNameTemplate"`
	SchemaChange   *string `json:"schemaChange"`
}
