terraform {
  required_providers {
    idc = {
      source = "cloud.intel.com/services/idc"
    }
    # random = {
    #   source = "hashicorp/random"
    #   version = "3.6.2"
    # }
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
  name = "testdemo99"
  availability_zone = "us-region-1a"
  tags = {
    environment = "Demo"
  }
}


resource "idc_kubernetes_cluster" "default" {
    async = true
    kubernetes_cluster = {
      name = "${local.name}-iks"
      availability_zone = local.availability_zone
      kubernetes_version  = "1.27"
      node_pools = [
        {
          name            = "${local.name}-ng"
          node_count      = 2
          node_type = "vm-small"
          user_data_url = ""        
        }
      ]
      ssh_public_key_names = ["var.ssh_key_name"]
      storage = {
          size_in_gb = 30
      } 
      load_balancer = {
          name = "${local.name}-lb"
          port = 443
          type = "public"
      }
    }
}

output "iks_order" {
	value = idc_kubernetes_cluster.default
}
