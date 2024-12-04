module "s3_objects" {
  source = "../../wrappers/object"

  items = {
    "readme" = {
      bucket = "my-bucket"
      key    = "docs/README.md"
      source = "./files/README.md"
      content_type = "text/markdown"
      tags = {
        Type = "Documentation"
      }
    }
    "config" = {
      bucket = "my-bucket"
      key    = "config/settings.json"
      content = jsonencode({
        environment = "production"
        debug = false
      })
      content_type = "application/json"
      object_lock_mode = "GOVERNANCE"
      object_lock_retain_until_date = "2024-12-31T00:00:00Z"
    }
  }
} 