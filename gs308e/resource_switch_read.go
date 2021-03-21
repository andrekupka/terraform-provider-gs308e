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
	if err := d.Set("port_count", handle.PortCount()); err != nil {
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
		cidr := fmt.Sprintf("%s/%d", network.IP, prefixLength)

		if err = d.Set("cidr", cidr); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("gateway", network.Gateway.String()); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func determineDefinedPortIds(d *schema.ResourceData) []int {
	portIds := make([]int, 0)

	ports := d.Get("port").(*schema.Set)

	for _, unsafePort := range ports.List() {
		port := unsafePort.(map[string]interface{})
		portIds = append(portIds, port["id"].(int))
	}

	return portIds
}

func readVlanMode(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	mode, err := handle.GetVLANMode(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	modeString := mapVLANMode(mode)
	if err = d.Set("vlan_mode", modeString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func readPorts(ctx context.Context, d *schema.ResourceData, handle client.Switch) diag.Diagnostics {
	pvids, err := handle.GetPVIDs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	actualVlans, err := handle.GetTaggedVLANs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	portIds := determineDefinedPortIds(d)
	var ports []interface{}

	for _, portId := range portIds {
		byteId := uint8(portId)

		taggedVlans := make([]int, 0)
		untaggedVlans := make([]int, 0)

		for tag, vlan := range actualVlans {
			tagged, ok := vlan.Members[byteId]
			if ok {
				if tagged {
					taggedVlans = append(taggedVlans, int(tag))
				} else {
					untaggedVlans = append(untaggedVlans, int(tag))
				}
			}
		}

		port := map[string]interface{}{
			"id":       portId,
			"pvid":     pvids[byteId].Value,
			"tagged":   taggedVlans,
			"untagged": untaggedVlans,
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

	diags = readVlanMode(ctx, d, handle)
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
