terraform {
  required_providers {
    tigris = {
      source = "tigrisdata/tigris"
    }
  }
}

provider "tigris" {
  endpoint = "https://t3.storage.dev"
}

variable "test_id" {
  type    = string
  default = "t"
}

resource "tigris_bucket" "logs" {
  bucket = "${var.test_id}-lifecycle-bucket"
}

resource "tigris_bucket_lifecycle" "logs" {
  bucket = tigris_bucket.logs.bucket

  rule {
    id     = "logs-to-infrequent-access"
    prefix = "logs/"
    status = "Enabled"

    transition {
      days         = 30
      storage_tier = "STANDARD_IA"
    }
  }

  rule {
    id     = "logs-to-archive-then-expire"
    prefix = "logs/"
    status = "Enabled"

    transition {
      days         = 90
      storage_tier = "GLACIER_IR"
    }

    expiration {
      days = 365
    }
  }

  rule {
    id     = "whole-bucket-expiry"
    status = "Enabled"

    expiration {
      days = 400
    }
  }

  rule {
    id     = "tmp-archive-immediately"
    prefix = "tmp/"
    status = "Enabled"

    transition {
      days         = 0
      storage_tier = "GLACIER"
    }
  }
}

output "bucket" {
  value = tigris_bucket_lifecycle.logs.bucket
}

output "rule_count" {
  value = length(tigris_bucket_lifecycle.logs.rule)
}
