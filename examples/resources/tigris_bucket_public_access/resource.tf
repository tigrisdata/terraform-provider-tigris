# Create the bucket
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

# Create bucket public access rule
resource "tigris_bucket_public_access" "example_bucket_public_access" {
  bucket              = tigris_bucket.example_bucket.bucket
  acl                 = "private"
  public_list_objects = false
}