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

data "intel-cloud_machine_images" "images" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

output "print_images" {
  value = data.intel-cloud_machine_images.images
}
