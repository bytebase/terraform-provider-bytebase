resource "bytebase_database" "database" {
  depends_on = [
    bytebase_instance.prod,
    bytebase_project.project-two,
    bytebase_setting.environments
  ]

  name        = "instances/prod-sample-instance/databases/hr_prod"
  project     = bytebase_project.project-two.name
  environment = bytebase_setting.environments.environment_setting[0].environment[1].name

  catalog {
    schemas {
      name = "public"
      tables {
        name = "salary"
        columns {
          name          = "from_date"
          semantic_type = "date-year-mask"
        }
        columns {
          name          = "amount"
          classification = "2-1"
        }
      }
    }
  }
}