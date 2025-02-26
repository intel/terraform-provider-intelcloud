terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.8"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

# data "cloudinit_config" "ansible" {
#   gzip          = true
#   base64_encode = true

#   part {
#     filename     = "cloud_init"
#     content_type = "text/cloud-config"
#     content = templatefile(
#       "cloud_init.yml", 
#       {
#         HUGGINGFACEHUB_API_TOKEN=var.huggingface_token
#       }
#     )
#   }
# }

# resource "intelcloud_sshkey" "example" {
#    metadata = {
#       name = var.ssh_key_name
#     }
#     spec = {
#       ssh_public_key = file(var.ssh_pubkey_path)
#       owner_email = var.ssh_user_email
#     }
# }

resource "intelcloud_instance" "example" {
  name = var.instance_name
  spec = {
    instance_type        = var.instance_types[var.instance_type]
    machine_image        = var.os_image
    ssh_public_key_names = [var.ssh_key_name]
    user_data            = file("./cloud_init.yaml")
  }
  # depends_on = [intelcloud_sshkey.example]
}

output "instance_order" {
  value = intelcloud_instance.example
}
