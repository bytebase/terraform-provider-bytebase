package api

// PorjectMessage is the API message for project.
type PorjectMessage struct {
	Name           string `json:"name"`
	Title          string `json:"title"`
	Key            string `json:"key"`
	Workflow       string `json:"workflow"`
	Visibility     string `json:"visibility"`
	TenantMode     string `json:"tenantMode"`
	DBNameTemplate string `json:"dbNameTemplate"`
	SchemaVersion  string `json:"schemaVersion"`
	SchemaChange   string `json:"schemaChange"`
	LGTMCheck      string `json:"lgtmCheck"`
}
