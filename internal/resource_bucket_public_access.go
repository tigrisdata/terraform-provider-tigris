package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
	"github.com/tigrisdata/terraform-provider-tigris/internal/types"
)

func resourceTigrisBucketPublicAccess() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket resource. This can be used to create and manage Tigris buckets.",
		CreateWithoutTimeout: resourceBucketPublicAccessCreate,
		ReadWithoutTimeout:   resourceBucketPublicAccessRead,
		UpdateWithoutTimeout: resourceBucketPublicAccessUpdate,
		DeleteWithoutTimeout: resourceBucketPublicAccessDelete,

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
			names.AttrAcl: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      string(types.BucketCannedACLPrivate),
				Description:  "The canned ACL to apply to the bucket.",
				ValidateFunc: validation.StringInSlice(bucketCannedACL_Values(), false),
			},
			names.AttrPublicListObjects: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to allow public listing of objects in the bucket.",
			},
		},
	}
}

func resourceBucketPublicAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)
	publicListObjects := d.Get(names.AttrPublicListObjects).(bool)

	input := &types.BucketUpdateInput{
		Bucket:                   bucketName,
		PublicObjectsListEnabled: &publicListObjects,
	}

	acl := types.BucketCannedACLPrivate
	if v, ok := d.GetOk(names.AttrAcl); ok {
		acl = types.BucketCannedACL(v.(string))
	}
	input.ACL = &acl

	tflog.Info(ctx, "Creating bucket public access config", map[string]interface{}{
		"bucket_name": bucketName,
	})

	if err := svc.UpdateBucket(ctx, input); err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket public access config, %w", err))
	}

	tflog.Info(ctx, "Bucket public access config created successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId(bucketName)

	return resourceBucketPublicAccessRead(ctx, d, meta)
}

func resourceBucketPublicAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	acl := metadata.GetBucketCannedACL()
	d.Set(names.AttrAcl, string(acl))

	d.Set(names.AttrPublicListObjects, metadata.GetPublicObjectsListEnabled())

	return nil
}

func resourceBucketPublicAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	input := &types.BucketUpdateInput{
		Bucket: bucketName,
	}
	needsUpdate := false

	tflog.Info(ctx, "Updating bucket public access configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	//
	// Bucket ACL.
	//
	if d.HasChange(names.AttrAcl) {
		acl := types.BucketCannedACL(d.Get(names.AttrAcl).(string))
		if acl == "" {
			acl = types.BucketCannedACLPrivate
		}
		input.ACL = &acl

		tflog.Info(ctx, "Will update bucket ACL", map[string]interface{}{
			"bucket_name": bucketName,
		})

		needsUpdate = true
	}

	//
	// Bucket Public Objects List.
	//
	if d.HasChange(names.AttrPublicListObjects) {
		publicListObjects := d.Get(names.AttrPublicListObjects).(bool)
		input.PublicObjectsListEnabled = &publicListObjects

		tflog.Info(ctx, "Will update bucket public list objects", map[string]interface{}{
			"bucket_name": bucketName,
		})

		needsUpdate = true
	}

	if needsUpdate {
		err := svc.UpdateBucket(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update bucket, %w", err))
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketPublicAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	acl := types.BucketCannedACLPrivate
	publicObjectsListEnabled := true
	input := &types.BucketUpdateInput{
		Bucket:                   bucketName,
		ACL:                      &acl,
		PublicObjectsListEnabled: &publicObjectsListEnabled,
	}

	tflog.Info(ctx, "Deleting bucket public access configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	err := svc.UpdateBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete bucket public access configuration, %w", err))
	}

	tflog.Info(ctx, "Bucket public access configuration deleted successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId("")
	return nil
}

func bucketCannedACL_Values() []string {
	var acl types.BucketCannedACL

	values := []string{}
	for _, value := range acl.Values() {
		values = append(values, string(value))
	}

	return values
}
