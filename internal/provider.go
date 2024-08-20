// provider.go
package internal

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// DefaultEndpoint is the default endpoint for Tigris object storage service.
	DefaultEndpoint = "https://fly.storage.tigris.dev"

	// DefaultRegion is the default region for Tigris object storage service.
	DefaultRegion = "auto"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AWS_ACCESS_KEY_ID", nil),
				Description: "The access key. It can also be sourced from the AWS_ACCESS_KEY_ID environment variable.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("AWS_SECRET_ACCESS_KEY", nil),
				Description: "The secret key. It can also be sourced from the AWS_SECRET_ACCESS_KEY environment variable.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     DefaultEndpoint,
				Description: "The endpoint for the Tigris object storage service.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tigris_bucket": resourceTigrisBucket(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)
	endpoint := d.Get("endpoint").(string)

	svc, err := NewClient(endpoint, DefaultRegion, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	return svc, nil
}
