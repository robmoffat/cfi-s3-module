variable "project_id" {
  description = "GCP project ID to deploy into"
  type        = string
}

variable "region" {
  description = "GCP region to deploy into"
  type        = string
  default     = "us-central1" # optional fallback
}

variable "location" {
  description = "GCP location for multi-region storage"
  type        = string
  default     = "US" # optional fallback
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "random_id" "suffix" {
  byte_length = 4
}

data "google_project" "current" {}

# Create a secure GCS bucket
resource "google_storage_bucket" "secure_bucket" {
  name          = "cfi-secure-bucket-${random_id.suffix.hex}"
  location      = var.location
  force_destroy = false

  # Enable versioning
  versioning {
    enabled = true
  }

  # Enable lifecycle management
  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }

  # Soft delete policy
  lifecycle_rule {
    condition {
      age = 7
    }
    action {
      type = "Delete"
    }
  }

  # Uniform bucket-level access
  uniform_bucket_level_access = true

  # Public access prevention
  public_access_prevention = "enforced"

  # Encryption configuration
  encryption {
    default_kms_key_name = google_kms_crypto_key.bucket_key.id
  }

  # Logging configuration
  logging {
    log_bucket        = google_storage_bucket.log_bucket.name
    log_object_prefix = "logs"
  }

  # Retention policy
  retention_policy {
    retention_period = 7776000 # 90 days in seconds
  }

  labels = {
    environment = "production"
    owner       = "cfi"
  }
}

# Create a log bucket for access logs
resource "google_storage_bucket" "log_bucket" {
  name          = "cfi-logs-bucket-${random_id.suffix.hex}"
  location      = var.location
  force_destroy = false

  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }

  labels = {
    environment = "production"
    owner       = "cfi"
    purpose     = "logs"
  }
}

# Create a KMS key for bucket encryption
resource "google_kms_key_ring" "bucket_keyring" {
  name     = "cfi-bucket-keyring-${random_id.suffix.hex}"
  location = var.region
}

resource "google_kms_crypto_key" "bucket_key" {
  name     = "cfi-bucket-key-${random_id.suffix.hex}"
  key_ring = google_kms_key_ring.bucket_keyring.id

  lifecycle {
    prevent_destroy = false
  }

  labels = {
    environment = "production"
    owner       = "cfi"
  }
}

# IAM binding for the KMS key
resource "google_kms_crypto_key_iam_binding" "crypto_key" {
  crypto_key_id = google_kms_crypto_key.bucket_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  members = [
    "serviceAccount:${data.google_project.current.number}-compute@developer.gserviceaccount.com",
    "user:terraform-user@${var.project_id}.iam.gserviceaccount.com"
  ]
}

# Create a sample object in the bucket
resource "google_storage_bucket_object" "sample_file" {
  name   = "sample/hello-world.txt"
  bucket = google_storage_bucket.secure_bucket.name
  content = "Hello, World! This is a sample file in the secure CFI bucket."

  depends_on = [google_storage_bucket.secure_bucket]
}

# Outputs
output "bucket_name" {
  description = "The name of the secure GCS bucket"
  value       = google_storage_bucket.secure_bucket.name
}

output "bucket_url" {
  description = "The URL of the secure GCS bucket"
  value       = google_storage_bucket.secure_bucket.url
}

output "kms_key_name" {
  description = "The name of the KMS key used for encryption"
  value       = google_kms_crypto_key.bucket_key.name
}

output "log_bucket_name" {
  description = "The name of the log bucket"
  value       = google_storage_bucket.log_bucket.name
}
