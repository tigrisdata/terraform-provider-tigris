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

func resourceTigrisBucketFork() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket fork resource. This creates a new bucket as a fork of an existing bucket, optionally from a specific snapshot.",
		CreateWithoutTimeout: resourceBucketForkCreate,
		ReadWithoutTimeout:   resourceBucketForkRead,
		DeleteWithoutTimeout: resourceBucketForkDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the new forked bucket.",
			},
			names.AttrForkSourceBucket: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the source bucket to fork from.",
			},
			names.AttrForkSourceBucketSnapshot: {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The snapshot version to fork from. If not specified, forks from the current state.",
			},
			names.AttrForkCreatedAt: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the fork was created.",
			},
		},
	}
}

func resourceBucketForkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)
	forkSourceBucket := d.Get(names.AttrForkSourceBucket).(string)

	if err := validBucketName(bucketName); err != nil {
		return diag.FromErr(fmt.Errorf("invalid bucket name, %w", err))
	}

	input := &types.BucketUpdateInput{
		Bucket:           bucketName,
		ForkSourceBucket: &forkSourceBucket,
	}

	if v, ok := d.GetOk(names.AttrForkSourceBucketSnapshot); ok {
		snapshot := v.(string)
		input.ForkSourceSnapshot = &snapshot
	}

	tflog.Info(ctx, "Creating bucket fork", map[string]interface{}{
		"bucket_name":        bucketName,
		"fork_source_bucket": forkSourceBucket,
	})

	err := svc.CreateBucket(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket fork, %w", err))
	}

	tflog.Info(ctx, "Bucket fork created successfully", map[string]interface{}{
		"bucket_name": bucketName,
	})

	d.SetId(bucketName)

	return resourceBucketForkRead(ctx, d, meta)
}

func resourceBucketForkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	tflog.Info(ctx, "Reading bucket fork", map[string]interface{}{
		"bucket_name": bucketName,
	})

	exists, err := svc.HeadBucket(ctx, bucketName)
	if !exists {
		tflog.Warn(ctx, "Forked bucket not found, removing from state", map[string]interface{}{
			"bucket_name": bucketName,
		})
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read forked bucket, %w", err))
	}

	d.Set(names.AttrBucket, bucketName)

	// Fetch metadata to get fork info.
	metadata, err := svc.GetBucketMetadata(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read bucket metadata, %w", err))
	}

	if metadata.ForkInfo != nil && len(metadata.ForkInfo.Parents) > 0 {
		parent := metadata.ForkInfo.Parents[0]
		d.Set(names.AttrForkSourceBucket, parent.BucketName)
		if parent.Snapshot != "" {
			d.Set(names.AttrForkSourceBucketSnapshot, parent.Snapshot)
		}
		d.Set(names.AttrForkCreatedAt, parent.ForkCreatedAt)
	}

	return nil
}

func resourceBucketForkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	tflog.Info(ctx, "Deleting forked bucket", map[string]interface{}{
		"bucket_name": bucketName,
	})

	// Forks are regular buckets, so delete normally.
	err := svc.DeleteBucket(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete forked bucket, %w", err))
	}

	d.SetId("")
	return nil
}
