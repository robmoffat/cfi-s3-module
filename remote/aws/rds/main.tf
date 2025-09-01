module "rds" {
  source  = "git::https://github.com/terraform-aws-modules/terraform-aws-rds-aurora.git/examples/postgresql?ref=v9.15.0"
  version = "0.0.7"
}
