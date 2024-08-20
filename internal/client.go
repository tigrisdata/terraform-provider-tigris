package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/tigrisdata/terraform-provider-tigris/internal/types"
)

const (
	// DefaultEndpoint is the default endpoint for Tigris object storage service.
	DefaultEndpoint = "https://fly.storage.tigris.dev"

	// DefaultRegion is the default region for Tigris object storage service.
	DefaultRegion = "auto"
)

type Client struct {
	cfg         aws.Config
	signer      *v4.Signer
	credentials aws.Credentials
	endpoint    string
	httpClient  *http.Client
	s3Client    *s3.Client
}

func NewClient(endpoint, accessKeyID, secretAccessKey string) (*Client, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(DefaultRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	// Create a signer
	signer := v4.NewSigner()

	// Create S3 service client
	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = DefaultRegion
	})

	return &Client{
		cfg:    cfg,
		signer: signer,
		credentials: aws.Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
		endpoint:   endpoint,
		httpClient: &http.Client{},
		s3Client:   svc,
	}, nil
}

func (c *Client) CreateBucket(ctx context.Context, input *types.BucketRequest) error {
	if err := validateBucketRequest(input); err != nil {
		return err
	}

	_, err := c.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(input.Bucket),
	})
	if err != nil {
		return err
	}

	return c.UpdateBucket(ctx, input)
}

func (c *Client) UpdateBucket(ctx context.Context, input *types.BucketRequest) error {
	if err := validateBucketRequest(input); err != nil {
		return err
	}

	// Set all the bucket attributes that need to be updated
	upReq := &types.BucketUpdateRequest{
		Website: input.Website,
	}
	body, err := json.Marshal(upReq)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.bucketURL(input.Bucket, nil), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create update request: %v", err)
	}

	// ACL updates are done through the header
	req.Header.Set("X-Amz-Acl", string(input.ACL))

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return fmt.Errorf("failed to send update request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with code: %d", resp.StatusCode)
	}

	var upResp types.BucketUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&upResp)
	if err != nil {
		return fmt.Errorf("failed to read update response: %v", err)
	}
	if upResp.Update != "success" && upResp.Update != "not modified" {
		return fmt.Errorf("update failed with error: %s", upResp.ErrorMessage)
	}

	return nil
}

func (c *Client) HeadBucket(ctx context.Context, bucketName string) (bool, error) {
	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	exists := true
	if err != nil {
		var notFoundErr *s3types.NotFound
		if ok := errors.As(err, &notFoundErr); ok {
			exists = false
			return exists, nil
		}
	}

	return exists, err
}

func (c *Client) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := c.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})

	return err
}

func (c *Client) GetBucketMetadata(ctx context.Context, bucketName string) (*types.BucketMetadata, error) {
	params := map[string]string{
		"metadata": "",
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.bucketURL(bucketName, params), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with code: %d", resp.StatusCode)
	}

	// Parse the response body into a BucketMetadata struct
	var metadata types.BucketMetadata
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to read bucket metadata: %v", err)
	}

	return &metadata, nil
}

func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	maxRetries := 5
	backoffDelay := 3 * time.Second
	maxBackoffDelay := 60 * time.Second

	var resp *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		// Clone the request to avoid issues with mutated request objects
		clonedReq, err := cloneRequest(req)
		if err != nil {
			return nil, fmt.Errorf("failed to clone request: %v", err)
		}

		resp, err = c.doSignedRequest(clonedReq)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}

		// Check if the response status code indicates a server-side error (5xx)
		if resp.StatusCode >= 500 {
			resp.Body.Close()

			// Exponential backoff before retrying
			time.Sleep(backoffDelay)
			backoffDelay *= 2 // Double the delay for each retry
			if backoffDelay > maxBackoffDelay {
				backoffDelay = maxBackoffDelay
			}

			continue
		}

		// Break out of the loop if the request was successful
		break
	}

	return resp, err
}

func (c *Client) doSignedRequest(req *http.Request) (*http.Response, error) {
	// Sign the request
	err := c.signRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %v", err)
	}

	// Send the signed request using the wrapped http.Client
	return c.httpClient.Do(req)
}

func (c *Client) bucketURL(bucketName string, queryParams map[string]string) string {
	baseURL := fmt.Sprintf("%s/%s", c.endpoint, bucketName)
	if len(queryParams) == 0 {
		return baseURL
	}

	// Add query parameters to the URL
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}

	return fmt.Sprintf("%s?%s", baseURL, query.Encode())
}

func (c *Client) signRequest(req *http.Request) error {
	// Get the current time for the request
	now := time.Now()

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Buffer the request body if it exists
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Determine the payload hash (empty string if no payload)
	payloadHash := hex.EncodeToString(sha256.New().Sum(bodyBytes))

	// Sign the request using the signer
	err := c.signer.SignHTTP(context.TODO(), c.credentials, req, payloadHash, "s3", DefaultRegion, now)
	if err != nil {
		return fmt.Errorf("failed to sign request: %v", err)
	}

	return nil
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	// Create a shallow copy of the request
	clonedReq := req.Clone(req.Context())

	// Clone the body if it exists and is seekable
	if req.Body != nil {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}

		// Restore the original body to be read again
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		// Set the cloned request's body
		clonedReq.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	}

	return clonedReq, nil
}

func validateBucketRequest(input *types.BucketRequest) error {
	if input.Bucket == "" {
		return errors.New("bucket name is required")
	}

	return nil
}
