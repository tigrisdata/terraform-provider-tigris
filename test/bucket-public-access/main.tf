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

resource "tigris_bucket" "bucket" {
  bucket = "${var.test_id}-pub-access-bucket"
}

resource "tigris_bucket_public_access" "public" {
  bucket              = tigris_bucket.bucket.bucket
  acl                 = "public-read"
  public_list_objects = true
}

output "acl" {
  value = tigris_bucket_public_access.public.acl
}

output "public_list_objects" {
  value = tigris_bucket_public_access.public.public_list_objects
}
