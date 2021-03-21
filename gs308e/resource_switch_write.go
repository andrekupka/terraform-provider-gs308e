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

func updateName(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	name, ok := d.GetOk("name")
	if ok {
		return handle.SetDeviceName(ctx, &protocol.DeviceName{Name: name.(string)})
	}
	return nil
}

func updateLoopDetection(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	loopDetection := d.Get("loop_detection")
	return handle.SetLoopDetection(ctx, loopDetection.(bool))
}

func updatePVIDs(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	untypedPorts, portsOk := d.GetOk("port")
	if !portsOk {
		return nil
	}

	ports := untypedPorts.(*schema.Set)
	pvids := protocol.PVIDs{}

	for _, untypedPort := range ports.List() {
		port := untypedPort.(map[string]interface{})
		id := port["id"].(int)
		pvid, pvidOk := port["pvid"]
		if pvidOk {
			portId := uint8(id)
			pvids[portId] = protocol.PVID{
				Port:  portId,
				Value: uint16(pvid.(int)),
			}
		}
	}

	if len(pvids) > 0 {
		return handle.SetPVIDs(ctx, pvids)
	}

	return nil
}

func updateTaggedVLANs(
	ctx context.Context, d *schema.ResourceData, handle client.Switch, taggedVlans protocol.TaggedVLANs,
) error {/*
	untypedPorts, portsOk := d.GetOk("port")
	if !portsOk {
		return nil
	}

	ports := untypedPorts.(*schema.Set)

	for _, untypedPort := range ports.List() {
		port := untypedPort.(map[string]interface{})
		id := port["id"].(int)
	}*/

	return nil
}

func updatePorts(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	taggedVlans, err := handle.GetTaggedVLANs(ctx)
	if err != nil {
		return err
	}

	err = updatePVIDs(ctx, d, handle)
	if err != nil {
		return err
	}

	err = updateTaggedVLANs(ctx, d, handle, taggedVlans)
	if err != nil {
		return err
	}

	return nil
}

func updateNetwork(ctx context.Context, d *schema.ResourceData, handle client.Switch) error {
	cidr, cidrOk := d.GetOk("cidr")
	gateway, gatewayOk := d.GetOk("gateway")
	if cidrOk && gatewayOk {
		ip, ipNet, _ := net.ParseCIDR(cidr.(string))

		network := protocol.DeviceNetwork{
			IP:      ip,
			Mask:    ipNet.Mask,
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

	err = updateLoopDetection(ctx, d, handle)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updatePorts(ctx, d, handle)
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

	if d.HasChanges("loop_detection") {
		err := updateLoopDetection(ctx, d, handle)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("port") {
		err := updatePorts(ctx, d, handle)
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
