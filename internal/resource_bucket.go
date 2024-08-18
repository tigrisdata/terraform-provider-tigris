package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTigrisBucket() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Tigris bucket resource. This can be used to create and manage Tigris buckets.",
		Create:      resourceS3BucketCreate,
		Read:        resourceS3BucketRead,
		Update:      resourceS3BucketUpdate,
		Delete:      resourceS3BucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Tigris bucket.",
			},
		},
	}
}

func resourceS3BucketCreate(d *schema.ResourceData, meta interface{}) error {
	svc := meta.(*s3.Client)

	bucketName := d.Get("bucket_name").(string)

	_, err := svc.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("unable to create bucket, %w", err)
	}

	d.SetId(bucketName)
	return resourceS3BucketRead(d, meta)
}

func resourceS3BucketRead(d *schema.ResourceData, meta interface{}) error {
	svc := meta.(*s3.Client)

	bucketName := d.Id()

	_, err := svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var notFoundErr *types.NotFound
		if ok := errors.As(err, &notFoundErr); ok {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("unable to read bucket, %w", err)
	}

	return nil
}

func resourceS3BucketUpdate(d *schema.ResourceData, meta interface{}) error {
	// Since S3 buckets have limited update capabilities, this might be a no-op
	return resourceS3BucketRead(d, meta)
}

func resourceS3BucketDelete(d *schema.ResourceData, meta interface{}) error {
	svc := meta.(*s3.Client)

	bucketName := d.Id()

	_, err := svc.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("unable to delete bucket, %w", err)
	}

	d.SetId("")
	return nil
}
