# Configure the Tigris provider using the required_providers stanza
# required with Terraform 0.13 and beyond. You may optionally use version
# directive to prevent breaking changes occurring unannounced.
terraform {
  required_providers {
    tigris = {
      source  = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {
  access_key = "your-access-key"
  secret_key = "your-secret-key"
}

# Create a bucket
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

# Create bucket public access rule
resource "tigris_bucket_public_access" "example_bucket_public_access" {
  bucket              = tigris_bucket.example_bucket.bucket
  acl                 = "private"
  public_list_objects = false
}