terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.8"
    }
  }
}


provider "intelcloud" {
  region = "us-region-3"
}

resource "intelcloud_filesystem" "example" {
  name = "tf-demo79"
  spec = {
    size_in_tb = 3
  }
}

output "filesystem_order" {
  value = intelcloud_filesystem.example
}
