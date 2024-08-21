package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
	"github.com/tigrisdata/terraform-provider-tigris/internal/types"
)

func resourceTigrisBucket() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket resource. This can be used to create and manage Tigris buckets.",
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Tigris bucket.",
			},
		},
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)
	if err := validBucketName(bucketName); err != nil {
		return diag.FromErr(fmt.Errorf("invalid bucket name, %w", err))
	}

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
	}

	tflog.Info(ctx, "Creating bucket", map[string]interface{}{
		"bucket_name": bucketName,
	})

	err := svc.CreateBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket, %w", err))
	}

	tflog.Info(ctx, "Bucket created successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId(bucketName)

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	tflog.Info(ctx, "Checking bucket existence", map[string]interface{}{
		"bucket_name": bucketName,
	})

	exists, err := svc.HeadBucket(ctx, bucketName)
	if !exists {
		tflog.Warn(ctx, "Bucket not found, removing from state", map[string]interface{}{
			"bucket_name": bucketName,
		})

		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read bucket, %w", err))
	}

	d.Set(names.AttrBucket, bucketName)

	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This resource cannot be updated
	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	err := svc.DeleteBucket(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete bucket, %w", err))
	}

	d.SetId("")
	return nil
}

// validBucketName validates bucket name. Buckets names have to be DNS-compliant.
func validBucketName(value string) error {
	if (len(value) < 3) || (len(value) > 63) {
		return fmt.Errorf("%q must contain from 3 to 63 characters", value)
	}
	if !regexache.MustCompile(`^[0-9a-z-.]+$`).MatchString(value) {
		return fmt.Errorf("only lowercase alphanumeric characters and hyphens allowed in %q", value)
	}
	if regexache.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`).MatchString(value) {
		return fmt.Errorf("%q must not be formatted as an IP address", value)
	}
	if strings.HasPrefix(value, `.`) {
		return fmt.Errorf("%q cannot start with a period", value)
	}
	if strings.HasSuffix(value, `.`) {
		return fmt.Errorf("%q cannot end with a period", value)
	}
	if strings.Contains(value, `..`) {
		return fmt.Errorf("%q can be only one period between labels", value)
	}

	return nil
}
