terraform {
  required_providers {
    intel-cloud = {
      source = "intel/intel-cloud"
      version = "0.0.1"
    }
  }
}


provider "intel-cloud" {
  region = "us-region-1"
}

resource "intelcloud_filesystem" "example" {
  name = "tf-demo79"
  spec = {
    size_in_tb = 1
  }
}

output "filesystem_order" {
  value = intelcloud_filesystem.example
}
