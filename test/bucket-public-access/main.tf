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

resource "tigris_bucket" "public" {
  bucket = "${var.test_id}-pub-access-bucket"
}

resource "tigris_bucket_public_access" "public" {
  bucket              = tigris_bucket.public.bucket
  acl                 = "public-read"
  public_list_objects = true
}

resource "tigris_bucket" "private" {
  bucket = "${var.test_id}-priv-access-bucket"
}

resource "tigris_bucket_public_access" "private" {
  bucket              = tigris_bucket.private.bucket
  acl                 = "private"
  public_list_objects = false
}

output "acl" {
  value = tigris_bucket_public_access.public.acl
}

output "public_list_objects" {
  value = tigris_bucket_public_access.public.public_list_objects
}

output "private_acl" {
  value = tigris_bucket_public_access.private.acl
}

output "private_public_list_objects" {
  value = tigris_bucket_public_access.private.public_list_objects
}
