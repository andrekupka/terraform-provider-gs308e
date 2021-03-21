package gs308e

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSwitch() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSwitchRead,
		CreateContext: resourceSwitchCreate,
		UpdateContext: resourceSwitchUpdate,
		DeleteContext: resourceSwitchDelete,
		Schema: map[string]*schema.Schema{
			"mac": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.IsMACAddress,
			},
			"model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"cidr", "dhcp"},
				RequiredWith: []string{"gateway"},
				ValidateFunc: validation.IsCIDR,
			},
			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"cidr"},
				ValidateFunc: validation.IsIPv4Address,
			},
			"dhcp": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
			},
			"vlan_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      string(TaggedVLAN),
				ValidateFunc: validation.StringInSlice([]string{string(PortBasedVLAN), string(TaggedVLAN)}, false),
			},
			"port": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"pvid": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"tagged": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"untagged": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
					},
				},
			},
		},
	}
}
