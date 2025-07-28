provider "aws" {
  region = local.region
}

variable "region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1" # optional fallback
}

locals {
  name   = "complete-s3-bucket"
}

module "s3_bucket" {
  source = "../../../modules/s3"

  bucket = local.name
  
  # Bucket configuration
  bucket_config = {
    force_destroy = true
    versioning_enabled = true
    intelligent_tiering = true
    
    lifecycle_rules = [
      {
        prefix = "logs/"
        enabled = true
        transition_days = 90
        storage_class = "GLACIER"
      }
    ]
  }

  # KMS configuration
  kms_key = {
    create = true
    deletion_window_in_days = 7
    enable_key_rotation = true
    key_administrators = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
  }

  # Logging configuration  
  logging = {
    mode = "create_new"
    retention_days = 90
    encryption_key = "create_new"
  }

  tags = {
    Environment = "dev"
    Owner       = "terraform"
    Project     = "complete-example"
  }
} 