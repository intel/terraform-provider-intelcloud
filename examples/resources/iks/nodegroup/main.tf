terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.15"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
}

locals {
  name              = random_pet.prefix.id
  availability_zone = "us-region-2a"
  tags = {
    environment = "Demo"
  }
}

resource "intelcloud_iks_node_group" "ng1" {
  cluster_uuid         = "<your-cluster-uuid>"
  name                 = "${local.name}-ng"
  node_count           = 2
  node_type            = "vm-spr-sml"
  userdata_url         = ""
  ssh_public_key_names = ["rk-win-key"]
}

output "iks_order" {
  value = intelcloud_iks_node_group.ng1
}
