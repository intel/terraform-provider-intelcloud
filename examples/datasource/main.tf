terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.5"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
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
