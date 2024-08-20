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

resource "tigris_bucket" "example" {
  bucket = "my-custom-bucket"
  acl    = "private"
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

## Configuration

### Provider Configuration

The provider can be configured with the following parameters:

- access_key: (Optional) The access key. Can also be sourced from the AWS_ACCESS_KEY_ID environment variable.
- secret_key: (Optional) The secret key. Can also be sourced from the AWS_SECRET_ACCESS_KEY environment variable.

### Resource Configuration

- bucket_name: (Required) The name of the Tigris bucket.
- acl: (Optional) The access control list for the bucket. Defaults to "private". Possible values are "private", and "public-read".

## Resources

### tigris_bucket

The tigris_bucket resource creates and manages a Tigris bucket. This resource supports the following actions:

- Create: Creates a new Tigris bucket.
- Read: Retrieves information about the existing Tigris bucket.
- Update: Updates the bucket configuration.
- Delete: Deletes the Tigris bucket.
- Import: Imports an existing Tigris bucket into Terraform’s state.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
