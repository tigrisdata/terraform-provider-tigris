package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
	"github.com/tigrisdata/terraform-provider-tigris/internal/types"
)

func resourceTigrisBucketWebsiteConfig() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket website configuration resource.",
		CreateWithoutTimeout: resourceBucketWebsiteCreate,
		ReadWithoutTimeout:   resourceBucketWebsiteRead,
		UpdateWithoutTimeout: resourceBucketWebsiteUpdate,
		DeleteWithoutTimeout: resourceBucketWebsiteDelete,

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
			names.AttrDomainName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The custom domain name to apply to the bucket.",
			},
		},
	}
}

func resourceBucketWebsiteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)
	website_domain := d.Get(names.AttrDomainName).(string)

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
		Website: &types.BucketWebsiteConfig{
			DomainName: website_domain,
		},
	}

	tflog.Info(ctx, "Creating bucket website config", map[string]interface{}{
		"bucket_name": bucketName,
	})

	if err := svc.UpdateBucket(ctx, input); err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket website config, %w", err))
	}

	tflog.Info(ctx, "Bucket website config created successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId(bucketName)

	return resourceBucketWebsiteRead(ctx, d, meta)
}

func resourceBucketWebsiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	tflog.Info(ctx, "Fetching bucket metadata", map[string]interface{}{
		"bucket_name": bucketName,
	})

	metadata, err := svc.GetBucketMetadata(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read bucket metadata, %w", err))
	}

	tflog.Info(ctx, "Fetched bucket metadata", map[string]interface{}{
		"bucket_name": bucketName,
	})

	if metadata.Website != nil && metadata.Website.DomainName != "" {
		d.Set(names.AttrDomainName, metadata.Website.DomainName)
	} else {
		tflog.Warn(ctx, "Bucket website configuration not found, removing from state", map[string]interface{}{
			"id": d.Id(),
		})

		d.SetId("")
	}

	return nil
}

func resourceBucketWebsiteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
	}
	needsUpdate := false

	tflog.Info(ctx, "Updating bucket website configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	//
	// Bucket Domain Name.
	//
	if d.HasChange(names.AttrDomainName) {
		input.Website = &types.BucketWebsiteConfig{
			DomainName: d.Get(names.AttrDomainName).(string),
		}

		tflog.Info(ctx, "Will update bucket domain name", map[string]interface{}{
			"bucket_name": bucketName,
		})

		needsUpdate = true
	}

	if needsUpdate {
		err := svc.UpdateBucket(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update bucket website configuration, %w", err))
		}
	}

	tflog.Info(ctx, "Bucket website configuration updated successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	return resourceBucketWebsiteRead(ctx, d, meta)
}

func resourceBucketWebsiteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
		Website: &types.BucketWebsiteConfig{
			DomainName: "",
		},
	}

	tflog.Info(ctx, "Deleting bucket website configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	err := svc.UpdateBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete bucket website configuration, %w", err))
	}

	tflog.Info(ctx, "Bucket website configuration deleted successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId("")
	return nil
}
