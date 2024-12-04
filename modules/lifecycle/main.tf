resource "aws_s3_bucket_lifecycle_configuration" "this" {
  bucket = var.bucket

  rule {
    id     = var.prefix != null ? "${var.prefix}-lifecycle" : "bucket-lifecycle"
    status = "Enabled"

    # Filter based on prefix if provided
    dynamic "filter" {
      for_each = var.prefix != null ? [1] : []
      content {
        prefix = var.prefix
      }
    }

    # Add transitions
    dynamic "transition" {
      for_each = var.transitions
      content {
        days          = transition.value.days
        storage_class = transition.value.storage_class
      }
    }

    # Add expiration if enabled
    dynamic "expiration" {
      for_each = var.expiration.enabled ? [1] : []
      content {
        days = var.expiration.days
        expired_object_delete_marker = var.expiration.expired_object_delete_marker
      }
    }

    # Add noncurrent version expiration if enabled
    dynamic "noncurrent_version_expiration" {
      for_each = var.noncurrent_version_expiration.enabled ? [1] : []
      content {
        noncurrent_days = var.noncurrent_version_expiration.days
      }
    }

    # Add incomplete multipart upload cleanup if enabled
    dynamic "abort_incomplete_multipart_upload" {
      for_each = var.abort_incomplete_multipart_upload.enabled ? [1] : []
      content {
        days_after_initiation = var.abort_incomplete_multipart_upload.days_after_initiation
      }
    }
  }
} 