# Create the bucket
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

# Create the bucket shadow config for data migration from S3-compatible storage to Tigris
resource "tigris_bucket_shadow_config" "example_shadow_config" {
  bucket                = tigris_bucket.example_bucket.bucket
  shadow_bucket         = "my-custom-bucket-shadow"
  shadow_access_key     = "your-shadow-bucket-access-key"
  shadow_secret_key     = "your-shadow-bucket-secret-key"
  shadow_region         = "us-west-2"
  shadown_endpoint      = "https://s3.us-west-2.amazonaws.com"
  shadow_write_through  = true
}