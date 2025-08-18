# GCP Secure Storage Example

This example demonstrates how to create a secure Google Cloud Storage bucket using the CFI S3 module patterns adapted for GCP.

## Features

- **Secure GCS Bucket**: Creates a Google Cloud Storage bucket with encryption enabled
- **KMS Encryption**: Uses Google Cloud KMS for customer-managed encryption keys
- **Versioning**: Enables object versioning for data protection
- **Lifecycle Management**: Implements automatic deletion policies (90 days for data, 7 days for soft delete)
- **Access Logging**: Creates a separate log bucket for access monitoring
- **Public Access Prevention**: Enforces private access only
- **Uniform Bucket-Level Access**: Enables consistent IAM permissions
- **Retention Policy**: Implements 90-day retention for compliance

## Prerequisites

- Google Cloud SDK installed and configured
- Terraform >= 1.0
- Appropriate GCP permissions to create:
  - Storage buckets
  - KMS key rings and crypto keys
  - IAM bindings

## Usage

1. **Set your GCP project ID**:
   ```bash
   export TF_VAR_project_id="your-gcp-project-id"
   ```

2. **Initialize Terraform**:
   ```bash
   terraform init
   ```

3. **Review the plan**:
   ```bash
   terraform plan
   ```

4. **Apply the configuration**:
   ```bash
   terraform apply
   ```

## Variables

| Variable | Description | Type | Default | Required |
|----------|-------------|------|---------|----------|
| `project_id` | GCP project ID to deploy into | `string` | - | Yes |
| `region` | GCP region for KMS resources | `string` | `"us-central1"` | No |
| `location` | GCP location for multi-region storage | `string` | `"US"` | No |

## Outputs

- `bucket_name`: The name of the secure GCS bucket
- `bucket_url`: The URL of the secure GCS bucket
- `kms_key_name`: The name of the KMS key used for encryption
- `log_bucket_name`: The name of the log bucket

## Security Features

This configuration implements several security best practices:

- **Encryption at Rest**: All data is encrypted using customer-managed KMS keys
- **Access Control**: Uniform bucket-level access with public access prevention
- **Audit Logging**: Comprehensive access logging to a separate bucket
- **Versioning**: Object versioning for data recovery and compliance
- **Lifecycle Management**: Automated cleanup of old data and versions
- **Retention Policies**: Configurable data retention for compliance requirements

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Note**: The KMS key and buckets will be permanently deleted. Ensure you have backups of any important data.
