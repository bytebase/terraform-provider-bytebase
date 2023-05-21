package api

// State is the state for a row.
type State string

const (
	// Active is the state for a normal row.
	Active State = "ACTIVE"
	// Deleted is the state for an removed row.
	Deleted State = "DELETED"
)

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
	// EngineTypeRedis is the database type for REDIS.
	EngineTypeRedis EngineType = "REDIS"
	// EngineTypeOracle is the database type for ORACLE.
	EngineTypeOracle EngineType = "ORACLE"
	// EngineTypeSpanner is the database type for SPANNER.
	EngineTypeSpanner EngineType = "SPANNER"
	// EngineTypeMSSQL is the database type for MSSQL.
	EngineTypeMSSQL EngineType = "MSSQL"
	// EngineTypeRedshift is the database type for REDSHIFT.
	EngineTypeRedshift EngineType = "REDSHIFT"
	// EngineTypeMariaDB is the database type for MARIADB.
	EngineTypeMariaDB EngineType = "MARIADB"
	// EngineTypeOceanbase is the database type for OCEANBASE.
	EngineTypeOceanbase EngineType = "OCEANBASE"
)
