variable "items" {
  description = "Map of S3 objects to create"
  type = map(object({
    bucket                = string
    key                   = string
    content              = optional(string)
    content_type         = optional(string)
    content_base64       = optional(string)
    source               = optional(string)
    force_destroy        = optional(bool, false)
    object_lock_mode     = optional(string)
    object_lock_retain_until_date = optional(string)
    tags                 = optional(map(string), {})
  }))
} 