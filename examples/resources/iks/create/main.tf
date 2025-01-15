terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}

provider "idc" {
  region = "us-region-1"
}

# provider "random" {
#   # Configuration options
# }

# resource "random_pet" "prefix" {}

locals {
  # name     = "${random_pet.prefix.id}"
  name              = "iks-tf-2"
  availability_zone = "us-region-1a"
  tags = {
    environment = "Demo"
  }
}

resource "idc_iks_cluster" "cluster1" {
  async              = false
  name               = "${local.name}-iks"
  availability_zone  = local.availability_zone
  kubernetes_version = "1.27"

  storage = {
    size_in_gb = 30
  }
}

resource "idc_iks_node_group" "ng1" {
  cluster_uuid = idc_iks_cluster.cluster1.id
  # cluster_uuid = "cl-ui2juj6vkq"
  name                 = "${local.name}-ng"
  node_count           = 1
  node_type            = "vm-spr-sml"
  userdata_url         = ""
  ssh_public_key_names = ["shrimac"]
  interfaces = [{
    name = "us-region-1a"
    vnet = "us-region-1a-default"
  }]
}

resource "idc_iks_lb" "lb1" {
  cluster_uuid = idc_iks_cluster.cluster1.id
  # cluster_uuid = "cl-ui2juj6vkq"
  load_balancers = [
    {
      name     = "${local.name}-lb-pub2"
      port     = 80
      vip_type = "public"
    }
  ]
  depends_on = [idc_iks_node_group.ng1]
}

# output "iks_order" {
# 	value = idc_iks_cluster.cluster1
# }
