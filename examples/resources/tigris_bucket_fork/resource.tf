# Source bucket with snapshots enabled
resource "tigris_bucket" "source" {
  bucket          = "my-source-bucket"
  enable_snapshot = true
}

# Create a snapshot to fork from
resource "tigris_bucket_snapshot" "example" {
  source_bucket = tigris_bucket.source.bucket
  snapshot_name = "pre-fork-snapshot"
}

# Create a fork of an existing bucket
resource "tigris_bucket_fork" "example" {
  bucket             = "my-forked-bucket"
  fork_source_bucket = tigris_bucket.source.bucket
}

# Create a fork from a specific snapshot
resource "tigris_bucket_fork" "from_snapshot" {
  bucket                      = "my-forked-from-snapshot"
  fork_source_bucket          = tigris_bucket.source.bucket
  fork_source_bucket_snapshot = tigris_bucket_snapshot.example.snapshot_version
}
