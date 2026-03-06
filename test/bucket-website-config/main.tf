terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {}

resource "tigris_bucket" "bucket" {
  bucket = "test-website-config-bucket"
}

resource "tigris_bucket_website_config" "website" {
  bucket      = tigris_bucket.bucket.bucket
  domain_name = "test-website-config-bucket"
}

output "domain_name" {
  value = tigris_bucket_website_config.website.domain_name
}
