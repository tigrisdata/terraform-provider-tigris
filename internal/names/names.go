package names

const (
	// Attributes for the terraform resources.
	AttrBucket             = "bucket"
	AttrAcl                = "acl"
	AttrPublicListObjects  = "public_list_objects"
	AttrDomainName         = "domain_name"
	AttrShadowAccessKey    = "shadow_access_key"
	AttrShadowSecretKey    = "shadow_secret_key"
	AttrShadowRegion       = "shadow_region"
	AttrShadowBucket       = "shadow_bucket"
	AttrShadowEndpoint     = "shadow_endpoint"
	AttrShadowWriteThrough = "shadow_write_through"

	// Headers for the requests to Tigris.
	HeaderContentType          = "Content-Type"
	HeaderAccept               = "Accept"
	HeaderAmzContentSha        = "X-Amz-Content-Sha256"
	HeaderAmzIdentityId        = "S3-Identity-Id"
	HeaderAmzAcl               = "X-Amz-Acl"
	HeaderAmzPublicListObjects = "X-Amz-Acl-Public-List-Objects-Enabled"
)
