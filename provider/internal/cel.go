package internal

// CEL attribute names for resource scope.
const (
	// CELAttributeResourceEnvironmentID is the environment ID of the resource.
	CELAttributeResourceEnvironmentID = "resource.environment_id"
	// CELAttributeResourceInstanceID is the instance ID of the resource.
	CELAttributeResourceInstanceID = "resource.instance_id"
	// CELAttributeResourceDatabaseName is the database name of the resource.
	CELAttributeResourceDatabaseName = "resource.database_name"
	// CELAttributeResourceSchemaName is the schema name of the resource.
	CELAttributeResourceSchemaName = "resource.schema_name"
	// CELAttributeResourceTableName is the table name of the resource.
	CELAttributeResourceTableName = "resource.table_name"
	// CELAttributeResourceColumnName is the column name of the resource.
	CELAttributeResourceColumnName = "resource.column_name"
	// CELAttributeResourceDatabase is the full database name of the resource (used in IAM policy conditions).
	CELAttributeResourceDatabase = "resource.database"
	// CELAttributeRequestTime is the timestamp of the request.
	CELAttributeRequestTime = "request.time"
)
