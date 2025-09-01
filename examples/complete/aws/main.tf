variable "region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1" # optional fallback
}

provider "aws" {
  region = var.region
}

resource "random_id" "suffix" {
  byte_length = 4
}

data "aws_caller_identity" "current" {}

module "bedrock" {
  source  = "aws-ia/bedrock/aws"
  version = "0.0.29"
  
  # Basic Bedrock configuration
  model_invocation_logging_configuration = {
    logging_config = {
      cloudwatch_config = {
        log_group_name = "bedrock-logs-${random_id.suffix.hex}"
        role_arn       = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/BedrockLoggingRole"
      }
      s3_config = {
        bucket_name = "bedrock-logs-${random_id.suffix.hex}"
        prefix      = "bedrock/"
      }
    }
  }

  tags = {
    Environment = "Production"
    Owner       = "CFI"
    Module      = "bedrock"
  }
}