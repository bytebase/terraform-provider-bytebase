package api

// DeploymentType is the type for deployment.
type DeploymentType string

const (
	// DeploymentTypeDatabaseCreate is the deployment type for creating databases.
	DeploymentTypeDatabaseCreate DeploymentType = "DATABASE_CREATE"
	// DeploymentTypeDatabaseDDL is the deployment type for updating database schemas (DDL).
	DeploymentTypeDatabaseDDL DeploymentType = "DATABASE_DDL"
	// DeploymentTypeDatabaseDDLGhost is the deployment type for updating database schemas using gh-ost.
	DeploymentTypeDatabaseDDLGhost DeploymentType = "DATABASE_DDL_GHOST"
	// DeploymentTypeDatabaseDML is the deployment type for updating database data (DML).
	DeploymentTypeDatabaseDML DeploymentType = "DATABASE_DML"
	// DeploymentTypeDatabaseRestorePITR is the deployment type for performing a Point-in-time Recovery.
	DeploymentTypeDatabaseRestorePITR DeploymentType = "DATABASE_RESTORE_PITR"
	// DeploymentTypeDatabaseDMLRollback is the deployment type for a generated rollback issue.
	DeploymentTypeDatabaseDMLRollback DeploymentType = "DATABASE_DML_ROLLBACK"
)
