terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.18"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
}

#provider "random" {
#  # Configuration options
#}
#
#/**** Random provider used to generate random names (like pet names) ****/
#resource "random_pet" "prefix" {}
#
#locals {
#  name              = random_pet.prefix.id
#  availability_zone = "us-region-2a"
#  tags = {
#    environment = "Demo"
#  }
#}

resource "intelcloud_iks_node_group" "ng1" {
  cluster_uuid = "cl-2agtscgbmy"
  name         = "bright-stinkbug-ng-new"
  node_count   = 4
  node_type    = "vm-spr-sml"
  #userdata_url         = "https://test.com/userdata"
  ssh_public_key_names = ["rk-win-key", "rk-tf-key"]
}

output "iks_order" {
  value = intelcloud_iks_node_group.ng1
}
