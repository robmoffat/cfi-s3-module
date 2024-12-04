module "wrapper" {
  source = "../../modules/object"

  for_each = var.items

  bucket                = each.value.bucket
  key                   = each.value.key
  content              = try(each.value.content, null)
  content_type         = try(each.value.content_type, null)
  content_base64       = try(each.value.content_base64, null)
  source               = try(each.value.source, null)
  force_destroy        = try(each.value.force_destroy, false)
  object_lock_mode     = try(each.value.object_lock_mode, null)
  object_lock_retain_until_date = try(each.value.object_lock_retain_until_date, null)
  tags                 = try(each.value.tags, {})
} 
