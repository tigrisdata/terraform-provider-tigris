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

func resourceTigrisBucketShadowConfig() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket shadow configuration resource.",
		CreateWithoutTimeout: resourceBucketShadowCreate,
		ReadWithoutTimeout:   resourceBucketShadowRead,
		UpdateWithoutTimeout: resourceBucketShadowUpdate,
		DeleteWithoutTimeout: resourceBucketShadowDelete,

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
			names.AttrShadowBucket: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the shadow bucket.",
			},
			names.AttrShadowAccessKey: {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The access key for the shadow bucket.",
			},
			names.AttrShadowSecretKey: {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The secret key for the shadow bucket.",
			},
			names.AttrShadowRegion: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us-east-1",
				Description: "The region for the shadow bucket.",
			},
			names.AttrShadowEndpoint: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://s3.us-east-1.amazonaws.com",
				Description: "The endpoint for the shadow bucket.",
			},
			names.AttrShadowWriteThrough: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to write through to the shadow bucket.",
			},
		},
	}
}

func resourceBucketShadowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)
	shadowConfig := &types.BucketShadowConfig{
		Name:         d.Get(names.AttrShadowBucket).(string),
		AccessKey:    d.Get(names.AttrShadowAccessKey).(string),
		SecretKey:    d.Get(names.AttrShadowSecretKey).(string),
		Region:       d.Get(names.AttrShadowRegion).(string),
		Endpoint:     d.Get(names.AttrShadowEndpoint).(string),
		WriteThrough: d.Get(names.AttrShadowWriteThrough).(bool),
	}

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
		Shadow: shadowConfig,
	}

	tflog.Info(ctx, "Creating bucket shadow config", map[string]interface{}{
		"bucket_name": bucketName,
	})

	if err := svc.UpdateBucket(ctx, input); err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket shadow config, %w", err))
	}

	tflog.Info(ctx, "Bucket shadow config created successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId(bucketName)

	return resourceBucketShadowRead(ctx, d, meta)
}

func resourceBucketShadowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if metadata.Shadow != nil && metadata.Shadow.Name != "" {
		d.Set(names.AttrShadowBucket, metadata.Shadow.Name)
		d.Set(names.AttrShadowAccessKey, metadata.Shadow.AccessKey)
		d.Set(names.AttrShadowSecretKey, metadata.Shadow.SecretKey)
		d.Set(names.AttrShadowRegion, metadata.Shadow.Region)
		d.Set(names.AttrShadowEndpoint, metadata.Shadow.Endpoint)
		d.Set(names.AttrShadowWriteThrough, metadata.Shadow.WriteThrough)
	} else {
		tflog.Warn(ctx, "Bucket shadow configuration not found, removing from state", map[string]interface{}{
			"id": d.Id(),
		})

		d.SetId("")
	}

	return nil
}

func resourceBucketShadowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
	}
	needsUpdate := false

	tflog.Info(ctx, "Updating bucket shadow configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	//
	// Bucket Shadow Config.
	//
	if d.HasChangesExcept(names.AttrBucket) {
		input.Shadow = &types.BucketShadowConfig{
			Name:         d.Get(names.AttrShadowBucket).(string),
			AccessKey:    d.Get(names.AttrShadowAccessKey).(string),
			SecretKey:    d.Get(names.AttrShadowSecretKey).(string),
			Region:       d.Get(names.AttrShadowRegion).(string),
			Endpoint:     d.Get(names.AttrShadowEndpoint).(string),
			WriteThrough: d.Get(names.AttrShadowWriteThrough).(bool),
		}

		tflog.Info(ctx, "Will update bucket shadow config", map[string]interface{}{
			"bucket_name": bucketName,
		})

		needsUpdate = true
	}

	if needsUpdate {
		err := svc.UpdateBucket(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update bucket shadow configuration, %w", err))
		}
	}

	tflog.Info(ctx, "Bucket shadow configuration updated successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	return resourceBucketWebsiteRead(ctx, d, meta)
}

func resourceBucketShadowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
		Shadow: &types.BucketShadowConfig{},
	}

	tflog.Info(ctx, "Deleting bucket shadow configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	err := svc.UpdateBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete bucket shadow configuration, %w", err))
	}

	tflog.Info(ctx, "Bucket shadow configuration deleted successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId("")
	return nil
}
