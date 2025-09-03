variable "region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1" # optional fallback
}

provider "aws" {
  region = var.region
}


data "aws_caller_identity" "current" {}

module "secure_s3" {
  source      = "../../../modules/s3"
  prefix      = "prod"
  bucket_name = "my-secure-s3-bucket-20250423"

  kms_key = {
    create = true
    deletion_window_in_days = 7
    enable_key_rotation = true
    key_administrators = []
    key_users = [
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:user/terraform-user",
      "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
    ]
  }


  logging = {
    mode = "create_new"
    enable_cloudwatch_logs = true
    retention_days = 30
    encryption_key = "create_new"
  }

  tags = {
    Environment = "Production"
    Owner       = "CFI"
  }
}