package gs308e

import "github.com/andrekupka/gs308e/client"

type ProviderContext struct {
	Controller client.Controller
	Passwords  map[string]string
}
