terraform {
  required_providers {
    gs308e = {
      version = "0.1"
      source = "andrekupka.de/ntgr/gs308e"
    }
  }
}

provider "gs308e" {
  interface = "enp6s0"
  passwords = {
    "38:94:ed:1a:57:f1" = "E8mipwCwsrxkD26UtPQX"
  }
}

resource "gs308e_switch" "switch_tv" {
  mac = "38:94:ed:1a:57:f1"
  name = "switch-tv"

  cidr = "10.7.1.13/24"
  gateway = "10.7.1.1"

  vlan_mode = "tagged"

  port {
    id = 7
    pvid = 1
  }

  port {
    id = 8
    pvid = 1
  }
}

/*
data "gs308e_switch" "switch_pc_data" {
  mac = "38:94:ed:1a:57:06"
}

data "gs308e_switch" "switch_tv_data" {
  mac = "38:94:ed:1a:57:f1"
}*/


output "switch_tv_resource" {
  value = gs308e_switch.switch_tv
}
/*
output "switch_pc_data" {
  value = data.gs308e_switch.switch_pc_data
}

output "switch_tv_data" {
  value = data.gs308e_switch.switch_tv_data
}*/