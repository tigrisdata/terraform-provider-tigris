package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
)

func resourceTigrisBucketSnapshot() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket snapshot resource. This creates a point-in-time snapshot of a Tigris bucket.",
		CreateWithoutTimeout: resourceBucketSnapshotCreate,
		ReadWithoutTimeout:   resourceBucketSnapshotRead,
		DeleteWithoutTimeout: resourceBucketSnapshotDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceBucketSnapshotImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrSourceBucket: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the source bucket to snapshot.",
			},
			names.AttrSnapshotName: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name for the snapshot.",
			},
			names.AttrSnapshotVersion: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version identifier of the snapshot.",
			},
			names.AttrSnapshotCreatedAt: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the snapshot was created.",
			},
		},
	}
}

func resourceBucketSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	sourceBucket := d.Get(names.AttrSourceBucket).(string)
	snapshotName := d.Get(names.AttrSnapshotName).(string)

	tflog.Info(ctx, "Creating bucket snapshot", map[string]interface{}{
		"source_bucket": sourceBucket,
		"snapshot_name": snapshotName,
	})

	snapshotVersion, err := svc.CreateSnapshot(ctx, sourceBucket, snapshotName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create snapshot, %w", err))
	}

	tflog.Info(ctx, "Bucket snapshot created successfully", map[string]interface{}{
		"source_bucket":    sourceBucket,
		"snapshot_version": snapshotVersion,
	})

	d.SetId(fmt.Sprintf("%s:%s", sourceBucket, snapshotVersion))
	d.Set(names.AttrSnapshotVersion, snapshotVersion)

	return resourceBucketSnapshotRead(ctx, d, meta)
}

func resourceBucketSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	sourceBucket, snapshotVersion, err := parseBucketSnapshotID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "Reading bucket snapshot", map[string]interface{}{
		"source_bucket":    sourceBucket,
		"snapshot_version": snapshotVersion,
	})

	d.Set(names.AttrSourceBucket, sourceBucket)
	d.Set(names.AttrSnapshotVersion, snapshotVersion)

	// Look up the snapshot by name to populate metadata.
	snapshotName := d.Get(names.AttrSnapshotName).(string)
	if snapshotName != "" {
		snapshot, err := svc.GetSnapshotByName(ctx, sourceBucket, snapshotName)
		if err != nil {
			tflog.Warn(ctx, "Snapshot not found by name, preserving state", map[string]interface{}{
				"source_bucket": sourceBucket,
				"snapshot_name": snapshotName,
			})
			return nil
		}
		d.Set(names.AttrSnapshotName, snapshot.Name)
		d.Set(names.AttrSnapshotCreatedAt, snapshot.CreatedAt)
	}

	return nil
}

func resourceBucketSnapshotDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// There is no snapshot deletion API. Remove from state only.
	d.SetId("")
	return nil
}

func resourceBucketSnapshotImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	sourceBucket, snapshotVersion, snapshotName, err := parseBucketSnapshotImportID(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%s:%s", sourceBucket, snapshotVersion))
	d.Set(names.AttrSourceBucket, sourceBucket)
	d.Set(names.AttrSnapshotVersion, snapshotVersion)
	d.Set(names.AttrSnapshotName, snapshotName)

	return []*schema.ResourceData{d}, nil
}

func parseBucketSnapshotID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid snapshot ID format %q, expected {source_bucket}:{snapshot_version}", id)
	}
	return parts[0], parts[1], nil
}

func parseBucketSnapshotImportID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("invalid snapshot import ID format %q, expected {source_bucket}:{snapshot_version}:{snapshot_name}", id)
	}
	return parts[0], parts[1], parts[2], nil
}
