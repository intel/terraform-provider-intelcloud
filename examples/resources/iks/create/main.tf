terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.10"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
}

provider "random" {
  # Configuration options
}

resource "random_pet" "prefix" {}

locals {
  name     = "${random_pet.prefix.id}"
  availability_zone = "us-region-2a"
  tags = {
    environment = "Demo"
  }
}

resource "intelcloud_iks_cluster" "cluster1" {
  name               = "${local.name}-iks"
  availability_zone  = local.availability_zone
  kubernetes_version = "1.28"

  storage = {
     size_in_tb = 30
  }
}

resource "intelcloud_iks_node_group" "ng1" {
  cluster_uuid = intelcloud_iks_cluster.cluster1.id
  name                 = "${local.name}-ng"
  node_count           = 1
  node_type            = "vm-spr-sml"
  userdata_url         = ""
  ssh_public_key_names = ["shrimac"]
  interfaces = [{
    name = "us-region-2a"
    vnet = "us-region-2a-default"
  }]
}

# resource "intelcloud_iks_lb" "lb1" {
#   cluster_uuid = idc_iks_cluster.cluster1.id
#   # cluster_uuid = "cl-ui2juj6vkq"
#   load_balancers = [
#     {
#       name     = "${local.name}-lb-pub2"
#       port     = 80
#       vip_type = "public"
#     }
#   ]
#   depends_on = [idc_iks_node_group.ng1]
# }

output "iks_order" {
	value = intelcloud_iks_cluster.cluster1
}
