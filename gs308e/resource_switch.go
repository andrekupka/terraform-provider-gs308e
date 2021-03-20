package gs308e

import (
	"context"
	"fmt"
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
				Computed: true,
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
					},
				},
			},
		},
	}
}

func getSwitch(ctx context.Context, d *schema.ResourceData, m interface{}) (client.Switch, diag.Diagnostics) {
	config := m.(ProviderContext)
	controller := config.Controller

	var diags diag.Diagnostics

	macString := d.Get("mac").(string)
	mac, err := net.ParseMAC(macString)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid MAC-Address provided",
			Detail:   fmt.Sprintf("%s is not a valid MAC-address", macString),
		})
		return nil, diags
	}

	password, ok := config.Passwords[macString]
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "No password defined for switch",
			Detail:   fmt.Sprintf("There is no password for switch %s", macString),
		})
		return nil, diags
	}

	handle, err := controller.UseSwitch(ctx, mac, password)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed initial contact with switch",
			Detail:   fmt.Sprintf("Initial contact with switch %s has failed", macString),
		})
		return nil, diags
	}
	return handle, diags
}

func resourceSwitchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READING-XXXXXXXXXXXXXXXXXXXXXXXXXX")
	handle, diags := getSwitch(ctx, d, m)
	if len(diags) > 0 {
		return diags
	}

	return readSwitch(ctx, d, handle)
}

func determineDefinedPortIds(d *schema.ResourceData) []int {
	var ids []int

	ports := d.Get("port").([]interface{})
	for _, unsafePort := range ports {
		port := unsafePort.(map[string]interface{})
		ids = append(ids, port["id"].(int))
	}

	return ids
}

func readSwitch(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := d.Set("model", handle.Model()); err != nil {
		return diag.FromErr(err)
	}

	name, err := handle.GetDeviceName(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve name",
			Detail:   fmt.Sprintf("Failed to retrieve name from switch"),
		})
		return diags
	}

	if err = d.Set("name", name.Name); err != nil {
		return diag.FromErr(err)
	}

	dhcp, err := handle.GetDHCP(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve dhcp config",
			Detail:   fmt.Sprintf("Failed to retrieve dhcp config from switch"),
		})
		return diags
	}

	if err = d.Set("dhcp", dhcp.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if !dhcp.Enabled {
		network, err := handle.GetDeviceNetwork(ctx)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve network",
				Detail:   fmt.Sprintf("Failed to retrieve network from switch"),
			})
			return diags
		}

		prefixLength, _ := network.Mask.Size()

		if err = d.Set("ip", network.IP.String()); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("prefix_length", prefixLength); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("gateway", network.Gateway.String()); err != nil {
			return diag.FromErr(err)
		}
	}

	pvids, err := handle.GetPVIDs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	definedPortIds := determineDefinedPortIds(d)
	var ports []interface{}

	for _, portId := range definedPortIds {
		port := map[string]interface{}{
			"id":   portId,
			"pvid": pvids[uint8(portId)].Value,
		}
		ports = append(ports, port)
	}

	if err = d.Set("port", ports); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(handle.HardwareAddr().String())

	return diags
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

/*
func updatePorts(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	portsValue, ok := d.GetOk("port")
	if !ok {
		return nil
	}

	ports := portsValue.([]map[string]interface{})

	return nil
}*/

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
