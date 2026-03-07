# Create a global bucket (default)
resource "tigris_bucket" "global_bucket" {
  bucket = "my-global-bucket"
}

# Create a bucket in a single region
resource "tigris_bucket" "regional_bucket" {
  bucket               = "my-regional-bucket"
  default_storage_tier = "STANDARD_IA"

  location {
    type    = "single"
    regions = ["sjc"]
  }
}

# Create a bucket with multi-region replication
resource "tigris_bucket" "multi_region_bucket" {
  bucket = "my-us-bucket"

  location {
    type    = "multi"
    regions = ["usa"]
  }
}

# Create a bucket with dual-region replication
resource "tigris_bucket" "dual_region_bucket" {
  bucket = "my-eu-bucket"

  location {
    type    = "dual"
    regions = ["fra", "ams"]
  }
}
