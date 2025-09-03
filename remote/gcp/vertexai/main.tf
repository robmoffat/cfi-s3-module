module "vertex_ai" {
  source  = "git::https://github.com/GoogleCloudPlatform/terraform-google-vertex-ai.git//examples/workbench-simple-example?ref=v2.1.1"
  project_id = "woven-precept-353210"
}
