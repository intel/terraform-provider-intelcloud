terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.6"
    }
  }
}


provider "intelcloud" {
  region = "us-staging-1"
}

resource "intelcloud_filesystem" "example" {
  name = "tf-demo79"
  spec = {
    size_in_tb = 2
  }
}

output "filesystem_order" {
  value = intelcloud_filesystem.example
}
