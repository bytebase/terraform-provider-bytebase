terraform {
  required_version = ">= 1.11"
  required_providers {
    bytebase = {
      source  = "bytebase/bytebase"
      version = "3.25.1"
    }
  }
}

provider "bytebase" {}

variable "sharepoint_tenant_id" {
  type        = string
  description = "The Microsoft Entra tenant ID."
}

variable "sharepoint_client_id" {
  type        = string
  description = "The Microsoft Entra application ID."
}

variable "sharepoint_client_secret" {
  type        = string
  sensitive   = true
  ephemeral   = true
  description = "The application secret. It is not stored in Terraform state or plans."
}

variable "sharepoint_client_secret_version" {
  type        = number
  default     = 1
  description = "A local positive rotation marker. Increment whenever the secret changes."

  validation {
    condition     = var.sharepoint_client_secret_version >= 1 && floor(var.sharepoint_client_secret_version) == var.sharepoint_client_secret_version
    error_message = "The secret version must be a positive integer."
  }
}

variable "sharepoint_destination_url" {
  type        = string
  description = "The SharePoint document library or folder URL receiving exports."
}

variable "project_resource_id" {
  type        = string
  default     = "sharepoint-example"
  description = "The resource ID of the project created by this example."
}

resource "bytebase_setting" "export" {
  name = "settings/EXPORT"

  export {
    sharepoint {
      tenant_id             = var.sharepoint_tenant_id
      client_id             = var.sharepoint_client_id
      client_secret         = var.sharepoint_client_secret
      client_secret_version = var.sharepoint_client_secret_version
    }
  }
}

resource "bytebase_project" "export" {
  resource_id = var.project_resource_id
  title       = "SharePoint Example"

  export_setting {
    sharepoint {
      url = var.sharepoint_destination_url
    }
  }

  depends_on = [bytebase_setting.export]
}

data "bytebase_setting" "export" {
  name = bytebase_setting.export.name
}

output "sharepoint_application" {
  value = {
    tenant_id         = try(data.bytebase_setting.export.export[0].sharepoint[0].tenant_id, null)
    client_id         = try(data.bytebase_setting.export.export[0].sharepoint[0].client_id, null)
    client_secret_set = try(data.bytebase_setting.export.export[0].sharepoint[0].client_secret_set, false)
  }
}

output "sharepoint_destination_url" {
  value = try(bytebase_project.export.export_setting[0].sharepoint[0].url, null)
}
