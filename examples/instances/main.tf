# Examples for creating instances with different authentication types and engine-specific configurations
terraform {
  required_providers {
    bytebase = {
      version = "3.14.0"
      # For local development, please use "terraform.local/bytebase/bytebase" instead
      source = "registry.terraform.io/bytebase/bytebase"
    }
  }
}

provider "bytebase" {
  # You need to replace the account and key with your Bytebase service account.
  service_account = "terraform@service.bytebase.com"
  service_key     = "bbs_BxVIp7uQsARl8nR92ZZV"
  # The Bytebase service URL. You can use the external URL in production.
  # Check the docs about external URL: https://www.bytebase.com/docs/get-started/install/external-url
  url = "https://bytebase.example.com"
}

###############################################################################
# Example 1: PASSWORD authentication (default)
# This is the default authentication type for most database engines.
###############################################################################
resource "bytebase_instance" "mysql_password" {
  resource_id = "mysql-password-example"
  environment = "environments/test"
  title       = "MySQL with Password Auth"
  engine      = "MYSQL"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "mysql.example.com"
    port     = "3306"
    username = "admin"
    password = "your-password"
    # authentication_type defaults to PASSWORD
  }
}

###############################################################################
# Example 2: PASSWORD with SSL/TLS
# Enable SSL connection with certificate verification.
###############################################################################
resource "bytebase_instance" "mysql_ssl" {
  resource_id = "mysql-ssl-example"
  environment = "environments/test"
  title       = "MySQL with SSL"
  engine      = "MYSQL"
  activation  = true

  data_sources {
    id                     = "admin"
    type                   = "ADMIN"
    host                   = "mysql.example.com"
    port                   = "3306"
    username               = "admin"
    password               = "your-password"
    use_ssl                = true
    verify_tls_certificate = true
    ssl_ca                 = file("${path.module}/certs/ca.pem")
    ssl_cert               = file("${path.module}/certs/client-cert.pem")
    ssl_key                = file("${path.module}/certs/client-key.pem")
  }
}

###############################################################################
# Example 3: PASSWORD with SSH Tunnel
# Connect through an SSH tunnel.
# Only available for MYSQL, TIDB, MARIADB, OCEANBASE, POSTGRES, REDIS
# with PASSWORD authentication.
###############################################################################
resource "bytebase_instance" "mysql_ssh_tunnel" {
  resource_id = "mysql-ssh-tunnel-example"
  environment = "environments/test"
  title       = "MySQL via SSH Tunnel"
  engine      = "MYSQL"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "127.0.0.1" # Database host as seen from SSH server
    port                = "3306"
    username            = "admin"
    password            = "your-password"
    authentication_type = "PASSWORD" # SSH tunnel requires PASSWORD auth
    # SSH tunnel configuration
    ssh_host        = "bastion.example.com"
    ssh_port        = "22"
    ssh_user        = "ubuntu"
    ssh_private_key = file("${path.module}/ssh/id_rsa")
    # Or use ssh_password instead of ssh_private_key
    # ssh_password = "ssh-password"
  }
}

###############################################################################
# Example 4: PASSWORD with External Secret (Vault)
# Use HashiCorp Vault to manage database credentials.
# Requires instance license and PASSWORD authentication.
###############################################################################
resource "bytebase_instance" "postgres_vault" {
  resource_id = "postgres-vault-example"
  environment = "environments/test"
  title       = "PostgreSQL with Vault"
  engine      = "POSTGRES"
  activation  = true # Required for external_secret

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "postgres.example.com"
    port                = "5432"
    username            = "admin"
    authentication_type = "PASSWORD" # external_secret requires PASSWORD auth
    external_secret {
      vault {
        url               = "https://vault.example.com:8200"
        token             = "hvs.your-vault-token"
        engine_name       = "secret"
        secret_name       = "database/postgres"
        password_key_name = "password"
      }
    }
  }
}

###############################################################################
# Example 5: PASSWORD with External Secret (AWS Secrets Manager)
###############################################################################
resource "bytebase_instance" "postgres_aws_secrets" {
  resource_id = "postgres-aws-secrets-example"
  environment = "environments/test"
  title       = "PostgreSQL with AWS Secrets Manager"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "postgres.example.com"
    port                = "5432"
    username            = "admin"
    authentication_type = "PASSWORD"
    external_secret {
      aws_secrets_manager {
        secret_name       = "prod/database/postgres"
        password_key_name = "password"
      }
    }
  }
}

###############################################################################
# Example 6: PASSWORD with External Secret (GCP Secret Manager)
###############################################################################
resource "bytebase_instance" "postgres_gcp_secrets" {
  resource_id = "postgres-gcp-secrets-example"
  environment = "environments/test"
  title       = "PostgreSQL with GCP Secret Manager"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "postgres.example.com"
    port                = "5432"
    username            = "admin"
    authentication_type = "PASSWORD"
    external_secret {
      gcp_secret_manager {
        secret_name = "projects/my-project/secrets/db-password"
      }
    }
  }
}

###############################################################################
# Example 7: AWS_RDS_IAM authentication
# Use AWS IAM authentication for RDS databases.
# The 'region' field is only available with AWS_RDS_IAM.
###############################################################################
resource "bytebase_instance" "postgres_aws_iam" {
  resource_id = "postgres-aws-iam-example"
  environment = "environments/test"
  title       = "PostgreSQL with AWS IAM Auth"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "mydb.123456789012.us-east-1.rds.amazonaws.com"
    port                = "5432"
    database            = "postgres"
    username            = "iam_user"
    authentication_type = "AWS_RDS_IAM"
    region              = "us-east-1" # Only available with AWS_RDS_IAM
    # Optional: Provide AWS credentials explicitly
    # If not provided, uses default credential chain (env vars, IAM role, etc.)
    aws_credential {
      access_key_id     = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      # Optional: For cross-account access
      # role_arn    = "arn:aws:iam::123456789012:role/RDSAccessRole"
      # external_id = "unique-external-id"
    }
  }
}

###############################################################################
# Example 8: AWS_RDS_IAM for Elasticsearch
# Elasticsearch supports PASSWORD or AWS_RDS_IAM authentication.
###############################################################################
resource "bytebase_instance" "elasticsearch_aws_iam" {
  resource_id = "elasticsearch-aws-iam-example"
  environment = "environments/test"
  title       = "Elasticsearch with AWS IAM"
  engine      = "ELASTICSEARCH"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "search-mydomain.us-east-1.es.amazonaws.com"
    port                = "443"
    authentication_type = "AWS_RDS_IAM"
    region              = "us-east-1"
    # AWS credentials are optional - uses default chain if not provided
  }
}

###############################################################################
# Example 9: AZURE_IAM authentication
# Use Azure Active Directory authentication for supported databases.
# MSSQL supports PASSWORD or AZURE_IAM; COSMOSDB only supports AZURE_IAM.
###############################################################################
resource "bytebase_instance" "mssql_azure_iam" {
  resource_id = "mssql-azure-iam-example"
  environment = "environments/test"
  title       = "SQL Server with Azure IAM"
  engine      = "MSSQL"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "myserver.database.windows.net"
    port                = "1433"
    database            = "mydb"
    authentication_type = "AZURE_IAM"
    # Optional: Provide Azure credentials explicitly
    azure_credential {
      tenant_id     = "00000000-0000-0000-0000-000000000000"
      client_id     = "11111111-1111-1111-1111-111111111111"
      client_secret = "your-client-secret"
    }
  }
}

###############################################################################
# Example 10: AZURE_IAM for CosmosDB
# CosmosDB ONLY supports AZURE_IAM authentication.
###############################################################################
resource "bytebase_instance" "cosmosdb" {
  resource_id = "cosmosdb-example"
  environment = "environments/test"
  title       = "CosmosDB Instance"
  engine      = "COSMOSDB"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "myaccount.documents.azure.com"
    port                = "443"
    authentication_type = "AZURE_IAM" # CosmosDB requires AZURE_IAM
    azure_credential {
      tenant_id     = "00000000-0000-0000-0000-000000000000"
      client_id     = "11111111-1111-1111-1111-111111111111"
      client_secret = "your-client-secret"
    }
  }
}

###############################################################################
# Example 11: GOOGLE_CLOUD_SQL_IAM authentication
# Use Google Cloud IAM for Cloud SQL or other GCP database services.
###############################################################################
resource "bytebase_instance" "cloudsql_gcp_iam" {
  resource_id = "cloudsql-gcp-iam-example"
  environment = "environments/test"
  title       = "Cloud SQL with GCP IAM"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "10.0.0.1" # Private IP or Cloud SQL Auth Proxy
    port                = "5432"
    database            = "postgres"
    username            = "service-account@project.iam"
    authentication_type = "GOOGLE_CLOUD_SQL_IAM"
    # Optional: Provide service account JSON
    gcp_credential {
      content = file("${path.module}/credentials/service-account.json")
    }
  }
}

###############################################################################
# Example 12: GOOGLE_CLOUD_SQL_IAM for Spanner
# Spanner ONLY supports GOOGLE_CLOUD_SQL_IAM authentication.
###############################################################################
resource "bytebase_instance" "spanner" {
  resource_id = "spanner-example"
  environment = "environments/test"
  title       = "Cloud Spanner Instance"
  engine      = "SPANNER"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "projects/my-project/instances/my-instance"
    database            = "my-database"
    authentication_type = "GOOGLE_CLOUD_SQL_IAM" # Spanner requires GCP IAM
    gcp_credential {
      content = file("${path.module}/credentials/service-account.json")
    }
  }
}

###############################################################################
# Example 13: GOOGLE_CLOUD_SQL_IAM for BigQuery
# BigQuery ONLY supports GOOGLE_CLOUD_SQL_IAM authentication.
###############################################################################
resource "bytebase_instance" "bigquery" {
  resource_id = "bigquery-example"
  environment = "environments/test"
  title       = "BigQuery Instance"
  engine      = "BIGQUERY"
  activation  = true

  data_sources {
    id                  = "admin"
    type                = "ADMIN"
    host                = "my-gcp-project"
    authentication_type = "GOOGLE_CLOUD_SQL_IAM" # BigQuery requires GCP IAM
    gcp_credential {
      content = file("${path.module}/credentials/service-account.json")
    }
  }
}

###############################################################################
# Example 14: MongoDB with engine-specific fields
# MongoDB-specific fields: srv, authentication_database, replica_set,
# direct_connection, additional_addresses
###############################################################################
resource "bytebase_instance" "mongodb" {
  resource_id = "mongodb-example"
  environment = "environments/test"
  title       = "MongoDB Instance"
  engine      = "MONGODB"
  activation  = true

  data_sources {
    id                      = "admin"
    type                    = "ADMIN"
    host                    = "mongodb.example.com"
    port                    = "27017"
    username                = "admin"
    password                = "your-password"
    authentication_database = "admin" # MongoDB-specific: auth database
    # MongoDB replica set configuration
    replica_set       = "rs0"
    direct_connection = false
    additional_addresses {
      host = "mongodb-2.example.com"
      port = "27017"
    }
    additional_addresses {
      host = "mongodb-3.example.com"
      port = "27017"
    }
  }
}

###############################################################################
# Example 15: MongoDB with SRV record
# Use DNS SRV record for MongoDB Atlas or similar services.
###############################################################################
resource "bytebase_instance" "mongodb_srv" {
  resource_id = "mongodb-srv-example"
  environment = "environments/test"
  title       = "MongoDB Atlas"
  engine      = "MONGODB"
  activation  = true

  data_sources {
    id                      = "admin"
    type                    = "ADMIN"
    host                    = "cluster0.mongodb.net" # SRV hostname
    username                = "admin"
    password                = "your-password"
    srv                     = true # Use DNS SRV record
    authentication_database = "admin"
  }
}

###############################################################################
# Example 16: Oracle with engine-specific fields
# Oracle-specific fields: sid, service_name
###############################################################################
resource "bytebase_instance" "oracle_sid" {
  resource_id = "oracle-sid-example"
  environment = "environments/test"
  title       = "Oracle with SID"
  engine      = "ORACLE"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "oracle.example.com"
    port     = "1521"
    username = "system"
    password = "your-password"
    sid      = "ORCL" # Oracle SID
  }
}

resource "bytebase_instance" "oracle_service" {
  resource_id = "oracle-service-example"
  environment = "environments/test"
  title       = "Oracle with Service Name"
  engine      = "ORACLE"
  activation  = true

  data_sources {
    id           = "admin"
    type         = "ADMIN"
    host         = "oracle.example.com"
    port         = "1521"
    username     = "system"
    password     = "your-password"
    service_name = "pdb1.example.com" # Oracle service name
  }
}

###############################################################################
# Example 17: Redis with engine-specific fields
# Redis-specific fields: redis_type, master_name, master_username, master_password
###############################################################################
resource "bytebase_instance" "redis_standalone" {
  resource_id = "redis-standalone-example"
  environment = "environments/test"
  title       = "Redis Standalone"
  engine      = "REDIS"
  activation  = true

  data_sources {
    id         = "admin"
    type       = "ADMIN"
    host       = "redis.example.com"
    port       = "6379"
    password   = "your-redis-password"
    redis_type = "STANDALONE"
  }
}

resource "bytebase_instance" "redis_sentinel" {
  resource_id = "redis-sentinel-example"
  environment = "environments/test"
  title       = "Redis Sentinel"
  engine      = "REDIS"
  activation  = true

  data_sources {
    id         = "admin"
    type       = "ADMIN"
    host       = "sentinel-1.example.com"
    port       = "26379"
    redis_type = "SENTINEL"
    # Sentinel-specific configuration
    master_name     = "mymaster"
    master_username = "default"
    master_password = "master-password"
  }
}

resource "bytebase_instance" "redis_cluster" {
  resource_id = "redis-cluster-example"
  environment = "environments/test"
  title       = "Redis Cluster"
  engine      = "REDIS"
  activation  = true

  data_sources {
    id         = "admin"
    type       = "ADMIN"
    host       = "redis-node-1.example.com"
    port       = "6379"
    password   = "cluster-password"
    redis_type = "CLUSTER"
  }
}

###############################################################################
# Example 18: Databricks with engine-specific field
# Databricks-specific field: warehouse_id
###############################################################################
resource "bytebase_instance" "databricks" {
  resource_id = "databricks-example"
  environment = "environments/test"
  title       = "Databricks Instance"
  engine      = "DATABRICKS"
  activation  = true

  data_sources {
    id           = "admin"
    type         = "ADMIN"
    host         = "dbc-12345678-abcd.cloud.databricks.com"
    port         = "443"
    username     = "token"
    password     = "dapi1234567890abcdef" # Databricks personal access token
    warehouse_id = "abc123def456"         # SQL warehouse ID
  }
}

###############################################################################
# Example 19: CockroachDB with engine-specific field
# CockroachDB-specific field: cluster
###############################################################################
resource "bytebase_instance" "cockroachdb" {
  resource_id = "cockroachdb-example"
  environment = "environments/test"
  title       = "CockroachDB Cloud"
  engine      = "COCKROACHDB"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "free-tier.gcp-us-central1.cockroachlabs.cloud"
    port     = "26257"
    database = "defaultdb"
    username = "user"
    password = "your-password"
    cluster  = "my-cluster-123" # CockroachDB cluster name
    use_ssl  = true
  }
}

###############################################################################
# Example 20: Hive with SASL/Kerberos authentication
# SASL config is only available for HIVE engine.
###############################################################################
resource "bytebase_instance" "hive_kerberos" {
  resource_id = "hive-kerberos-example"
  environment = "environments/test"
  title       = "Hive with Kerberos"
  engine      = "HIVE"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "hive.example.com"
    port     = "10000"
    database = "default"
    sasl_config {
      kerberos {
        primary                = "hive"
        instance               = "hive.example.com"
        realm                  = "EXAMPLE.COM"
        keytab                 = base64encode(file("${path.module}/kerberos/hive.keytab"))
        kdc_host               = "kdc.example.com"
        kdc_port               = "88"
        kdc_transport_protocol = "tcp"
      }
    }
  }
}

###############################################################################
# Example 21: Instance with extra connection parameters
# Custom connection parameters for specific driver requirements.
###############################################################################
resource "bytebase_instance" "postgres_extra_params" {
  resource_id = "postgres-extra-params-example"
  environment = "environments/test"
  title       = "PostgreSQL with Extra Params"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "postgres.example.com"
    port     = "5432"
    database = "mydb"
    username = "admin"
    password = "your-password"
    extra_connection_parameters = {
      "sslmode"          = "verify-full"
      "connect_timeout"  = "10"
      "application_name" = "bytebase"
    }
  }
}

###############################################################################
# Example 22: DynamoDB
# DynamoDB doesn't require host or authentication credentials.
###############################################################################
resource "bytebase_instance" "dynamodb" {
  resource_id = "dynamodb-example"
  environment = "environments/test"
  title       = "DynamoDB Instance"
  engine      = "DYNAMODB"
  activation  = true

  data_sources {
    id   = "admin"
    type = "ADMIN"
  }
}

###############################################################################
# Example 23: Trino (distributed SQL query engine)
###############################################################################
resource "bytebase_instance" "trino" {
  resource_id = "trino-example"
  environment = "environments/test"
  title       = "Trino Instance"
  engine      = "TRINO"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "trino.example.com"
    port     = "8080"
    username = "admin"
    password = "your-password"
  }
}

###############################################################################
# Example 24: Cassandra
###############################################################################
resource "bytebase_instance" "cassandra" {
  resource_id = "cassandra-example"
  environment = "environments/test"
  title       = "Cassandra Instance"
  engine      = "CASSANDRA"
  activation  = true

  data_sources {
    id       = "admin"
    type     = "ADMIN"
    host     = "cassandra.example.com"
    port     = "9042"
    username = "cassandra"
    password = "your-password"
  }
}

###############################################################################
# Data Sources - Query existing instances
###############################################################################

List all instances in an environment
data "bytebase_instance_list" "all" {
  environment = "environments/test"
  engines = [
    "MYSQL",
    "POSTGRES"
  ]
}

output "all_instances" {
  value = data.bytebase_instance_list.all
}

# Get a specific instance
data "bytebase_instance" "mysql" {
  resource_id        = "mysql-password-example"
  list_all_databases = true
}

output "mysql_instance" {
  value = data.bytebase_instance.mysql
}
