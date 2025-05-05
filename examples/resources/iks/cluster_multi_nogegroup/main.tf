/************* Terraform provider and version number ***************/
terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.19"
    }
  }
}

provider "intelcloud" {
  region = var.idc_region
}

provider "random" {
  # Configuration options
}

/**** Random provider used to generate random names (like pet names) ****/
resource "random_pet" "prefix" {}

locals {
  name              = random_pet.prefix.id
  availability_zone = var.idc_availability_zone
  tags = {
    environment = "demo"
  }
  # list of nodegroups user wants to create
  node_group_list = {
    "ng1" = {
      node_type  = "vm-spr-sml"
      node_count = 1
      ssh_keys   = ["rk-win-key", "rk-tf-key"]
    }
    "ng2" = {
      node_type  = "vm-spr-sml"
      node_count = 1
      ssh_keys   = ["rk-win-key"]
    }
  }
}

/***** Create an Intel Cloud IKS (Intel Kubernetes Service) cluster *******/
resource "intelcloud_iks_cluster" "cluster" {
  name               = "${local.name}-iks"
  kubernetes_version = var.kubernetes_version

  storage = {
    size_in_tb = var.size_in_tb
  }
  # specify custom timeouts for the resource
  timeouts {
    resource_timeout = "30m"
  }
}


resource "intelcloud_iks_node_group" "node_group" {
  for_each             = { for k, v in local.node_group_list : k => v if var.node_groups_create }
  name                 = try(each.value.name, each.key)
  cluster_uuid         = intelcloud_iks_cluster.cluster.id
  node_count           = try(each.value.node_count, var.node_group_defaults.node_count, 1)
  node_type            = try(each.value.node_type, var.node_group_defaults.node_type, "vm-spr-sml")
  ssh_public_key_names = try(each.value.ssh_keys, var.node_group_defaults.ssh_keys, [])
}
