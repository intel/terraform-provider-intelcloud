terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.11"
    }
  }
}


provider "intelcloud" {
  region = var.idc_region
}

resource "intelcloud_filesystem" "example" {
  name = "tf-filesystem-demo"
  spec = {
    size_in_tb = var.size_in_tb
  }
  timeouts {
    resource_timeout = "3m"
  }
}

output "filesystem_order" {
  value = intelcloud_filesystem.example
}
