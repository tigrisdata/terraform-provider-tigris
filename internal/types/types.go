package types

type BucketCannedACL string

// Enum values for BucketCannedACL.
const (
	BucketCannedACLPrivate    BucketCannedACL = "private"
	BucketCannedACLPublicRead BucketCannedACL = "public-read"
)

func (BucketCannedACL) Values() []BucketCannedACL {
	return []BucketCannedACL{
		BucketCannedACLPrivate,
		BucketCannedACLPublicRead,
	}
}

type BucketMetadata struct {
	Name          string               `json:"name"`
	CacheControl  string               `json:"cache_control"`
	ObjectRegions string               `json:"object_regions"`
	MD            *BucketMD            `json:"md"`
	Shadow        *BucketShadowConfig  `json:"shadow_bucket"`
	Website       *BucketWebsiteConfig `json:"website"`
}

type BucketMD struct {
	ACL                      *BucketCannedACL `json:"X-Amz-Acl"`
	PublicObjectsListEnabled *string          `json:"x-amz-acl-public-list-objects-enabled"`
}

func (b *BucketMetadata) GetBucketCannedACL() BucketCannedACL {
	if b.MD == nil || b.MD.ACL == nil {
		return BucketCannedACLPrivate
	}

	return *b.MD.ACL
}

func (b *BucketMetadata) GetPublicObjectsListEnabled() bool {
	if b.MD == nil || b.MD.PublicObjectsListEnabled == nil {
		return true
	}

	if *b.MD.PublicObjectsListEnabled == "true" {
		return true
	}

	return false
}

type BucketWebsiteConfig struct {
	DomainName string `json:"domain_name"`
}

type BucketShadowConfig struct {
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Region       string `json:"region"`
	Name         string `json:"name"`
	Endpoint     string `json:"endpoint"`
	WriteThrough bool   `json:"write_through"`
}

// BucketUpdateInput is the input for the UpdateBucket function.
type BucketUpdateInput struct {
	// The name of the bucket to create.
	Bucket string

	// The canned ACL to apply to the bucket.
	ACL *BucketCannedACL

	// Whether to enable public object listing.
	PublicObjectsListEnabled *bool

	// The website configuration for the bucket.
	Website *BucketWebsiteConfig

	// The shadow bucket configuration for the bucket.
	Shadow *BucketShadowConfig
}

// BucketUpdateRequest is the request body for the UpdateBucket API.
type BucketUpdateRequest struct {
	Website *BucketWebsiteConfig `json:"website"`
	Shadow  *BucketShadowConfig  `json:"shadow_bucket"`
}

type BucketUpdateResponse struct {
	// The success status of the update.
	Update string `json:"Update"`

	// The error message if the update failed.
	ErrorMessage string `json:"Message"`

	// The error code if the update failed.
	ErrorCode string `json:"Code"`
}
