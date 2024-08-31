# Create the bucket
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

# Create the configuration for bucket custom domain
resource "tigris_bucket_website_config" "example_website_config" {
  bucket      = tigris_bucket.example_bucket.bucket
  domain_name = tigris_bucket.example_bucket.bucket
}