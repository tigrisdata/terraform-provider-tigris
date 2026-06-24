package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tigrisdata/terraform-provider-tigris/internal/names"
)

// newS3TestClient builds a Client whose S3 client talks to the given test
// server using path-style addressing.
func newS3TestClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(DefaultRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(serverURL)
		o.Region = DefaultRegion
		o.UsePathStyle = true
	})

	return &Client{
		s3Client:    svc,
		endpoint:    serverURL,
		credentials: aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"},
	}
}

func writeXML(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("failed to write response body: %v", err)
	}
}

// TestGetBucketLifecycle_EmptyConfig covers the Tigris behaviour where a bucket
// with no lifecycle returns HTTP 200 with an empty configuration rather than the
// NoSuchLifecycleConfiguration error AWS S3 returns. It must read as no rules.
func TestGetBucketLifecycle_EmptyConfig(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?><LifecycleConfiguration></LifecycleConfiguration>`)
	}))
	defer srv.Close()

	client := newS3TestClient(t, srv.URL)

	rules, err := client.GetBucketLifecycle(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(rules))
	}
}

// TestGetBucketLifecycle_NoSuchConfiguration covers the AWS-style path: a 404
// NoSuchLifecycleConfiguration must also read as no rules, not an error.
func TestGetBucketLifecycle_NoSuchConfiguration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, http.StatusNotFound, `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchLifecycleConfiguration</Code><Message>The lifecycle configuration does not exist</Message></Error>`)
	}))
	defer srv.Close()

	client := newS3TestClient(t, srv.URL)

	rules, err := client.GetBucketLifecycle(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected no rules, got %d", len(rules))
	}
}

func TestGetBucketLifecycle_WithRules(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Rule>
    <ID>to-ia</ID>
    <Filter><Prefix>nginx/</Prefix></Filter>
    <Status>Enabled</Status>
    <Transition><Days>30</Days><StorageClass>STANDARD_IA</StorageClass></Transition>
    <Expiration><Days>365</Days></Expiration>
  </Rule>
</LifecycleConfiguration>`)
	}))
	defer srv.Close()

	client := newS3TestClient(t, srv.URL)

	rules, err := client.GetBucketLifecycle(context.Background(), "my-bucket")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	r := rules[0]
	if aws.ToString(r.ID) != "to-ia" {
		t.Errorf("expected id to-ia, got %q", aws.ToString(r.ID))
	}
	if got := lifecycleRulePrefix(r); got != "nginx/" {
		t.Errorf("expected prefix nginx/, got %q", got)
	}
	if r.Status != s3types.ExpirationStatusEnabled {
		t.Errorf("expected status Enabled, got %q", r.Status)
	}
	if len(r.Transitions) != 1 || r.Transitions[0].Days == nil || *r.Transitions[0].Days != 30 {
		t.Errorf("expected one transition at 30 days, got %+v", r.Transitions)
	}
	if r.Transitions[0].StorageClass != s3types.TransitionStorageClassStandardIa {
		t.Errorf("expected STANDARD_IA transition, got %q", r.Transitions[0].StorageClass)
	}
	if r.Expiration == nil || r.Expiration.Days == nil || *r.Expiration.Days != 365 {
		t.Errorf("expected expiration at 365 days, got %+v", r.Expiration)
	}
}

func TestPutBucketLifecycle(t *testing.T) {
	t.Parallel()

	var gotQuery, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newS3TestClient(t, srv.URL)

	rules := []s3types.LifecycleRule{
		{
			ID:          aws.String("to-ia"),
			Status:      s3types.ExpirationStatusEnabled,
			Filter:      &s3types.LifecycleRuleFilterMemberPrefix{Value: "nginx/"},
			Transitions: []s3types.Transition{{Days: aws.Int32(30), StorageClass: s3types.TransitionStorageClassStandardIa}},
		},
	}

	if err := client.PutBucketLifecycle(context.Background(), "my-bucket", rules); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if !strings.Contains(gotQuery, "lifecycle") {
		t.Errorf("expected lifecycle query, got %q", gotQuery)
	}
}

func TestDeleteBucketLifecycle_Idempotent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "success", status: http.StatusNoContent},
		{name: "no such bucket", status: http.StatusNotFound, body: `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code><Message>not found</Message></Error>`},
		{name: "no such configuration", status: http.StatusNotFound, body: `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchLifecycleConfiguration</Code><Message>not found</Message></Error>`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.body == "" {
					w.WriteHeader(tc.status)
					return
				}
				writeXML(t, w, tc.status, tc.body)
			}))
			defer srv.Close()

			client := newS3TestClient(t, srv.URL)

			if err := client.DeleteBucketLifecycle(context.Background(), "my-bucket"); err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateLifecycleRuleMaps(t *testing.T) {
	t.Parallel()

	transition := func(days int, tier string) []interface{} {
		return []interface{}{map[string]interface{}{
			names.AttrDays:        days,
			names.AttrStorageTier: tier,
		}}
	}
	expiration := func(days int) []interface{} {
		return []interface{}{map[string]interface{}{names.AttrDays: days}}
	}
	rule := func(id string, t, e []interface{}) map[string]interface{} {
		return map[string]interface{}{
			names.AttrRuleID:     id,
			names.AttrTransition: t,
			names.AttrExpiration: e,
		}
	}

	cases := []struct {
		name    string
		rules   []interface{}
		wantErr bool
	}{
		{
			name:  "valid transition and expiration rules",
			rules: []interface{}{rule("a", transition(30, "STANDARD_IA"), nil), rule("b", nil, expiration(365))},
		},
		{
			name:  "zero-day transition is allowed",
			rules: []interface{}{rule("a", transition(0, "GLACIER_IR"), nil)},
		},
		{
			name:    "duplicate id",
			rules:   []interface{}{rule("dup", transition(30, "STANDARD_IA"), nil), rule("dup", nil, expiration(10))},
			wantErr: true,
		},
		{
			name:    "no action",
			rules:   []interface{}{rule("a", nil, nil)},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateLifecycleRuleMaps(tc.rules)
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}

// TestLifecycleRulesRoundTrip checks that expand and flatten are inverses over
// the supported surface (prefix filter, single transition, expiration).
func TestLifecycleRulesRoundTrip(t *testing.T) {
	t.Parallel()

	rules := []s3types.LifecycleRule{
		{
			ID:          aws.String("to-ia"),
			Status:      s3types.ExpirationStatusEnabled,
			Filter:      &s3types.LifecycleRuleFilterMemberPrefix{Value: "nginx/"},
			Transitions: []s3types.Transition{{Days: aws.Int32(30), StorageClass: s3types.TransitionStorageClassStandardIa}},
		},
		{
			ID:          aws.String("to-glacier-then-expire"),
			Status:      s3types.ExpirationStatusEnabled,
			Filter:      &s3types.LifecycleRuleFilterMemberPrefix{Value: "nginx/"},
			Transitions: []s3types.Transition{{Days: aws.Int32(90), StorageClass: s3types.TransitionStorageClassGlacierIr}},
			Expiration:  &s3types.LifecycleExpiration{Days: aws.Int32(365)},
		},
	}

	ruleResource := resourceTigrisBucketLifecycle().Schema[names.AttrRule].Elem.(*schema.Resource)
	set := schema.NewSet(schema.HashResource(ruleResource), nil)
	flat, err := flattenLifecycleRules(rules)
	if err != nil {
		t.Fatalf("flatten failed: %v", err)
	}
	for _, item := range flat {
		set.Add(item)
	}

	got := expandLifecycleRules(set)
	if len(got) != len(rules) {
		t.Fatalf("expected %d rules, got %d", len(rules), len(got))
	}

	byID := make(map[string]s3types.LifecycleRule, len(got))
	for _, r := range got {
		byID[aws.ToString(r.ID)] = r
	}

	ia := byID["to-ia"]
	if got := lifecycleRulePrefix(ia); got != "nginx/" {
		t.Errorf("expected prefix nginx/, got %q", got)
	}
	if len(ia.Transitions) != 1 || ia.Transitions[0].Days == nil || *ia.Transitions[0].Days != 30 {
		t.Errorf("unexpected transition: %+v", ia.Transitions)
	}

	glacier := byID["to-glacier-then-expire"]
	if glacier.Expiration == nil || glacier.Expiration.Days == nil || *glacier.Expiration.Days != 365 {
		t.Errorf("expected expiration 365, got %+v", glacier.Expiration)
	}
	if glacier.Transitions[0].StorageClass != s3types.TransitionStorageClassGlacierIr {
		t.Errorf("expected GLACIER_IR, got %q", glacier.Transitions[0].StorageClass)
	}
}

// TestFlattenLifecycleRules_DateBasedRejected ensures a date-based transition read
// from Tigris fails loudly rather than being silently flattened to days = 0, since
// this provider models only day-based transitions.
func TestFlattenLifecycleRules_DateBasedRejected(t *testing.T) {
	t.Parallel()

	rules := []s3types.LifecycleRule{
		{
			ID:     aws.String("date-based"),
			Status: s3types.ExpirationStatusEnabled,
			Filter: &s3types.LifecycleRuleFilterMemberPrefix{Value: ""},
			Transitions: []s3types.Transition{{
				Date:         aws.Time(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)),
				StorageClass: s3types.TransitionStorageClassGlacierIr,
			}},
		},
	}

	if _, err := flattenLifecycleRules(rules); err == nil {
		t.Fatal("expected an error for a date-based transition, got nil")
	}
}

// TestFlattenLifecycleRules_DateExpirationRejected ensures a date-based expiration
// read from Tigris fails loudly rather than being silently dropped, since this
// provider models only day-based expiration.
func TestFlattenLifecycleRules_DateExpirationRejected(t *testing.T) {
	t.Parallel()

	rules := []s3types.LifecycleRule{
		{
			ID:     aws.String("date-expiry"),
			Status: s3types.ExpirationStatusEnabled,
			Filter: &s3types.LifecycleRuleFilterMemberPrefix{Value: ""},
			Expiration: &s3types.LifecycleExpiration{
				Date: aws.Time(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
		},
	}

	if _, err := flattenLifecycleRules(rules); err == nil {
		t.Fatal("expected an error for a date-based expiration, got nil")
	}
}
