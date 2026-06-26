# Create the bucket
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

# Age and expire objects under the logs/ prefix.
# Tigris allows a single transition per rule, so multi-tier aging is expressed
# as one rule per tier.
resource "tigris_bucket_lifecycle" "example_bucket_lifecycle" {
  bucket = tigris_bucket.example_bucket.bucket

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
}
