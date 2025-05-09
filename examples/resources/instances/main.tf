terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.15"
    }
  }
}


provider "intelcloud" {
  region = var.idc_region
}

data "intelcloud_machine_images" "image" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = [var.machine_image]
    }
  ]
}

resource "intelcloud_instance" "example" {
  name = "tf-demo-instance"
  spec = {
    instance_type        = var.instance_type
    machine_image        = var.machine_image
    ssh_public_key_names = [var.ssh_public_key_names]
  }
  timeouts {
    resource_timeout = "3m"
  }
}

output "instance_order" {
  value = intelcloud_instance.example
}
