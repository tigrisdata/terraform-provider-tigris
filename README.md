# Terraform Provider for Tigris Buckets

This repository contains a Terraform provider that allows you to create and manage Tigris buckets. This provider supports creating, reading, updating, deleting, and importing Tigris buckets with additional customization options like specifying access keys.

## Usage

Below is an example of how to use the Tigris provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    tigris = {
      source  = "https://github.com/tigrisdata/terraform-provider-tigris"
    }
  }
}

provider "tigris" {
  access_key = "your-access-key"
  secret_key = "your-secret-key"
}

resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}

resource "tigris_bucket_public_access" "example_bucket_public_access" {
  bucket              = tigris_bucket.example_bucket.bucket
  acl                 = "private"
  public_list_objects = false
}

resource "tigris_bucket_website_config" "example_website_config" {
  bucket      = tigris_bucket.example_bucket.bucket
  domain_name = tigris_bucket.example_bucket.bucket
}

resource "tigris_bucket_shadow_config" "example_shadow_config" {
  bucket                = tigris_bucket.example_bucket.bucket
  shadow_bucket         = "my-custom-bucket-shadow"
  shadow_access_key     = "your-shadow-bucket-access-key"
  shadow_secret_key     = "your-shadow-bucket-secret-key"
  shadow_region         = "us-west-2"
  shadown_endpoint      = "https://s3.us-west-2.amazonaws.com"
  shadow_write_through  = true
}
```

### Applying the Configuration

1. Initialize Terraform:

```shell
terraform init
```

2. Apply the configuration:

```shell
terraform apply
```

## Provider

The Tigris provider allows you to manage Tigris buckets.

### Configuration

The provider can be configured with the following parameters:

- access_key: (Optional) The access key. Can also be sourced from the AWS_ACCESS_KEY_ID environment variable.
- secret_key: (Optional) The secret key. Can also be sourced from the AWS_SECRET_ACCESS_KEY environment variable.

## Resources

### tigris_bucket

The tigris_bucket resource creates and manages a Tigris bucket. This resource supports the following actions:

- Create: Creates a new Tigris bucket.
- Read: Retrieves information about the existing Tigris bucket.
- Update: Updates the bucket configuration.
- Delete: Deletes the Tigris bucket.
- Import: Imports an existing Tigris bucket into Terraform’s state.

#### Configuration

- bucket: (Required) The name of the Tigris bucket.

```hcl
resource "tigris_bucket" "example_bucket" {
  bucket = "my-custom-bucket"
}
```

### tigris_bucket_public_access

The tigris_bucket_public_access resource creates and manages a Tigris bucket public access configuration. This resource supports the following actions:

- Create: Creates a new Tigris bucket public access configuration.
- Read: Retrieves information about the existing Tigris bucket public access configuration.
- Update: Updates the bucket public access configuration.
- Delete: Deletes the Tigris bucket public access configuration.
- Import: Imports an existing Tigris bucket public access configuration into Terraform’s state.

#### Configuration

- bucket: (Required) The name of the Tigris bucket.
- acl: (Optional) The access control list for the bucket. Defaults to "private". Possible values are "private", and "public-read".

```hcl
resource "tigris_bucket_public_access" "example_bucket_public_access" {
  bucket              = "my-custom-bucket"
  acl                 = "private"
  public_list_objects = false
}
```

### tigris_bucket_website_config

The tigris_bucket_website_config resource creates and manages a Tigris bucket website configuration. This is used to configure custom domain name for the bucket.

This resource supports the following actions:

- Create: Creates a new Tigris bucket website configuration.
- Read: Retrieves information about the existing Tigris bucket website configuration.
- Update: Updates the bucket website configuration.
- Delete: Deletes the Tigris bucket website configuration.
- Import: Imports an existing Tigris bucket website configuration into Terraform’s state.

#### Configuration

- bucket: (Required) The name of the Tigris bucket.
- domain_name: (Required) The domain name for the bucket website.

```hcl
resource "tigris_bucket_website_config" "example_website_config" {
  bucket      = images.example.com
  domain_name = images.example.com
}
```

Before using this resource, you must have a bucket created using the tigris_bucket resource. The domain_name must match the bucket name and there must be a CNAME DNS record setup. The CNAME record should point to the Tigris bucket endpoint (e.g., images.example.com CNAME images.example.com.fly.storage.tigris.dev).

### tigris_bucket_shadow_config

The tigris_bucket_shadow_config resource creates and manages a Tigris bucket shadow configuration. The shadow configuration is used to setup a source bucket (shadow bucket) that will be used to migrate data to the Tigris bucket. You can read more about how this migration works [here](https://www.tigrisdata.com/docs/migration/).

This resource supports the following actions:

- Create: Creates a new Tigris bucket shadow configuration.
- Read: Retrieves information about the existing Tigris bucket shadow configuration.
- Update: Updates the bucket shadow configuration.
- Delete: Deletes the Tigris bucket shadow configuration.
- Import: Imports an existing Tigris bucket shadow configuration into Terraform’s state.

#### Configuration

- bucket: (Required) The name of the Tigris bucket.
- shadow_bucket: (Required) The name of the shadow bucket.
- shadow_access_key: (Required) The access key for the shadow bucket.
- shadow_secret_key: (Required) The secret key for the shadow bucket.
- shadow_region: (Optional) The region for the shadow bucket. Defaults to "us-east-1".
- shadow_endpoint: (Optional) The endpoint for the shadow bucket. Defaults to "https://s3.us-east-1.amazonaws.com".
- shadow_write_through: (Optional) Whether to write through to the shadow bucket. Defaults to true.

```hcl
resource "tigris_bucket_shadow_config" "example_shadow_config" {
  bucket                = "my-custom-bucket"
  shadow_bucket         = "my-custom-bucket-shadow"
  shadow_access_key     = "your-shadow-bucket-access-key"
  shadow_secret_key     = "your-shadow-bucket-secret-key"
  shadow_region         = "us-west-2"
  shadown_endpoint      = "https://s3.us-west-2.amazonaws.com"
  shadow_write_through  = true
}
```

## Developing

### Documentation

The documentation for this provider is generated using terraform-plugin-docs. To generate the documentation, run the following command:

```shell
make docs
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
