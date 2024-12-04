variable "lifecycle_rules" {
  description = "Map of lifecycle rules to apply to buckets"
  type = map(object({
    bucket = string
    prefix = optional(string)
    
    # Transition rules for current versions
    transitions = optional(list(object({
      days          = number
      storage_class = string
      # Validate storage class in module
    })), [])
    
    # Expiration settings for current versions
    expiration = optional(object({
      enabled = bool
      days    = optional(number)
      expired_object_delete_marker = optional(bool, false)
    }))
    
    # Cleanup settings for old versions
    noncurrent_version_expiration = optional(object({
      enabled = bool
      days    = optional(number)
    }))
    
    # Cleanup for incomplete multipart uploads
    abort_incomplete_multipart_upload = optional(object({
      enabled = bool
      days_after_initiation = optional(number, 7)
    }))
    
    tags = optional(map(string), {})
  }))
  
  validation {
    condition = alltrue([
      for rule in values(var.lifecycle_rules) : alltrue([
        for transition in coalesce(rule.transitions, []) :
        contains([
          "STANDARD_IA",
          "ONEZONE_IA",
          "INTELLIGENT_TIERING",
          "GLACIER",
          "DEEP_ARCHIVE"
        ], transition.storage_class)
      ])
    ])
    error_message = "Invalid storage class in transitions. Must be one of: STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER, DEEP_ARCHIVE"
  }

  validation {
    condition = alltrue([
      for rule in values(var.lifecycle_rules) :
      rule.expiration == null || !rule.expiration.enabled || rule.expiration.days != null
    ])
    error_message = "When expiration is enabled, days must be specified"
  }
} 