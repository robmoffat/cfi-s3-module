# Wrapper for lifecycle rules
module "lifecycle_wrapper" {
  source = "../../modules/lifecycle"

  for_each = var.lifecycle_rules

  bucket = each.value.bucket
  prefix = each.value.prefix
  
  transitions = try(each.value.transitions, [])
  expiration = try(each.value.expiration, null)
  
  tags = try(each.value.tags, {})
} 