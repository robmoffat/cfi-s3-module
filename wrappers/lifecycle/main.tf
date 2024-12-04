# Wrapper for lifecycle rules
module "lifecycle_wrapper" {
  source = "../../modules/lifecycle"

  for_each = var.lifecycle_rules

  bucket = each.value.bucket
  prefix = try(each.value.prefix, null)
  
  # Enhanced transition rules
  transitions = try(each.value.transitions, [])
  
  # Expiration configuration
  expiration = try(each.value.expiration, {
    enabled = false
    days = null
    expired_object_delete_marker = false
  })
  
  # Version cleanup
  noncurrent_version_expiration = try(each.value.noncurrent_version_expiration, {
    enabled = false
    days = null
  })
  
  # Incomplete multipart upload cleanup
  abort_incomplete_multipart_upload = try(each.value.abort_incomplete_multipart_upload, {
    enabled = false
    days_after_initiation = 7
  })

  tags = try(each.value.tags, {})
} 