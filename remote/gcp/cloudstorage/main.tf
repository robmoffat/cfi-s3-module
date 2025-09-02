module "cloud_storage" {
  source  = "git::https://github.com/terraform-google-modules/terraform-google-cloud-storage.git/examples/simple_bucket?ref=v11.1.0"
  project_id = "woven-precept-353210"
}
