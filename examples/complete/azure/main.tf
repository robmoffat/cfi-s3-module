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

variable "storage_account_name" {
  description = "Name of the storage account (must be globally unique)"
  type        = string
  default     = ""
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
    managed-by  = "terraform"
  }
}

# Key Vault for storing storage account keys and secrets
resource "azurerm_key_vault" "this" {
  name                        = "kv-cfi-storage-${random_id.suffix.hex}"
  location                    = azurerm_resource_group.this.location
  resource_group_name         = azurerm_resource_group.this.name
  enabled_for_disk_encryption = true
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  sku_name = "standard"

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    key_permissions = [
      "Get", "List", "Create", "Delete", "Update", "Import", "Backup", "Restore", "Recover", "Purge"
    ]

    secret_permissions = [
      "Get", "List", "Set", "Delete", "Backup", "Restore", "Recover", "Purge"
    ]

    certificate_permissions = [
      "Get", "List", "Create", "Delete", "Update", "Import", "Backup", "Restore", "Recover", "Purge"
    ]
  }

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    managed-by  = "terraform"
  }
}

# Storage Account
resource "azurerm_storage_account" "this" {
  name                     = var.storage_account_name != "" ? var.storage_account_name : "stcfistorage${random_id.suffix.hex}"
  resource_group_name      = azurerm_resource_group.this.name
  location                 = azurerm_resource_group.this.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"

  # Enable advanced threat protection
  enable_https_traffic_only = true
  min_tls_version          = "TLS1_2"

  # Enable blob public access (disabled for security)
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

  # Network rules
  network_rules {
    default_action = "Deny"
    bypass         = ["AzureServices"]
    
    # Allow specific IP ranges if needed
    # ip_rules = ["1.2.3.4", "5.6.7.8"]
  }

  # Customer-managed key encryption
  identity {
    type = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.this.id]
  }

  key_vault_key_id = azurerm_key_vault_key.this.versionless_id

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    managed-by  = "terraform"
  }
}

# User-assigned managed identity
resource "azurerm_user_assigned_identity" "this" {
  name                = "id-cfi-storage-${random_id.suffix.hex}"
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    managed-by  = "terraform"
  }
}

# Key Vault key for storage account encryption
resource "azurerm_key_vault_key" "this" {
  name         = "key-storage-encryption"
  key_vault_id = azurerm_key_vault.this.id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey"
  ]

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    managed-by  = "terraform"
  }
}

# Access policy for the managed identity to access Key Vault
resource "azurerm_key_vault_access_policy" "storage" {
  key_vault_id = azurerm_key_vault.this.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_user_assigned_identity.this.principal_id

  key_permissions = [
    "Get", "UnwrapKey", "WrapKey"
  ]
}

# Storage container for logs
resource "azurerm_storage_container" "logs" {
  name                  = "logs"
  storage_account_name  = azurerm_storage_account.this.name
  container_access_type = "private"
}

# Storage container for data
resource "azurerm_storage_container" "data" {
  name                  = "data"
  storage_account_name  = azurerm_storage_account.this.name
  container_access_type = "private"
}

# Log Analytics workspace for monitoring
resource "azurerm_log_analytics_workspace" "this" {
  name                = "law-cfi-storage-${random_id.suffix.hex}"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    managed-by  = "terraform"
  }
}

# Diagnostic settings for storage account
resource "azurerm_monitor_diagnostic_setting" "storage" {
  name                       = "diag-storage-${random_id.suffix.hex}"
  target_resource_id         = azurerm_storage_account.this.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id

  log {
    category = "StorageRead"
    enabled  = true

    retention_policy {
      enabled = true
      days    = 30
    }
  }

  log {
    category = "StorageWrite"
    enabled  = true

    retention_policy {
      enabled = true
      days    = 30
    }
  }

  log {
    category = "StorageDelete"
    enabled  = true

    retention_policy {
      enabled = true
      days    = 30
    }
  }

  metric {
    category = "Transaction"
    enabled  = true

    retention_policy {
      enabled = true
      days    = 30
    }
  }

  metric {
    category = "Capacity"
    enabled  = true

    retention_policy {
      enabled = true
      days    = 30
    }
  }
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

output "storage_account_primary_access_key" {
  description = "The primary access key for the storage account"
  value       = azurerm_storage_account.this.primary_access_key
  sensitive   = true
}

output "storage_account_primary_connection_string" {
  description = "The primary connection string for the storage account"
  value       = azurerm_storage_account.this.primary_connection_string
  sensitive   = true
}

output "key_vault_name" {
  description = "The name of the Key Vault"
  value       = azurerm_key_vault.this.name
}

output "key_vault_id" {
  description = "The ID of the Key Vault"
  value       = azurerm_key_vault.this.id
}

output "log_analytics_workspace_id" {
  description = "The ID of the Log Analytics workspace"
  value       = azurerm_log_analytics_workspace.this.id
}
