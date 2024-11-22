# Secure S3 Bucket Terraform Module

This module creates a secure S3 bucket with comprehensive security controls and optional supporting resources.

## Features

- Configurable KMS encryption
- Optional logging bucket creation
- CloudWatch logs integration
- Intelligent tiering support
- Flexible lifecycle rules
- Comprehensive security controls implementing CCC standards

## Usage

Basic usage:

```hcl
module "secure_s3" {
  source = "git::https://github.com/org/terraform-aws-secure-s3.git?ref=v1.0.0"
  
  prefix      = "prod"
  bucket_name = "my-application-data"
  
  tags = {
    Environment = "Production"
    Owner       = "Platform Team"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0.0 |
| aws | >= 4.0.0 |

## Providers

| Name | Version |
|------|---------|
| aws | >= 4.0.0 |

## Resources Created

This module creates the following resources:

- S3 Bucket with security configurations
- Optional KMS Key for encryption
- Optional S3 Bucket for access logging
- Optional CloudWatch Log Group
- Associated IAM policies and bucket policies

## Input Variables

See [variables.tf](./variables.tf) for detailed descriptions of all input variables.

## Outputs

See [outputs.tf](./outputs.tf) for detailed descriptions of all outputs.

## Security Controls

This module implements the following security controls:
- CCC.C01: Prevent unencrypted requests
- CCC.C02: Ensure data encryption at rest
- [List continues...]

## Contributing

Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 Licensed. See [LICENSE](./LICENSE) for full details.
# cfi-s3-module
