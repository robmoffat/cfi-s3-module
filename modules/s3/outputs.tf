# outputs.tf

output "bucket_id" {
  description = "The name of the bucket"
  value       = aws_s3_bucket.main.id
}

output "bucket_arn" {
  description = "The ARN of the bucket"
  value       = aws_s3_bucket.main.arn
}

output "kms_key_arn" {
  description = "The ARN of the KMS key"
  value       = var.kms_key.create ? aws_kms_key.this[0].arn : var.kms_key.key_arn
}

output "kms_key_id" {
  description = "The ID of the KMS key"
  value       = var.kms_key.create ? aws_kms_key.this[0].id : null
}

output "log_bucket_id" {
  description = "The name of the log bucket"
  value       = var.logging.mode == "create_new" ? aws_s3_bucket.logs[0].id : null
}

output "log_bucket_arn" {
  description = "The ARN of the log bucket"
  value       = var.logging.mode == "create_new" ? aws_s3_bucket.logs[0].arn : null
}

output "cloudwatch_log_group_arn" {
  description = "The ARN of the CloudWatch log group"
  value       = var.logging.mode != "disabled" ? aws_cloudwatch_log_group.this[0].arn : null
}
