package types

import "time"

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
	Name             string         `json:"name"`
	CacheControl     string         `json:"cache_control"`
	ObjectRegions    string         `json:"object_regions"`
	MD               *BucketMD      `json:"md"`
	Shadow           *BucketShadow  `json:"shadow_bucket"`
	Website          *BucketWebsite `json:"website"`
	CreatedAt        time.Time      `json:"created_at"`
	InitialCreatedAt time.Time      `json:"initial_created_at"`
}

func (b *BucketMetadata) GetBucketCannedACL() BucketCannedACL {
	if b.MD == nil || b.MD.ACL == "" {
		return BucketCannedACLPrivate
	}

	return b.MD.ACL
}

type BucketMD struct {
	ACL BucketCannedACL `json:"X-Amz-Acl"`
}

type BucketWebsite struct {
	DomainName string `json:"domain_name"`
}

type BucketShadow struct {
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Region       string `json:"region"`
	Name         string `json:"name"`
	Endpoint     string `json:"endpoint"`
	WriteThrough bool   `json:"write_through"`
}

type BucketRequest struct {
	// The name of the bucket to create.
	Bucket string

	// The canned ACL to apply to the bucket.
	ACL BucketCannedACL

	// The website configuration for the bucket.
	Website *BucketWebsite
}

type BucketUpdateRequest struct {
	Website *BucketWebsite
}

type BucketUpdateResponse struct {
	// The success status of the update.
	Update string `json:"Update"`

	// The error message if the update failed.
	ErrorMessage string `json:"Message"`

	// The error code if the update failed.
	ErrorCode string `json:"Code"`
}
