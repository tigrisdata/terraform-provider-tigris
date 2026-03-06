terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {}

# Global bucket (default behavior, no location block needed)
resource "tigris_bucket" "global" {
  bucket = "test-global-bucket"
}

# Single-region bucket
resource "tigris_bucket" "single_region" {
  bucket = "test-single-region-bucket"

  location {
    type    = "single"
    regions = ["sjc"]
  }
}

# Multi-region bucket
resource "tigris_bucket" "multi_region" {
  bucket = "test-multi-region-bucket"

  location {
    type    = "multi"
    regions = ["usa"]
  }
}

# Dual-region bucket
resource "tigris_bucket" "dual_region" {
  bucket = "test-dual-region-bucket"

  location {
    type    = "dual"
    regions = ["fra", "ams"]
  }
}

output "global_bucket" {
  value = tigris_bucket.global.bucket
}

output "single_region_location" {
  value = tigris_bucket.single_region.location
}

output "multi_region_location" {
  value = tigris_bucket.multi_region.location
}

output "dual_region_location" {
  value = tigris_bucket.dual_region.location
}
