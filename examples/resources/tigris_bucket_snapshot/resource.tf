# Create a bucket with snapshots enabled
resource "tigris_bucket" "source" {
  bucket          = "my-source-bucket"
  enable_snapshot = true
}

# Create a snapshot of the bucket
resource "tigris_bucket_snapshot" "example" {
  source_bucket = tigris_bucket.source.bucket
  snapshot_name = "my-snapshot"
}
