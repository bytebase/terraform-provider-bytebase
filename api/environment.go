package api

// EnvironmentTier is the protection info for environment.
type EnvironmentTier string

const (
	// EnvironmentTierProtected is the PROTECTED tier.
	EnvironmentTierProtected EnvironmentTier = "PROTECTED"
	// EnvironmentTierUnProtected is the UNPROTECTED tier.
	EnvironmentTierUnProtected EnvironmentTier = "UNPROTECTED"
)

// EnvironmentMessage is the API message for an environment.
type EnvironmentMessage struct {
	// Domain specific fields
	// Format: environments/{unique resource id}
	Name  string          `json:"name"`
	Title string          `json:"title"`
	Order int             `json:"order"`
	State State           `json:"state,omitempty"`
	Tier  EnvironmentTier `json:"tier"`
}

// ListEnvironmentMessage is the API message for list environment response.
type ListEnvironmentMessage struct {
	Environments  []*EnvironmentMessage `json:"environments"`
	NextPageToken string                `json:"nextPageToken"`
}

// EnvironmentPatchMessage is the API message to patch the environment.
type EnvironmentPatchMessage struct {
	Name  string           `json:"name"`
	Title *string          `json:"title,omitempty"`
	Order *int             `json:"order,omitempty"`
	Tier  *EnvironmentTier `json:"tier,omitempty"`
}
