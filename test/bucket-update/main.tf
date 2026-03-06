terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {}

# Bucket whose location can be updated in-place.
#
# To test the update path:
#   1. Apply with the initial location (single region, sjc)
#   2. Change the location block below (e.g. to multi-region "usa")
#   3. Apply again and verify the update succeeds
resource "tigris_bucket" "updatable" {
  bucket = "test-updatable-bucket"

  location {
    type    = "single"
    regions = ["sjc"]
  }
}

output "bucket" {
  value = tigris_bucket.updatable.bucket
}

output "location" {
  value = tigris_bucket.updatable.location
}
