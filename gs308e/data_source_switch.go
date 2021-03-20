package gs308e

import (
	"context"
	"fmt"
	"github.com/andrekupka/gs308e/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
)

func dataSourceSwitch() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSwitchRead,
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
				Computed: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mask": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Computed: true,
			},/*
			"port": {
				Type: schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema {
						"number": {
							Type: schema.TypeInt,
							Computed: true,
						},
						"pvid": {
							Type: schema.TypeInt,
						},
					},
				},
			},*/
		},
	}
}

func dataSourceSwitchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	controller := m.(client.Controller)

	var diags diag.Diagnostics

	macString := d.Get("mac").(string)
	mac, err := net.ParseMAC(macString)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid MAC-Address provided",
			Detail:   fmt.Sprintf("%s is not a valid MAC-address", macString),
		})
		return diags
	}
	handle, err := controller.UseSwitch(ctx, mac, "")
	
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to contact switch",
			Detail:   fmt.Sprintf("Failed to contact switch %s, error was: %s", mac, err),
		})
		return diags
	}

	info, err := handle.GetInfo(ctx)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve switch info",
			Detail:   fmt.Sprintf("Could not retrieve switch info for %s, error was: %s", mac, err),
		})
		return diags
	}

	if err = d.Set("name", info.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("model", info.Model); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("ip", info.IpAddr.String()); err != nil {
		return diag.FromErr(err)
	}

	mask := maskToString(info.IpMask)
	if err = d.Set("mask", mask); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("gateway", info.GatewayIp.String()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(info.MacAddr.String())

	return diags
}
