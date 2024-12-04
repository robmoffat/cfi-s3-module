variable "bucket_config" {
  description = <<-EOT
    Configuration for the S3 bucket behavior and features.
    
    force_destroy          - (Optional) Allow bucket to be destroyed even with content. Default: false
    versioning_enabled     - (Optional) Enable versioning for objects. Default: true
    mfa_delete            - (Optional) Require MFA for delete operations. Default: true
    retention_days        - (Optional) Default retention period for objects. Default: 90
    max_object_size       - (Optional) Maximum allowed object size in bytes. Default: 5368709120 (5GB)
    intelligent_tiering   - (Optional) Enable S3 Intelligent-Tiering. Default: false
    
    lifecycle_rules       - (Optional) List of lifecycle rules. Each rule supports:
      prefix            - Path prefix identifying objects to which rule applies
      enabled           - Whether rule is enabled
      expiration_days   - Days until objects expire
      transition_days   - Days until objects transition
      storage_class     - AWS storage class to transition objects to
      
    replication_config   - (Optional) Cross-region replication configuration:
      enabled           - Whether to enable replication
      destination_bucket - ARN of destination bucket
      destination_region - Region of destination bucket
      storage_class     - Storage class for replicated objects
      
    cors_rules          - (Optional) CORS rules for the bucket:
      allowed_headers   - List of allowed headers
      allowed_methods   - List of allowed HTTP methods
      allowed_origins   - List of allowed origins
      expose_headers    - List of exposed headers
      max_age_seconds   - Cache time for CORS rules
    
    Example:
    ```hcl
    bucket_config = {
      force_destroy = false
      versioning_enabled = true
      mfa_delete = true
      retention_days = 90
      intelligent_tiering = true
      
      lifecycle_rules = [
        {
          prefix = "logs/"
          enabled = true
          expiration_days = 90
          transition_days = 30
          storage_class = "STANDARD_IA"
        }
      ]
      
      replication_config = {
        enabled = true
        destination_bucket = "arn:aws:s3:::dest-bucket"
        destination_region = "eu-west-1"
        storage_class = "STANDARD"
      }
      
      cors_rules = [
        {
          allowed_headers = ["*"]
          allowed_methods = ["GET", "HEAD"]
          allowed_origins = ["https://example.com"]
          expose_headers  = ["ETag"]
          max_age_seconds = 3000
        }
      ]
    }
    ```
  EOT
  
  type = object({
    force_destroy          = optional(bool, false)
    versioning_enabled     = optional(bool, true)
    mfa_delete            = optional(bool, true)
    retention_days        = optional(number, 90)
    max_object_size       = optional(number, 5368709120)
    intelligent_tiering   = optional(bool, false)
    
    lifecycle_rules       = optional(list(object({
      prefix            = string
      enabled           = bool
      expiration_days   = optional(number)
      transition_days   = optional(number)
      storage_class     = optional(string)
    })), [])
    
    replication_config   = optional(object({
      enabled           = bool
      destination_bucket = string
      destination_region = string
      storage_class     = optional(string, "STANDARD")
    }))
    
    cors_rules          = optional(list(object({
      allowed_headers   = optional(list(string), ["*"])
      allowed_methods   = list(string)
      allowed_origins   = list(string)
      expose_headers    = optional(list(string), [])
      max_age_seconds   = optional(number, 3600)
    })), [])
  })
  
  default = {}

  validation {
    condition     = var.bucket_config.retention_days == null || var.bucket_config.retention_days >= 1
    error_message = "retention_days must be at least 1 day"
  }

  validation {
    condition     = var.bucket_config.max_object_size == null || var.bucket_config.max_object_size >= 0
    error_message = "max_object_size must be a positive number"
  }

  validation {
    condition = alltrue([
      for rule in coalesce(var.bucket_config.lifecycle_rules, []) :
      rule.storage_class == null || contains([
        "STANDARD_IA",
        "ONEZONE_IA",
        "GLACIER",
        "DEEP_ARCHIVE",
        "INTELLIGENT_TIERING"
      ], rule.storage_class)
    ])
    error_message = "Invalid storage class specified in lifecycle rules"
  }

  validation {
    condition = alltrue([
      for rule in coalesce(var.bucket_config.cors_rules, []) :
      alltrue([
        for method in rule.allowed_methods :
        contains(["GET", "PUT", "POST", "DELETE", "HEAD"], method)
      ])
    ])
    error_message = "Invalid HTTP method in CORS rules"
  }
}

# variables.tf

variable "prefix" {
  description = "Prefix to be used for all created resources. Should be RFC 1123 compliant."
  type        = string
  default     = ""

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*[a-z0-9]$", var.prefix)) || var.prefix == ""
    error_message = "Prefix must be RFC 1123 compliant: contain only lowercase alphanumeric characters or '-', start with alphanumeric, end with alphanumeric."
  }
}

variable "tags" {
  description = "A map of tags to be applied to all resources. Must include required tags per security standards."
  type        = map(string)
  default     = {}

  validation {
    condition     = contains(keys(var.tags), "Environment") && contains(keys(var.tags), "Owner")
    error_message = "Tags must include 'Environment' and 'Owner' as required by security standards."
  }
}

variable "bucket_name" {
  description = "Name of the S3 bucket. Must comply with AWS S3 naming rules and not exceed 63 characters including prefix."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]*[a-z0-9]$", var.bucket_name))
    error_message = "Bucket name must comply with AWS S3 naming rules: lowercase alphanumeric characters, dots, and hyphens."
  }
}

# Example of a complex variable with comprehensive documentation
variable "kms_key" {
  description = <<-EOT
    Configuration for KMS key. Either create a new key or use an existing one.
    
    create        - Whether to create a new KMS key
    key_arn      - ARN of existing KMS key if create is false
    deletion_window_in_days - Duration in days before key is deleted (7-30 days)
    enable_key_rotation    - Whether to enable automatic key rotation
    key_administrators    - List of IAM ARNs that can administer the key
    key_users            - List of IAM ARNs that can use the key
    
    Example:
    ```hcl
    kms_key = {
      create = true
      deletion_window_in_days = 7
      enable_key_rotation = true
      key_administrators = ["arn:aws:iam::123456789012:user/admin"]
      key_users = ["arn:aws:iam::123456789012:role/app-role"]
    }
    ```
  EOT
  type = object({
    create                  = bool
    key_arn                = optional(string)
    deletion_window_in_days = optional(number, 7)
    enable_key_rotation    = optional(bool, true)
    key_administrators    = optional(list(string), [])
    key_users            = optional(list(string), [])
  })

  validation {
    condition     = (var.kms_key.deletion_window_in_days >= 7 && var.kms_key.deletion_window_in_days <= 30) || var.kms_key.deletion_window_in_days == null
    error_message = "deletion_window_in_days must be between 7 and 30 days"
  }
}

variable "logging" {
  description = <<-EOT
    Configuration for bucket logging and monitoring.
    
    Mode can be one of:
    - "create_new"    : Creates new logging resources
    - "use_existing"  : Uses existing logging resources
    - "disabled"      : Disables logging
    
    When mode = "use_existing":
    - bucket_arn is required
    - bucket_name is required
    - log_prefix is optional (defaults to "s3-logs/")
    
    When mode = "create_new":
    - retention_days is optional (defaults to 90)
    - encryption_key can be either "create_new" or an existing KMS key ARN
    
    Example (Create New):
    ```hcl
    logging = {
      mode = "create_new"
      retention_days = 90
      encryption_key = "create_new"
    }
    ```
    
    Example (Use Existing):
    ```hcl
    logging = {
      mode = "use_existing"
      bucket_arn = "arn:aws:s3:::existing-logs"
      bucket_name = "existing-logs"
      log_prefix = "custom-prefix/"
    }
    ```
  EOT
  
  type = object({
    mode = string
    # For use_existing mode
    bucket_arn = optional(string)
    bucket_name = optional(string)
    log_prefix = optional(string, "s3-logs/")
    # For create_new mode
    retention_days = optional(number, 90)
    encryption_key = optional(string, "create_new")
  })

  validation {
    condition = contains(["create_new", "use_existing", "disabled"], var.logging.mode)
    error_message = "logging.mode must be one of: create_new, use_existing, disabled"
  }

  validation {
    condition = (
      var.logging.mode != "use_existing" || 
      (var.logging.bucket_arn != null && var.logging.bucket_name != null)
    )
    error_message = "When logging.mode = use_existing, bucket_arn and bucket_name are required"
  }
}
