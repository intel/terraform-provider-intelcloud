terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.8"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

locals {
  # name     = "${random_pet.prefix.id}"
  name              = "testdemo97"
  availability_zone = "us-region-1a"
  tags = {
    environment = "Demo"
  }
}

resource "idc_iks_node_group" "ng1" {
  cluster_uuid         = "cl-lc2ze6pu4i"
  name                 = "${local.name}-ng"
  node_count           = 2
  node_type            = "vm-spr-sml"
  userdata_url         = ""
  ssh_public_key_names = ["shrimac"]
  interfaces = [{
    name = "us-staging-3a"
    vnet = "us-staging-3a-default"
  }]
}

output "iks_order" {
  value = idc_iks_node_group.ng1
}
