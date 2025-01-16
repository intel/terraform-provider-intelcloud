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

data "intel-cloud_machine_images" "image" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

resource "intel-cloud_instance" "example" {
  name = "tf-demo-instance"
  spec = {
    instance_type        = "vm-spr-sml"
    machine_image        = data.intelcloud_machine_images.image.result.name
    ssh_public_key_names = ["shrimac"]
  }
}

output "instance_order" {
  value = intel-cloud_instance.example
}
