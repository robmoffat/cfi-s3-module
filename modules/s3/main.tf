
# main.tf

locals {
  replication_enabled = try(var.bucket_config.replication_config.enabled, false)
  bucket_name     = var.prefix != "" ? "${var.prefix}-${var.bucket_name}" : var.bucket_name
  
  log_bucket_name = var.logging.mode == "create_new" ? (
    "${local.bucket_name}-logs"
  ) : var.logging.mode == "use_existing" ? (
    var.logging.bucket_name
  ) : null

  cloudwatch_log_group_name = var.logging.mode != "disabled" ? (
    "/aws/s3/${local.bucket_name}"
  ) : null

  common_tags = merge(
    var.tags,
    {
      "managed-by" = "terraform"
      "module"     = "secure-s3"
    }
  )
}

# KMS key creation if enabled
resource "aws_kms_key" "this" {
  count = var.kms_key.create ? 1 : 0

  description             = "KMS key for ${local.bucket_name} S3 bucket"
  deletion_window_in_days = var.kms_key.deletion_window_in_days
  enable_key_rotation     = var.kms_key.enable_key_rotation
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = concat([
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = concat(
            ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"],
            var.kms_key.key_administrators
          )
        }
        Action   = "kms:*"
        Resource = "*"
      }
    ],
    var.kms_key.key_users != null ? [
      {
        Sid    = "Allow Key Users"
        Effect = "Allow"
        Principal = {
          AWS = var.kms_key.key_users
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey"
        ]
        Resource = "*",
        Condition = {
          ArnLike = {
             "kms:EncryptionContext:aws:logs:arn" = "arn:aws:logs:us-east-1:232348204608:log-group:/aws/s3/*"
          }
        }
      }
    ] : [])
  })

  tags = local.common_tags
}

resource "aws_kms_alias" "this" {
  count = var.kms_key.create ? 1 : 0

  name          = "alias/${local.bucket_name}"
  target_key_id = aws_kms_key.this[0].key_id
}

# Log bucket if enabled
resource "aws_s3_bucket" "logs" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = local.log_bucket_name
  force_destroy = false

  tags = merge(local.common_tags, {
    Name = local.log_bucket_name
    Type = "logs"
  })
}

resource "aws_s3_bucket_versioning" "logs" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id
  versioning_configuration {
    status = "Enabled"
    mfa_delete = "Disabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn
      sse_algorithm     = "aws:kms"
    }
  }
}

# CloudWatch Log Group if enabled
resource "aws_cloudwatch_log_group" "this" {
  count = var.logging.mode != "disabled" ? 1 : 0

  name              = local.cloudwatch_log_group_name
  retention_in_days = var.logging.retention_days
  kms_key_id       = var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn

  tags = local.common_tags
}

# Main S3 bucket
resource "aws_s3_bucket" "main" {
  bucket = local.bucket_name
  force_destroy = var.bucket_config.force_destroy

  tags = merge(local.common_tags, {
    Name = local.bucket_name
  })
}

resource "aws_s3_bucket_versioning" "main" {
  bucket = aws_s3_bucket.main.id
  versioning_configuration {
    status     = var.bucket_config.versioning_enabled ? "Enabled" : "Disabled"
    mfa_delete = var.bucket_config.mfa_delete ? "Enabled" : "Disabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_logging" "main" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = aws_s3_bucket.main.id

  target_bucket = aws_s3_bucket.logs[0].id
  target_prefix = "logs/"
}

resource "aws_s3_bucket_public_access_block" "main" {
  bucket = aws_s3_bucket.main.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Optional Intelligent Tiering
resource "aws_s3_bucket_intelligent_tiering_configuration" "main" {
  count = var.bucket_config.intelligent_tiering ? 1 : 0

  bucket = aws_s3_bucket.main.id
  name   = "EntireBucket"

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }

  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 90
  }
}

# Lifecycle configuration with correct resource type
resource "aws_s3_bucket_lifecycle_configuration" "main" {
  count  = length(coalesce(var.bucket_config.lifecycle_rules, [])) > 0 ? 1 : 0
  bucket = aws_s3_bucket.main.id

  dynamic "rule" {
    for_each = var.bucket_config.lifecycle_rules
    content {
      id     = "${rule.value.prefix}-lifecycle-rule"
      status = rule.value.enabled ? "Enabled" : "Disabled"
      
      filter {
        prefix = rule.value.prefix
      }

      dynamic "transition" {
        for_each = rule.value.transition_days != null ? [1] : []
        content {
          days          = rule.value.transition_days
          storage_class = rule.value.storage_class
        }
      }

      dynamic "expiration" {
        for_each = rule.value.expiration_days != null ? [1] : []
        content {
          days = rule.value.expiration_days
        }
      }
    }
  }

  depends_on = [aws_s3_bucket_versioning.main]
}

# Additional S3 bucket configurations based on enhanced bucket_config

# CORS Configuration
resource "aws_s3_bucket_cors_configuration" "main" {
  count = length(coalesce(var.bucket_config.cors_rules, [])) > 0 ? 1 : 0
  
  bucket = aws_s3_bucket.main.id

  dynamic "cors_rule" {
    for_each = var.bucket_config.cors_rules
    content {
      allowed_headers = cors_rule.value.allowed_headers
      allowed_methods = cors_rule.value.allowed_methods
      allowed_origins = cors_rule.value.allowed_origins
      expose_headers  = cors_rule.value.expose_headers
      max_age_seconds = cors_rule.value.max_age_seconds
    }
  }
}


resource "aws_s3_bucket_replication_configuration" "main" {
  count = local.replication_enabled ? 1 : 0

  bucket = aws_s3_bucket.main.id
  role   = aws_iam_role.replication[0].arn

  rule {
    id     = "EntireBucketReplication"
    status = "Enabled"

    destination {
      bucket        = try(var.bucket_config.replication_config.destination_bucket, null)
      storage_class = try(var.bucket_config.replication_config.storage_class, "STANDARD")

      encryption_configuration {
        replica_kms_key_id = var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn
      }
    }

    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
  }

  depends_on = [aws_s3_bucket_versioning.main]
}

# IAM role for replication
resource "aws_iam_role" "replication" {
  count = local.replication_enabled ? 1 : 0

  name = "${local.bucket_name}-replication"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# IAM policy for replication
resource "aws_iam_role_policy" "replication" {
  count = local.replication_enabled ? 1 : 0

  name = "${local.bucket_name}-replication-policy"
  role = aws_iam_role.replication[0].name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetReplicationConfiguration",
          "s3:ListBucket"
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.main.arn
        ]
      },
      {
        Action = [
          "s3:GetObjectVersionForReplication",
          "s3:GetObjectVersionAcl",
          "s3:GetObjectVersionTagging"
        ]
        Effect = "Allow"
        Resource = [
          "${aws_s3_bucket.main.arn}/*"
        ]
      },
      {
        Action = [
          "s3:ReplicateObject",
          "s3:ReplicateDelete",
          "s3:ReplicateTags"
        ]
        Effect = "Allow"
        Resource = [
          "${try(var.bucket_config.replication_config.destination_bucket, "")}/*"
        ]
      },
      {
        Action = [
          "kms:Decrypt"
        ]
        Effect = "Allow"
        Resource = [
          var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn
        ]
      }
    ]
  })
}

# Add public access block for logging bucket
resource "aws_s3_bucket_public_access_block" "logs" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Add logging for the logging bucket (recommended for audit trail)
resource "aws_s3_bucket_logging" "logs" {
  count = var.logging.mode == "create_new" ? 1 : 0

  bucket = aws_s3_bucket.logs[0].id

  target_bucket = aws_s3_bucket.logs[0].id
  target_prefix = "self-logs/"
}

