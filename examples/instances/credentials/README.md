# GCP Service Account (Mock)

This is a **mock GCP service account JSON for example purposes only**. Do not use in production.

## Files

- `service-account.json` - GCP service account credentials

## Usage

Replace this file with your actual GCP service account JSON:

```bash
# Download from GCP Console or use gcloud CLI
gcloud iam service-accounts keys create ./service-account.json \
  --iam-account=terraform@my-project.iam.gserviceaccount.com
```

## Required Permissions

For Cloud SQL IAM authentication, the service account needs:
- `roles/cloudsql.client` - Cloud SQL Client role
- `roles/cloudsql.instanceUser` - Cloud SQL Instance User role

For Spanner:
- `roles/spanner.databaseUser` - Spanner Database User role

For BigQuery:
- `roles/bigquery.user` - BigQuery User role
- `roles/bigquery.dataViewer` - BigQuery Data Viewer role

## Security Note

Never commit real service account keys to version control. Consider using:
- Workload Identity Federation
- Environment variables (GOOGLE_APPLICATION_CREDENTIALS)
- Terraform variables with sensitive flag
