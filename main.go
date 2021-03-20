package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"gs308e-provider/gs308e"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return gs308e.Provider()
		},
	})
}
