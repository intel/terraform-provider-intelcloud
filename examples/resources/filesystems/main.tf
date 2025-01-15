terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}

provider "intelcloud" {
  region = "us-staging-1"
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
