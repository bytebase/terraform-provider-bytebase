# Built-in Test Instance
resource "bytebase_instance" "test" {
  depends_on  = [bytebase_setting.environments]
  resource_id = "test-sample-instance"
  environment = "environments/test"
  title       = "Test Sample Instance"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id       = "admin data source test-sample-instance"
    type     = "ADMIN"
    host     = "/tmp"
    port     = "8083"
    username = "bbsample"
    password = ""
  }
}

# Built-in Prod Instance
resource "bytebase_instance" "prod" {
  depends_on  = [bytebase_setting.environments]
  resource_id = "prod-sample-instance"
  environment = "environments/prod"
  title       = "Prod Sample Instance"
  engine      = "POSTGRES"
  activation  = true

  data_sources {
    id       = "admin data source prod-sample-instance"
    type     = "ADMIN"
    host     = "/tmp"
    port     = "8084"
    username = "bbsample"
    password = ""
  }
}