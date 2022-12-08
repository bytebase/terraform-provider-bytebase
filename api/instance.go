package api

// Instance is the API message for an instance.
type Instance struct {
	ID int `jsonapi:"primary,instance" json:"id"`

	// Related fields
	Environment    string        `json:"environment"`
	DataSourceList []*DataSource `json:"dataSourceList"`

	// Domain specific fields
	Name          string `json:"name"`
	Engine        string `json:"engine"`
	EngineVersion string `json:"engineVersion"`
	ExternalLink  string `json:"externalLink"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	Database      string `json:"database"`
}

// InstanceFind is the API message for finding instance.
type InstanceFind struct {
	// Domain specific fields
	Name string `url:"name,omitempty"`
}

// InstanceCreate is the API message for creating an instance.
type InstanceCreate struct {
	// Related fields
	Environment    string              `json:"environment"`
	DataSourceList []*DataSourceCreate `json:"dataSourceList"`

	// Domain specific fields
	Name         string `json:"name"`
	Engine       string `json:"engine"`
	ExternalLink string `json:"externalLink"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	Database     string `json:"database"`
}

// InstancePatch is the API message for patching an instance.
type InstancePatch struct {
	// Related fields
	DataSourceList []*DataSourceCreate `json:"dataSourceList"`

	// Domain specific fields
	Name         *string `json:"name,omitempty"`
	ExternalLink *string `json:"externalLink,omitempty"`
	Host         *string `json:"host,omitempty"`
	Port         *string `json:"port,omitempty"`
	Database     *string `json:"database,omitempty"`
}

// HasChange returns if the patch struct has the value to update.
func (p *InstancePatch) HasChange() bool {
	return p.Name != nil ||
		p.ExternalLink != nil ||
		p.Host != nil ||
		p.Port != nil ||
		p.Database != nil ||
		p.DataSourceList != nil
}
