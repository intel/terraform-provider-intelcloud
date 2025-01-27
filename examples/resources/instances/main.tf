terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.6"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
}

data "intelcloud_machine_images" "image" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

resource "intelcloud_instance" "example" {
  name = "tf-demo-instance"
  spec = {
    instance_type        = "vm-spr-sml"
    machine_image        = data.intelcloud_machine_images.image.result.name
    ssh_public_key_names = ["test-key"]
  }
}

output "instance_order" {
  value = intelcloud_instance.example
}
