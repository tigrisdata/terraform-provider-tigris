package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
	"github.com/tigrisdata/terraform-provider-tigris/internal/types"
)

func resourceTigrisBucketLifecycle() *schema.Resource {
	return &schema.Resource{
		Description:          "Provides a Tigris bucket lifecycle configuration resource. Each rule ages or expires objects matching an optional key prefix. Tigris supports a single transition per rule, so multi-tier aging is expressed as one rule per tier.",
		CreateWithoutTimeout: resourceBucketLifecycleCreate,
		ReadWithoutTimeout:   resourceBucketLifecycleRead,
		UpdateWithoutTimeout: resourceBucketLifecycleUpdate,
		DeleteWithoutTimeout: resourceBucketLifecycleDelete,

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
				ForceNew:    true,
				Description: "The name of the Tigris bucket.",
			},
			names.AttrRule: {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				MaxItems:    10,
				Description: "A lifecycle rule. Rules are unordered; each is identified by its id. A bucket can have at most 10 rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRuleID: {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Unique identifier for the rule. Up to 36 characters.",
							ValidateFunc: validation.StringLenBetween(1, 36),
						},
						names.AttrStatus: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      string(s3types.ExpirationStatusEnabled),
							Description:  "Whether the rule is applied. Enabled or Disabled.",
							ValidateFunc: validation.StringInSlice(expirationStatus_Values(), false),
						},
						names.AttrPrefix: {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Object key prefix the rule applies to. An empty prefix matches the whole bucket.",
						},
						names.AttrTransition: {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Transition the matching objects to a colder storage tier. At most one transition per rule.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDays: {
										Type:         schema.TypeInt,
										Required:     true,
										Description:  "Number of days after object creation before the transition. Use 0 to transition immediately.",
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrStorageTier: {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Destination storage tier. One of STANDARD_IA, GLACIER, GLACIER_IR.",
										ValidateFunc: validation.StringInSlice(transitionStorageTier_Values(), false),
									},
								},
							},
						},
						names.AttrExpiration: {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Expire (delete) the matching objects.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDays: {
										Type:         schema.TypeInt,
										Required:     true,
										Description:  "Number of days after object creation before expiration.",
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceBucketLifecycleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Get(names.AttrBucket).(string)

	if err := validateLifecycleConfig(d); err != nil {
		return diag.FromErr(err)
	}

	rules := expandLifecycleRules(d.Get(names.AttrRule).(*schema.Set))

	tflog.Info(ctx, "Creating bucket lifecycle configuration", map[string]interface{}{
		"bucket_name": bucketName,
		"rule_count":  len(rules),
	})

	if err := svc.PutBucketLifecycle(ctx, bucketName, rules); err != nil {
		return diag.FromErr(fmt.Errorf("unable to create bucket lifecycle configuration, %w", err))
	}

	d.SetId(bucketName)

	return resourceBucketLifecycleRead(ctx, d, meta)
}

func resourceBucketLifecycleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	tflog.Info(ctx, "Checking bucket existence", map[string]interface{}{
		"bucket_name": bucketName,
	})

	exists, err := svc.HeadBucket(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read bucket, %w", err))
	}
	if !exists {
		tflog.Warn(ctx, "Bucket not found, removing from state", map[string]interface{}{
			"bucket_name": bucketName,
		})

		d.SetId("")
		return nil
	}

	rules, err := svc.GetBucketLifecycle(ctx, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read bucket lifecycle configuration, %w", err))
	}
	if len(rules) == 0 {
		tflog.Warn(ctx, "Bucket has no lifecycle configuration, removing from state", map[string]interface{}{
			"bucket_name": bucketName,
		})

		d.SetId("")
		return nil
	}

	d.Set(names.AttrBucket, bucketName)

	flat, err := flattenLifecycleRules(rules)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(names.AttrRule, flat); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set lifecycle rules, %w", err))
	}

	return nil
}

func resourceBucketLifecycleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	if err := validateLifecycleConfig(d); err != nil {
		return diag.FromErr(err)
	}

	rules := expandLifecycleRules(d.Get(names.AttrRule).(*schema.Set))

	tflog.Info(ctx, "Updating bucket lifecycle configuration", map[string]interface{}{
		"bucket_name": bucketName,
		"rule_count":  len(rules),
	})

	if err := svc.PutBucketLifecycle(ctx, bucketName, rules); err != nil {
		return diag.FromErr(fmt.Errorf("unable to update bucket lifecycle configuration, %w", err))
	}

	return resourceBucketLifecycleRead(ctx, d, meta)
}

func resourceBucketLifecycleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	svc := meta.(*Client)

	bucketName := d.Id()

	tflog.Info(ctx, "Deleting bucket lifecycle configuration", map[string]interface{}{
		"bucket_name": bucketName,
	})

	if err := svc.DeleteBucketLifecycle(ctx, bucketName); err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete bucket lifecycle configuration, %w", err))
	}

	d.SetId("")
	return nil
}

// validateLifecycleConfig enforces the cross-field rules that a single attribute
// ValidateFunc cannot express: unique rule ids and at least one action per rule.
//
// This runs at apply time rather than in a CustomizeDiff because the SDKv2
// ResourceDiff does not reliably expose nested scalar values inside a TypeSet
// during planning, whereas ResourceData does.
func validateLifecycleConfig(d *schema.ResourceData) error {
	return validateLifecycleRuleMaps(d.Get(names.AttrRule).(*schema.Set).List())
}

func validateLifecycleRuleMaps(rules []interface{}) error {
	seen := make(map[string]struct{}, len(rules))

	for _, raw := range rules {
		m, _ := raw.(map[string]interface{})
		id, _ := m[names.AttrRuleID].(string)

		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate lifecycle rule id %q: rule ids must be unique", id)
		}
		seen[id] = struct{}{}

		transition, _ := m[names.AttrTransition].([]interface{})
		expiration, _ := m[names.AttrExpiration].([]interface{})

		if len(transition) == 0 && len(expiration) == 0 {
			return fmt.Errorf("lifecycle rule %q must set at least one of transition or expiration", id)
		}
	}

	return nil
}

func expandLifecycleRules(set *schema.Set) []s3types.LifecycleRule {
	list := set.List()
	rules := make([]s3types.LifecycleRule, 0, len(list))

	for _, raw := range list {
		m := raw.(map[string]interface{})

		id, _ := m[names.AttrRuleID].(string)
		status, _ := m[names.AttrStatus].(string)
		prefix, _ := m[names.AttrPrefix].(string)

		rule := s3types.LifecycleRule{
			ID:     aws.String(id),
			Status: s3types.ExpirationStatus(status),
			Filter: &s3types.LifecycleRuleFilterMemberPrefix{Value: prefix},
		}

		if t, ok := m[names.AttrTransition].([]interface{}); ok && len(t) == 1 {
			tm, _ := t[0].(map[string]interface{})
			storageTier, _ := tm[names.AttrStorageTier].(string)
			days, _ := tm[names.AttrDays].(int)
			rule.Transitions = []s3types.Transition{{
				Days:         aws.Int32(int32(days)),
				StorageClass: s3types.TransitionStorageClass(storageTier),
			}}
		}

		if e, ok := m[names.AttrExpiration].([]interface{}); ok && len(e) == 1 {
			em, _ := e[0].(map[string]interface{})
			days, _ := em[names.AttrDays].(int)
			rule.Expiration = &s3types.LifecycleExpiration{
				Days: aws.Int32(int32(days)),
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenLifecycleRules(rules []s3types.LifecycleRule) ([]interface{}, error) {
	out := make([]interface{}, 0, len(rules))

	for _, r := range rules {
		m := map[string]interface{}{
			names.AttrRuleID: aws.ToString(r.ID),
			names.AttrStatus: string(r.Status),
			names.AttrPrefix: lifecycleRulePrefix(r),
		}

		if len(r.Transitions) > 0 {
			t := r.Transitions[0]
			if t.Days == nil {
				// Tigris also accepts date-based transitions, which this provider does
				// not model. Fail loudly rather than silently flatten to days = 0.
				return nil, fmt.Errorf("lifecycle rule %q has a date-based transition, which is not supported by this provider; manage that rule outside Terraform or convert it to a days-based transition", aws.ToString(r.ID))
			}
			m[names.AttrTransition] = []interface{}{
				map[string]interface{}{
					names.AttrDays:        int(*t.Days),
					names.AttrStorageTier: string(t.StorageClass),
				},
			}
		}

		if r.Expiration != nil {
			if r.Expiration.Days == nil {
				// Tigris also accepts date-based and delete-marker expirations, which
				// this provider does not model. Fail loudly rather than drop the action.
				return nil, fmt.Errorf("lifecycle rule %q has a date-based expiration, which is not supported by this provider; manage that rule outside Terraform or use a days-based expiration", aws.ToString(r.ID))
			}
			m[names.AttrExpiration] = []interface{}{
				map[string]interface{}{
					names.AttrDays: int(*r.Expiration.Days),
				},
			}
		}

		out = append(out, m)
	}

	return out, nil
}

// lifecycleRulePrefix extracts the key prefix from a rule, handling both the
// modern Filter form and the deprecated top-level Prefix.
func lifecycleRulePrefix(r s3types.LifecycleRule) string {
	if r.Filter != nil {
		if p, ok := r.Filter.(*s3types.LifecycleRuleFilterMemberPrefix); ok {
			return p.Value
		}
	}
	if r.Prefix != nil {
		return *r.Prefix
	}

	return ""
}

func expirationStatus_Values() []string {
	return []string{
		string(s3types.ExpirationStatusEnabled),
		string(s3types.ExpirationStatusDisabled),
	}
}

func transitionStorageTier_Values() []string {
	return []string{
		string(types.StorageTierStandardIA),
		string(types.StorageTierGlacier),
		string(types.StorageTierGlacierIR),
	}
}
