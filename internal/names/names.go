package names

const (
	// Attributes for the terraform resources.
	AttrBucket     = "bucket"
	AttrAcl        = "acl"
	AttrDomainName = "domain_name"

	// Headers for the requests to Tigiris.
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAmzContentSha = "X-Amz-Content-Sha256"
	HeaderAmzIdentityId = "S3-Identity-Id"
)
