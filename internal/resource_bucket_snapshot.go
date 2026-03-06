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

	version, err := svc.CreateSnapshot(ctx, sourceBucket, snapshotName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create snapshot, %w", err))
	}

	tflog.Info(ctx, "Bucket snapshot created successfully", map[string]interface{}{
		"source_bucket":    sourceBucket,
		"snapshot_version": version,
	})

	d.SetId(fmt.Sprintf("%s:%s", sourceBucket, version))
	d.Set(names.AttrSnapshotVersion, version)

	return resourceBucketSnapshotRead(ctx, d, meta)
}

func resourceBucketSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	sourceBucket, version, err := parseBucketSnapshotID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "Reading bucket snapshot", map[string]interface{}{
		"source_bucket":    sourceBucket,
		"snapshot_version": version,
	})

	d.Set(names.AttrSourceBucket, sourceBucket)
	d.Set(names.AttrSnapshotVersion, version)

	// List snapshots to find this one and get its metadata.
	snapshots, err := svc.ListSnapshots(ctx, sourceBucket)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to list snapshots, %w", err))
	}

	// Look for snapshot by name. When snapshot_name is set, match by name.
	// When snapshot_name is not set, match snapshots with an empty name.
	snapshotName := d.Get(names.AttrSnapshotName).(string)
	for _, s := range snapshots {
		if s.Name == snapshotName {
			d.Set(names.AttrSnapshotCreatedAt, s.CreatedAt)
			return nil
		}
	}

	// If we didn't find a match, the snapshot still exists (we have the version).
	// The list API doesn't return versions directly, so we just preserve state.
	return nil
}

func resourceBucketSnapshotDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// There is no snapshot deletion API. Remove from state only.
	d.SetId("")
	return nil
}

func resourceBucketSnapshotImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	sourceBucket, version, err := parseBucketSnapshotID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set(names.AttrSourceBucket, sourceBucket)
	d.Set(names.AttrSnapshotVersion, version)

	return []*schema.ResourceData{d}, nil
}

func parseBucketSnapshotID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid snapshot ID format %q, expected {source_bucket}:{snapshot_version}", id)
	}
	return parts[0], parts[1], nil
}
