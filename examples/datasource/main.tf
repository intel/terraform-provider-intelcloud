terraform {
  required_providers {
    idc = {
      source = "hashicorps/idc"
    }
  }
}


provider "idc" {
   region = "us-region-1"
}

data "idc_machine_images" "images" {
  most_recent = true
  filters = [
    {
      name = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

# data "idc_instance_types" "insttypes" {}

output "print_images" {
	value = data.idc_machine_images.images
}

# output "print_insttypes" {
# 	value = data.idc_instance_types.insttypes
# }