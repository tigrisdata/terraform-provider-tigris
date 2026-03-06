terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {}

# Basic bucket with defaults
resource "tigris_bucket" "basic" {
  bucket = "test-basic-bucket"
}

# Bucket with a non-default storage tier
resource "tigris_bucket" "standard_ia" {
  bucket               = "test-standard-ia-bucket"
  default_storage_tier = "STANDARD_IA"
}

output "basic_bucket" {
  value = tigris_bucket.basic.bucket
}

output "standard_ia_bucket" {
  value = tigris_bucket.standard_ia.bucket
}

output "standard_ia_storage_tier" {
  value = tigris_bucket.standard_ia.default_storage_tier
}
