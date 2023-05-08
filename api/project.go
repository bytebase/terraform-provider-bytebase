package api

// ProjectWorkflow is the workflow for project.
type ProjectWorkflow string

const (
	// ProjectWorkflowUI is the UI workflow type.
	ProjectWorkflowUI ProjectWorkflow = "UI"
	// ProjectWorkflowVCS is the VCS workflow type.
	ProjectWorkflowVCS ProjectWorkflow = "VCS"
)

// ProjectVisibility is the visibility for project.
type ProjectVisibility string

const (
	// ProjectVisibilityPublic is the public visibility type for project.
	ProjectVisibilityPublic ProjectVisibility = "VISIBILITY_PUBLIC"
	// ProjectVisibilityPrivate is the private visibility type for project.
	ProjectVisibilityPrivate ProjectVisibility = "VISIBILITY_PRIVATE"
)

// ProjectTenantMode is the tenant mode for project.
type ProjectTenantMode string

const (
	// ProjectTenantModeDisabled means the tenant mode for the project is disabled.
	ProjectTenantModeDisabled ProjectTenantMode = "TENANT_MODE_DISABLED"
	// ProjectTenantModeEnabled means the tenant mode for the project is enabled.
	ProjectTenantModeEnabled ProjectTenantMode = "TENANT_MODE_ENABLED"
)

// ProjectSchemaVersion is the schema version type for project.
type ProjectSchemaVersion string

const (
	// ProjectSchemaVersionTimestamp the timestamp schema version type in the project.
	ProjectSchemaVersionTimestamp ProjectSchemaVersion = "TIMESTAMP"
	// ProjectSchemaVersionSemantic the semantic schema version type in the project.
	ProjectSchemaVersionSemantic ProjectSchemaVersion = "SEMANTIC"
)

// ProjectSchemaChange is the schema change type for project.
type ProjectSchemaChange string

const (
	// ProjectSchemaChangeDDL the DDL schema change type in the project.
	ProjectSchemaChangeDDL ProjectSchemaChange = "DDL"
	// ProjectSchemaChangeSDL the SDL schema change type in the project.
	ProjectSchemaChangeSDL ProjectSchemaChange = "SDL"
)

// ProjectMessage is the API message for project.
type ProjectMessage struct {
	Name           string               `json:"name"`
	Title          string               `json:"title"`
	Key            string               `json:"key"`
	Workflow       ProjectWorkflow      `json:"workflow"`
	Visibility     ProjectVisibility    `json:"visibility"`
	TenantMode     ProjectTenantMode    `json:"tenantMode"`
	DBNameTemplate string               `json:"dbNameTemplate"`
	SchemaVersion  ProjectSchemaVersion `json:"schemaVersion"`
	SchemaChange   ProjectSchemaChange  `json:"schemaChange"`
	State          State                `json:"state,omitempty"`
}

// ListProjectMessage is the API message for list project response.
type ListProjectMessage struct {
	Projects      []*ProjectMessage `json:"projects"`
	NextPageToken string            `json:"nextPageToken"`
}

// ProjectPatchMessage is the API message to patch the project.
type ProjectPatchMessage struct {
	Title          *string              `json:"title,omitempty"`
	Key            *string              `json:"key,omitempty"`
	Workflow       *ProjectWorkflow     `json:"workflow"`
	TenantMode     *ProjectTenantMode   `json:"tenantMode"`
	DBNameTemplate *string              `json:"dbNameTemplate"`
	SchemaChange   *ProjectSchemaChange `json:"schemaChange"`
}
