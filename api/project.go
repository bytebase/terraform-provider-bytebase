package api

// ProjectWorkflow is the workflow for project.
type ProjectWorkflow string

const (
	// ProjectWorkflowUI is the UI workflow type.
	ProjectWorkflowUI ProjectWorkflow = "UI"
	// ProjectWorkflowVCS is the VCS workflow type.
	ProjectWorkflowVCS ProjectWorkflow = "VCS"
)

// ProjectMessage is the API message for project.
type ProjectMessage struct {
	// Format: projects/{unique resource id}
	Name     string          `json:"name"`
	Title    string          `json:"title"`
	Key      string          `json:"key"`
	Workflow ProjectWorkflow `json:"workflow"`
	State    State           `json:"state,omitempty"`
}

// ListProjectMessage is the API message for list project response.
type ListProjectMessage struct {
	Projects      []*ProjectMessage `json:"projects"`
	NextPageToken string            `json:"nextPageToken"`
}

// ProjectPatchMessage is the API message to patch the project.
type ProjectPatchMessage struct {
	// Format: projects/{unique resource id}
	Name  string  `json:"name"`
	Title *string `json:"title,omitempty"`
	Key   *string `json:"key,omitempty"`
}
