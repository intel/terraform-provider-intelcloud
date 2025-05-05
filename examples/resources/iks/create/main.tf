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
}

/***** Create an Intel Cloud IKS (Intel Kubernetes Service) cluster *******/
resource "intelcloud_iks_cluster" "cluster1" {
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

/********** Create a node group to attach to the Kubernetes cluster **********/
resource "intelcloud_iks_node_group" "ng1" {
  cluster_uuid         = intelcloud_iks_cluster.cluster1.id
  name                 = "${local.name}-ng"
  node_count           = var.node_count
  node_type            = var.node_type
  userdata_url         = ""
  ssh_public_key_names = var.ssh_public_key_names
  timeouts {
    resource_timeout = "15m"
  }
}

# Output the details of the created IKS cluster
output "iks_order" {
  value = intelcloud_iks_cluster.cluster1
}

#################### Loadbalancer is not supported for now #####################
/*
 resource "intelcloud_iks_lb" "lb1" {
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
*/

