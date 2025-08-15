variable "location" {
  description = "Azure region to deploy into"
  type        = string
  default     = "East US" # optional fallback
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-cfi-storage"
}

provider "azurerm" {
  features {}
}

resource "random_id" "suffix" {
  byte_length = 4
}

data "azurerm_client_config" "current" {}

# Resource Group
resource "azurerm_resource_group" "this" {
  name     = var.resource_group_name
  location = var.location

  tags = {
    Environment = "Production"
    Owner       = "CFI"
  }
}

# Storage Account
resource "azurerm_storage_account" "this" {
  name                     = "stcfistorage${random_id.suffix.hex}"
  resource_group_name      = azurerm_resource_group.this.name
  location                 = azurerm_resource_group.this.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"

  # Security features
  enable_https_traffic_only = true
  min_tls_version          = "TLS1_2"
  allow_nested_items_to_be_public = false

  # Enable versioning
  blob_properties {
    versioning_enabled = true
    
    # Lifecycle management
    delete_retention_policy {
      days = 90
    }
    
    # Soft delete for blobs
    container_delete_retention_policy {
      days = 7
    }
  }

  tags = {
    Environment = "Production"
    Owner       = "CFI"
  }
}

# Storage container for data
resource "azurerm_storage_container" "data" {
  name                  = "data"
  storage_account_name  = azurerm_storage_account.this.name
  container_access_type = "private"
}

# Outputs
output "storage_account_name" {
  description = "The name of the storage account"
  value       = azurerm_storage_account.this.name
}

output "storage_account_id" {
  description = "The ID of the storage account"
  value       = azurerm_storage_account.this.id
}
