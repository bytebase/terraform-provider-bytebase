# Instance Examples

This directory contains examples for creating and querying database instances with various authentication types and engine-specific configurations.

## Authentication Types

| Type | Description | Supported Engines |
|------|-------------|-------------------|
| `PASSWORD` | Username/password (default) | Most engines |
| `AWS_RDS_IAM` | AWS IAM authentication | POSTGRES, MYSQL, ELASTICSEARCH, and others |
| `AZURE_IAM` | Azure AD authentication | MSSQL, COSMOSDB |
| `GOOGLE_CLOUD_SQL_IAM` | GCP IAM authentication | POSTGRES, MYSQL, SPANNER, BIGQUERY |

### Engine-specific Requirements

- **COSMOSDB**: Only supports `AZURE_IAM`
- **SPANNER, BIGQUERY**: Only support `GOOGLE_CLOUD_SQL_IAM`
- **MSSQL**: Supports `PASSWORD` or `AZURE_IAM`
- **ELASTICSEARCH**: Supports `PASSWORD` or `AWS_RDS_IAM`

## Engine-specific Fields

| Field | Engine | Description |
|-------|--------|-------------|
| `srv`, `authentication_database`, `replica_set`, `direct_connection`, `additional_addresses` | MONGODB | MongoDB connection options |
| `sid`, `service_name` | ORACLE | Oracle connection identifiers |
| `redis_type`, `master_name`, `master_username`, `master_password` | REDIS | Redis deployment configuration |
| `warehouse_id` | DATABRICKS | SQL warehouse ID |
| `cluster` | COCKROACHDB | CockroachDB cluster name |
| `sasl_config` | HIVE | SASL/Kerberos authentication |
| `region` | Any (with `AWS_RDS_IAM`) | AWS region for IAM auth |

## Features Requiring PASSWORD Authentication

These features only work with `authentication_type = "PASSWORD"`:

- **SSH Tunnel**: `ssh_host`, `ssh_port`, `ssh_user`, `ssh_password`, `ssh_private_key`
  - Only available for: MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS
- **External Secret**: `external_secret` (Vault, AWS Secrets Manager, GCP Secret Manager)

## Extra Connection Parameters

The `extra_connection_parameters` field is only available for:
- MYSQL, MARIADB, OCEANBASE, POSTGRES, ORACLE, MSSQL, MONGODB

## Examples

### Basic PASSWORD Authentication

```terraform
resource "bytebase_instance" "mysql" {
  resource_id = "mysql-example"
  environment = "environments/test"
  title       = "MySQL Instance"
  engine      = "MYSQL"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "mysql.example.com"
    port     = "3306"
    username = "admin"
    password = "your-password"
  }
}
```

### AWS IAM Authentication

```terraform
resource "bytebase_instance" "postgres_aws" {
  resource_id = "postgres-aws-iam"
  environment = "environments/test"
  title       = "PostgreSQL with AWS IAM"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "mydb.us-east-1.rds.amazonaws.com"
    port                = "5432"
    username            = "iam_user"
    authentication_type = "AWS_RDS_IAM"
    region              = "us-east-1"
    aws_credential {
      access_key_id     = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key = "your-secret-key"
    }
  }
}
```

### MongoDB with Replica Set

```terraform
resource "bytebase_instance" "mongodb" {
  resource_id = "mongodb-example"
  environment = "environments/test"
  title       = "MongoDB Replica Set"
  engine      = "MONGODB"
  activation  = true

  data_sources {
    id                      = "admin"
    type                    = "ADMIN"
    host                    = "mongodb-1.example.com"
    port                    = "27017"
    username                = "admin"
    password                = "your-password"
    authentication_database = "admin"
    replica_set             = "rs0"
    additional_addresses {
      host = "mongodb-2.example.com"
      port = "27017"
    }
  }
}
```

### SSH Tunnel Connection

```terraform
resource "bytebase_instance" "mysql_ssh" {
  resource_id = "mysql-ssh"
  environment = "environments/test"
  title       = "MySQL via SSH"
  engine      = "MYSQL"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "127.0.0.1"
    port                = "3306"
    username            = "admin"
    password            = "your-password"
    authentication_type = "PASSWORD"
    ssh_host            = "bastion.example.com"
    ssh_port            = "22"
    ssh_user            = "ubuntu"
    ssh_private_key     = file("~/.ssh/id_rsa")
  }
}
```

## Query Instances

```terraform
# List all instances
data "bytebase_instance_list" "all" {
  environment = "environments/test"
  engines     = ["MYSQL", "POSTGRES"]
}

# Get specific instance
data "bytebase_instance" "mysql" {
  resource_id        = "mysql-example"
  list_all_databases = true
}
```

## Running the Examples

1. Update the provider configuration in `main.tf` with your Bytebase credentials
2. Run the [setup](../setup/) example first to create environments
3. Initialize and apply:

```bash
terraform init
terraform plan
terraform apply
```
