package gs308e

import (
	"context"
	"fmt"
	"github.com/andrekupka/gs308e/client"
	"github.com/andrekupka/gs308e/nsdp/protocol"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
)

type VLANMode string

const (
	PortBasedVLAN VLANMode = "port"
	TaggedVLAN VLANMode = "tagged"
)

func mapVLANMode(mode *protocol.VLANMode) VLANMode {
	switch mode.Code {
	case protocol.EasyPortBased, protocol.AdvancedPortBased:
		return PortBasedVLAN
	case protocol.EasyTagged, protocol.AdvancedTagged:
		return TaggedVLAN
	}
	panic(fmt.Sprintf("invalid vlan mode code: %d", mode.Code))
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
