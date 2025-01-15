terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

data "intelcloud_machine_images" "images" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

output "print_images" {
  value = data.intelcloud_machine_images.images
}
