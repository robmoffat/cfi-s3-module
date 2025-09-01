# data.tf
data "aws_caller_identity" "current" {}

locals {
  current_region = "us-east-1"
}

