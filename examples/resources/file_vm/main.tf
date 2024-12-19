terraform {
  required_providers {
    idc = {
      source = "cloud.intel.com/services/idc"
    }
  }
}

provider "idc" {
   region = var.idc_region
}

resource "idc_sshkey" "sshkey-1" {
   metadata = {
      name = var.ssh_key_name
    }
    spec = {
      ssh_public_key = file(var.ssh_pubkey_path)
      owner_email = var.ssh_user_email
    }
}

resource "idc_instance" "myinstance-1" {
  instance = {
    name = var.instance_name
    spec = {
      availability_zone = var.idc_availability_zone
      instance_type = var.instance_types[var.instance_type]
      machine_image = var.os_image
      interface_specs = [{
        name =  var.instance_interface_spec.name
        vnet = var.instance_interface_spec.vnet
      }]
      ssh_public_key_names = [var.ssh_key_name]
    }
  }
  depends_on = [idc_sshkey.sshkey-1]
}

resource "idc_filesystem" "fsvol-1" {
  filesystem = {
      name = var.filesystem_name
      description = var.filesystem_description
      spec = {
        size_in_gb = var.filesystem_size_in_gb
        filesystem_type = var.filesystem_type
      }
  }
}

resource "idc_object_storage_bucket" "bucket1" {
  name = "tf-bucket-1"
  versioned = false
}


output "filesystem_order" {
	value = idc_filesystem.example
}

