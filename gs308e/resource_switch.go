package gs308e

import (
	"context"
	"github.com/andrekupka/gs308e/client"
	"github.com/andrekupka/gs308e/nsdp/protocol"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net"
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
			},
			"model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"prefix_length", "gateway"},
			},
			"prefix_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"ip", "gateway"},
			},
			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"ip", "prefix_length"},
			},
			"dhcp": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ip", "prefix_length", "gateway"},
			},
			"port": {
				Type:     schema.TypeList,
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
							Computed: true,
						},
						"vlan": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag": {
										Type: schema.TypeInt,
										Required: true,
									},
									"tagged": {
										Type: schema.TypeBool,
										Optional: true,
										Default: false,
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

func updateName(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	name, ok := d.GetOk("name")
	if ok {
		return handle.SetDeviceName(ctx, &protocol.DeviceName{Name: name.(string)})
	}
	return nil
}

func updateNetwork(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	ip, ipOk := d.GetOk("ip")
	prefixLength, prefixOk := d.GetOk("prefix_length")
	gateway, gatewayOk := d.GetOk("gateway")
	if ipOk && prefixOk && gatewayOk {
		mask := net.CIDRMask(prefixLength.(int), 32)
		network := protocol.DeviceNetwork{
			IP:      net.ParseIP(ip.(string)),
			Mask:    mask,
			Gateway: net.ParseIP(gateway.(string)),
		}

		err := handle.SetDeviceNetwork(ctx, &network)
		if err != nil {
			return err
		}
	}

	dhcp, dhcpOk := d.GetOk("dhcp")
	if dhcpOk {
		return handle.SetDHCP(ctx, &protocol.DHCP{Enabled: dhcp.(bool)})
	}

	return nil
}

func resourceSwitchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("CREATING-XXXXXXXXXXXXXXXXXXXXXXXXXX")
	handle, diags := getSwitch(ctx, d, m)
	if len(diags) > 0 {
		return diags
	}

	err := updateName(ctx, d, handle)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateNetwork(ctx, d, handle)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(handle.HardwareAddr().String())

	return readSwitch(ctx, d, handle)
}

func resourceSwitchUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("UPDATING-XXXXXXXXXXXXXXXXXXXXXXXXXX")
	handle, diags := getSwitch(ctx, d, m)
	if len(diags) > 0 {
		return diags
	}

	if d.HasChanges("name") {
		err := updateName(ctx, d, handle)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("ip", "prefix_length", "gateway", "dhcp") {
		err := updateNetwork(ctx, d, handle)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(handle.HardwareAddr().String())

	return readSwitch(ctx, d, handle)
}

func resourceSwitchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("DELETING-XXXXXXXXXXXXXXXXXXXXXXXXXX")
	var diags diag.Diagnostics
	// we just virtually delete the switch but cannot reset switch config
	d.SetId("")
	return diags
}
