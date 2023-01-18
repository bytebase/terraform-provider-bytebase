package api

// EngineType is the type of the instance engine.
type EngineType string

const (
	// EngineTypeMySQL is the database type for MYSQL.
	EngineTypeMySQL EngineType = "MYSQL"
	// EngineTypePostgres is the database type for POSTGRES.
	EngineTypePostgres EngineType = "POSTGRES"
	// EngineTypeTiDB is the database type for TiDB.
	EngineTypeTiDB EngineType = "TIDB"
	// EngineTypeSnowflake is the database type for SNOWFLAKE.
	EngineTypeSnowflake EngineType = "SNOWFLAKE"
	// EngineTypeClickHouse is the database type for CLICKHOUSE.
	EngineTypeClickHouse EngineType = "CLICKHOUSE"
	// EngineTypeMongoDB is the database type for MongoDB.
	EngineTypeMongoDB EngineType = "MONGODB"
	// EngineTypeSQLite is the database type for SQLite.
	EngineTypeSQLite EngineType = "SQLITE"
)

// InstanceMessage is the API message for an instance.
type InstanceMessage struct {
	UID          string               `json:"uid"`
	Name         string               `json:"name"`
	State        State                `json:"state,omitempty"`
	Title        string               `json:"title"`
	Engine       EngineType           `json:"engine"`
	ExternalLink string               `json:"externalLink"`
	DataSources  []*DataSourceMessage `json:"dataSources"`
}

// InstancePatchMessage is the API message to patch the instance.
type InstancePatchMessage struct {
	Title        *string              `json:"title,omitempty"`
	ExternalLink *string              `json:"externalLink,omitempty"`
	DataSources  []*DataSourceMessage `json:"dataSources,omitempty"`
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
