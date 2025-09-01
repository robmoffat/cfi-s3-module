variable "region" {
  description = "AWS region"
  type        = string
}

module "vpc" {
  source  = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git//examples/complete?ref=v6.0.1" 
}
