package gs308e

import (
	"context"
	"fmt"
	"github.com/andrekupka/gs308e/client"
	"github.com/andrekupka/gs308e/nsdp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"interface": {
				Type:     schema.TypeString,
				Required: true,
			},
			"passwords": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"gs308e_switch": resourceSwitch(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"gs308e_switch": dataSourceSwitch(),
		},
		ConfigureContextFunc: configureProviderContext,
	}
}

func configureProviderContext(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	interfaceName := d.Get("interface").(string)

	hardwareAddr, _, err := GetHardwareAndIpAddress(interfaceName)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to use interface",
			Detail:   fmt.Sprintf("Unable to use interface %s", interfaceName),
		})
		return nil, diags
	}

	exchange := nsdp.NewExchange(nsdp.GetHostAddress(net.IPv4zero), nsdp.GetBroadcastAddress())
	controller := client.NewController(hardwareAddr, exchange)
	controller.Start(context.Background())

	passwords := d.Get("passwords").(map[string]interface{})

	mappedPasswords := make(map[string]string)
	for mac, password := range passwords {
		mappedPasswords[mac] = password.(string)
	}

	return ProviderContext{
		Controller: controller,
		Passwords:  mappedPasswords,
	}, diags
}

func GetHardwareAndIpAddress(interfaceName string) (hardwareAddr net.HardwareAddr, ip net.IP, err error) {
	networkInterface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return
	}

	hardwareAddr = networkInterface.HardwareAddr

	addresses, err := networkInterface.Addrs()
	if err != nil {
		return
	}

	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok {
			ip = ipnet.IP
			break
		}
	}

	return
}
