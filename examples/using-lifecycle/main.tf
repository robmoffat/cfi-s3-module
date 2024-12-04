module "bucket_lifecycle" {
  source = "../../wrappers/lifecycle"

  lifecycle_rules = {
    "logs_cleanup" = {
      bucket = "my-logs-bucket"
      prefix = "logs/"
      
      transitions = [
        {
          days = 30
          storage_class = "STANDARD_IA"
        },
        {
          days = 90
          storage_class = "GLACIER"
        }
      ]
      
      expiration = {
        enabled = true
        days = 365
      }
      
      noncurrent_version_expiration = {
        enabled = true
        days = 90
      }
      
      abort_incomplete_multipart_upload = {
        enabled = true
        days_after_initiation = 7
      }
      
      tags = {
        Type = "Logs"
        Retention = "1-year"
      }
    }
  }
} 