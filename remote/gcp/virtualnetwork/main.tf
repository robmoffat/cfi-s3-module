module "vpc" {
  source = "git::https://github.com/terraform-google-modules/terraform-google-network.git//examples/basic_auto_mode?ref=v11.1.1"
  project_id = var.project_id
}
