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

# Source bucket to fork from (must have snapshots enabled)
resource "tigris_bucket" "source" {
  bucket          = "${var.test_id}-fork-src-bucket"
  enable_snapshot = true
}

# Fork the source bucket (direct fork, no snapshot)
resource "tigris_bucket_fork" "fork" {
  bucket             = "${var.test_id}-forked-bucket"
  fork_source_bucket = tigris_bucket.source.bucket
}

output "fork_bucket" {
  value = tigris_bucket_fork.fork.bucket
}

output "fork_created_at" {
  value = tigris_bucket_fork.fork.fork_created_at
}
