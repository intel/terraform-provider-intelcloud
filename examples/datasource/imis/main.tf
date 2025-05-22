terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.19"
    }
  }
}


provider "intelcloud" {
  region = "us-staging-1"
  endpoints = {
    api  = "https://us-staging-1-sdk-api.eglb.intel.com"
    auth = "https://client-token.staging.api.idcservice.net"
  }
}

// Commenting this temporarily as it is won't work without merging the PR for imis datasource
#data "intelcloud_imis" "filtered_imis" {
#  clusteruuid = "<your-cluster-uuid>"
#  filters = [
#    {
#      name   = "instance-type"
#      values = ["vm-spr-sml"]
#    },
#  ]
#}
#
#output "print_images" {
#  value = data.intelcloud_imis.filtered_imis
#}
