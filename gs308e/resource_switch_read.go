package gs308e

import (
	"context"
	"fmt"
	"github.com/andrekupka/gs308e/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceSwitchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("READING-XXXXXXXXXXXXXXXXXXXXXXXXXX")
	handle, diags := getSwitch(ctx, d, m)
	if len(diags) > 0 {
		return diags
	}

	return readSwitch(ctx, d, handle)
}

func readBasic(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	if err := d.Set("model", handle.Model()); err != nil {
		return diag.FromErr(err)
	}

	name, err := handle.GetDeviceName(ctx)
	if err != nil {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve name",
			Detail:   fmt.Sprintf("Failed to retrieve name from switch"),
		}}
	}

	if err = d.Set("name", name.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func readNetwork(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	dhcp, err := handle.GetDHCP(ctx)
	if err != nil {
		return diag.Diagnostics{diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve dhcp config",
			Detail:   fmt.Sprintf("Failed to retrieve dhcp config from switch"),
		}}
	}

	if err = d.Set("dhcp", dhcp.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if !dhcp.Enabled {
		network, err := handle.GetDeviceNetwork(ctx)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve network",
				Detail:   fmt.Sprintf("Failed to retrieve network from switch"),
			}}
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
	return nil
}

type portIdWithVlanTags struct {
	id int
	tags []int
}

func determineDefinedPortsAndVlansIds(d *schema.ResourceData) []portIdWithVlanTags {
	portInfos := make([]portIdWithVlanTags, 0)

	ports := d.Get("port").(*schema.Set)

	for _, unsafePort := range ports.List() {
		port := unsafePort.(map[string]interface{})
		vlans := port["vlan"].(*schema.Set)

		var tags []int
		for _, unsafeVlan := range vlans.List() {
			vlan := unsafeVlan.(map[string]interface{})
			tags = append(tags, vlan["tag"].(int))
		}

		portInfos = append(portInfos, portIdWithVlanTags{
			id:   port["id"].(int),
			tags: tags,
		})
	}

	return portInfos
}

func readPorts(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	pvids, err := handle.GetPVIDs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	taggedVlans, err := handle.GetTaggedVLANs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	portsAndTags := determineDefinedPortsAndVlansIds(d)
	var ports []interface{}

	for _, portAndTag := range portsAndTags {
		byteId := uint8(portAndTag.id)

		var vlans []interface{}
		for _, tag := range portAndTag.tags {
			vlan := taggedVlans[uint16(tag)]
			tagged, ok := vlan.Members[byteId]
			if ok {
				vlans = append(vlans, map[string]interface{}{
					"tag":    tag,
					"tagged": tagged,
				})
			}
		}

		port := map[string]interface{}{
			"id":   portAndTag.id,
			"pvid": pvids[byteId].Value,
			"vlan": vlans,
		}
		ports = append(ports, port)
	}

	if ports == nil {
		return nil
	}
	if err = d.Set("port", ports); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func readSwitch(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = readBasic(ctx, d, handle)
	if diags != nil {
		return diags
	}

	diags = readNetwork(ctx, d, handle)
	if diags != nil {
		return diags
	}

	diags = readPorts(ctx, d, handle)
	if diags != nil {
		return diags
	}

	d.SetId(handle.HardwareAddr().String())

	return diags
}
