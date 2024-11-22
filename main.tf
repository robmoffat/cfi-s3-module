module "secure_s3" {
  source = "./module/s3"
  
  prefix      = "prod"
  bucket_name = "my-application-data"
  
  tags = {
    Environment = "Production"
    Project     = "MyApp"
    Owner = "CFI"
  }

  kms_key = {
    create               = true
    deletion_window_in_days = 7
    enable_key_rotation    = true
    key_administrators    = ["arn:aws:iam::123456789012:user/admin"]
    key_users            = ["arn:aws:iam::123456789012:role/app-role"]
  }

  logging = {
    create_log_bucket = true
    enable_cloudwatch_logs = true
    cloudwatch_log_retention_days = 30
  }

  bucket_config = {
    versioning_enabled = true
    intelligent_tiering = true
    lifecycle_rules = [
      {
        prefix = "archive/"
        enabled = true
        transition_days = 90
        storage_class = "GLACIER"
      }
    ]
  }
}
