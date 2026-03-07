terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {}

variable "test_id" {
  type    = string
  default = "t"
}

# Source bucket with snapshots enabled
resource "tigris_bucket" "source" {
  bucket          = "${var.test_id}-snapshot-src-bucket"
  enable_snapshot = true
}

# Create a snapshot of the source bucket
resource "tigris_bucket_snapshot" "snap" {
  source_bucket = tigris_bucket.source.bucket
  snapshot_name = "${var.test_id}-snapshot"
}

# Fork from the snapshot
resource "tigris_bucket_fork" "from_snapshot" {
  bucket                       = "${var.test_id}-fork-from-snap-bucket"
  fork_source_bucket           = tigris_bucket.source.bucket
  fork_source_bucket_snapshot  = tigris_bucket_snapshot.snap.snapshot_version
}

output "snapshot_version" {
  value = tigris_bucket_snapshot.snap.snapshot_version
}

output "snapshot_created_at" {
  value = tigris_bucket_snapshot.snap.snapshot_created_at
}

output "fork_from_snapshot_created_at" {
  value = tigris_bucket_fork.from_snapshot.fork_created_at
}
