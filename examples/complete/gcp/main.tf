terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5"
    }
  }
}

#
# --------------------------
# Variables
# --------------------------
#
variable "project_id" {
  description = "GCP project ID to deploy into"
  type        = string
  # You can override via TF_VAR_project_id in CI
  default     = "woven-precept-353210"
}

variable "location" {
  description = "Region for KMS and GCS (must match). For KMS use lower-case (e.g., us-central1)."
  type        = string
  default     = "us-central1"
}

variable "bucket_name_prefix" {
  description = "Prefix for the secure bucket"
  type        = string
  default     = "cfi-secure-bucket"
}

variable "log_bucket_name_prefix" {
  description = "Prefix for the log bucket"
  type        = string
  default     = "cfi-logs-bucket"
}

#
# --------------------------
# Provider & project data
# --------------------------
#
provider "google" {
  project = var.project_id
  region  = var.location
}

data "google_project" "current" {}

locals {
  project_number = data.google_project.current.number
  # GCS service account that must use the CMEK for default object encryption
  gcs_service_account = "service-${local.project_number}@gs-project-accounts.iam.gserviceaccount.com"
}

#
# --------------------------
# Enable required APIs
# --------------------------
#
resource "google_project_service" "enable_storage" {
  project = var.project_id
  service = "storage.googleapis.com"
}

resource "google_project_service" "enable_kms" {
  project = var.project_id
  service = "cloudkms.googleapis.com"
}

#
# --------------------------
# Random suffix for unique names
# --------------------------
#
resource "random_id" "suffix" {
  byte_length = 4
}

#
# --------------------------
# KMS: key ring & key (CMEK)
# --------------------------
#
resource "google_kms_key_ring" "bucket_keyring" {
  name     = "cfi-bucket-keyring-${random_id.suffix.hex}"
  location = var.location

  depends_on = [
    google_project_service.enable_kms
  ]
}

resource "google_kms_crypto_key" "bucket_key" {
  name            = "cfi-bucket-key-${random_id.suffix.hex}"
  key_ring        = google_kms_key_ring.bucket_keyring.id
  rotation_period = "2592000s" # 30 days

  # Use automatic purpose for GCS (ENCRYPT_DECRYPT)
  depends_on = [
    google_kms_key_ring.bucket_keyring
  ]
}

# Grant Cloud Storage service account permission to use the key
# (Fixes: "Permission denied on Cloud KMS key. Please ensure that your Cloud Storage service account has been authorized to use this key.")
resource "google_kms_crypto_key_iam_member" "gcs_uses_key" {
  crypto_key_id = google_kms_crypto_key.bucket_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${local.gcs_service_account}"
}

#
# --------------------------
# Logging bucket (optional but typical)
# --------------------------
#
resource "google_storage_bucket" "log_bucket" {
  name     = "${var.log_bucket_name_prefix}-${random_id.suffix.hex}"
  location = upper(var.location)   # GCS bucket expects e.g. "US-CENTRAL1". KMS uses "us-central1".
  # No CMEK on logs bucket required, but you can add it if you want consistency.

  uniform_bucket_level_access = true

  depends_on = [
    google_project_service.enable_storage
  ]
}

#
# --------------------------
# Secure bucket using CMEK
# --------------------------
#
resource "google_storage_bucket" "secure_bucket" {
  name     = "${var.bucket_name_prefix}-${random_id.suffix.hex}"
  location = upper(var.location)   # must correspond to the same region as the KMS key
  uniform_bucket_level_access = true

  # Server-side encryption with CMEK
  encryption {
    default_kms_key_name = google_kms_crypto_key.bucket_key.id
  }

  # (Optional) Access logging into the log bucket
  logging {
    log_bucket = google_storage_bucket.log_bucket.name
    # log_object_prefix = "access-logs/"
  }

  # Ensure the GCS SA has been granted access to the key before bucket creation references it
  depends_on = [
    google_project_service.enable_storage,
    google_kms_crypto_key_iam_member.gcs_uses_key
  ]
}

#
# --------------------------
# (Optional) If your CI/SA needs to encrypt/decrypt with the key directly,
# grant it here. Replace the email with your deploy SA if needed.
# --------------------------
#
# resource "google_kms_crypto_key_iam_member" "ci_sa_uses_key" {
#   crypto_key_id = google_kms_crypto_key.bucket_key.id
#   role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
#   member        = "serviceAccount:github-actions-service-account@${var.project_id}.iam.gserviceaccount.com"
# }

#
# --------------------------
# Outputs
# --------------------------
#
output "bucket_name" {
  value = google_storage_bucket.secure_bucket.name
}

output "log_bucket_name" {
  value = google_storage_bucket.log_bucket.name
}

output "kms_key" {
  value = google_kms_crypto_key.bucket_key.id
}

output "kms_keyring" {
  value = google_kms_key_ring.bucket_keyring.id
}
