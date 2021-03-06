package linode

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/linode/linodego"
)

// Provider creates and manages the resources in a Linode configuration.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LINODE_TOKEN", nil),
				Description: "The token that allows you access to your Linode account",
			},
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("LINODE_URL", nil),
				Description:  "The HTTP(S) API address of the Linode API to use.",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"ua_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LINODE_UA_PREFIX", nil),
				Description: "An HTTP User-Agent Prefix to prepend in API requests.",
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LINODE_API_VERSION", nil),
				Description: "An HTTP User-Agent Prefix to prepend in API requests.",
			},

			"skip_instance_ready_poll": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip waiting for a linode_instance resource to be running.",
			},

			"min_retry_delay_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Minimum delay in milliseconds before retrying a request.",
			},
			"max_retry_delay_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum delay in milliseconds before retrying a request.",
			},

			"event_poll_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LINODE_EVENT_POLL_MS", 300),
				Description: "The rate in milliseconds to poll for events.",
			},

			"lke_event_poll_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "The rate in milliseconds to poll for LKE events.",
			},

			"lke_node_ready_poll_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     500,
				Description: "The rate in milliseconds to poll for an LKE node to be ready.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"linode_account":                dataSourceLinodeAccount(),
			"linode_domain":                 dataSourceLinodeDomain(),
			"linode_domain_record":          dataSourceLinodeDomainRecord(),
			"linode_firewall":               dataSourceLinodeFirewall(),
			"linode_image":                  dataSourceLinodeImage(),
			"linode_images":                 dataSourceLinodeImages(),
			"linode_instances":              dataSourceLinodeInstances(),
			"linode_instance_backups":       dataSourceLinodeInstanceBackups(),
			"linode_instance_type":          dataSourceLinodeInstanceType(),
			"linode_kernel":                 dataSourceLinodeKernel(),
			"linode_lke_cluster":            dataSourceLinodeLKECluster(),
			"linode_networking_ip":          dataSourceLinodeNetworkingIP(),
			"linode_nodebalancer":           dataSourceLinodeNodeBalancer(),
			"linode_nodebalancer_config":    dataSourceLinodeNodeBalancerConfig(),
			"linode_nodebalancer_node":      dataSourceLinodeNodeBalancerNode(),
			"linode_object_storage_cluster": dataSourceLinodeObjectStorageCluster(),
			"linode_profile":                dataSourceLinodeProfile(),
			"linode_region":                 dataSourceLinodeRegion(),
			"linode_sshkey":                 dataSourceLinodeSSHKey(),
			"linode_stackscript":            dataSourceLinodeStackscript(),
			"linode_user":                   dataSourceLinodeUser(),
			"linode_vlans":                  dataSourceLinodeVLANs(),
			"linode_volume":                 dataSourceLinodeVolume(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"linode_domain":                resourceLinodeDomain(),
			"linode_domain_record":         resourceLinodeDomainRecord(),
			"linode_firewall":              resourceLinodeFirewall(),
			"linode_image":                 resourceLinodeImage(),
			"linode_instance":              resourceLinodeInstance(),
			"linode_instance_ip":           resourceLinodeInstanceIP(),
			"linode_lke_cluster":           resourceLinodeLKECluster(),
			"linode_nodebalancer":          resourceLinodeNodeBalancer(),
			"linode_nodebalancer_config":   resourceLinodeNodeBalancerConfig(),
			"linode_nodebalancer_node":     resourceLinodeNodeBalancerNode(),
			"linode_object_storage_bucket": resourceLinodeObjectStorageBucket(),
			"linode_object_storage_key":    resourceLinodeObjectStorageKey(),
			"linode_object_storage_object": resourceLinodeObjectStorageObject(),
			"linode_rdns":                  resourceLinodeRDNS(),
			"linode_sshkey":                resourceLinodeSSHKey(),
			"linode_stackscript":           resourceLinodeStackscript(),
			"linode_token":                 resourceLinodeToken(),
			"linode_user":                  resourceLinodeUser(),
			"linode_volume":                resourceLinodeVolume(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}
	return provider
}

type ProviderMeta struct {
	Client linodego.Client
	Config *Config
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	config := &Config{
		AccessToken: d.Get("token").(string),
		APIURL:      d.Get("url").(string),
		APIVersion:  d.Get("api_version").(string),
		UAPrefix:    d.Get("ua_prefix").(string),

		SkipInstanceReadyPoll: d.Get("skip_instance_ready_poll").(bool),

		MinRetryDelayMilliseconds: d.Get("min_retry_delay_ms").(int),
		MaxRetryDelayMilliseconds: d.Get("max_retry_delay_ms").(int),

		EventPollMilliseconds:    d.Get("event_poll_ms").(int),
		LKEEventPollMilliseconds: d.Get("lke_event_poll_ms").(int),

		LKENodeReadyPollMilliseconds: d.Get("lke_node_ready_poll_ms").(int),
	}
	config.terraformVersion = terraformVersion
	client := config.Client()

	// Ping the API for an empty response to verify the configuration works
	if _, err := client.ListTypes(context.Background(), linodego.NewListOptions(100, "")); err != nil {
		return nil, fmt.Errorf("Error connecting to the Linode API: %s", err)
	}
	return &ProviderMeta{
		Client: client,
		Config: config,
	}, nil
}
