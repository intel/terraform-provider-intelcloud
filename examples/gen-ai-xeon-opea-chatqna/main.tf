terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}

provider "idc" {
  region = var.idc_region
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

# resource "idc_sshkey" "example" {
#    metadata = {
#       name = var.ssh_key_name
#     }
#     spec = {
#       ssh_public_key = file(var.ssh_pubkey_path)
#       owner_email = var.ssh_user_email
#     }
# }

resource "idc_instance" "example" {
  name = var.instance_name
  spec = {
    instance_type        = var.instance_types[var.instance_type]
    machine_image        = var.os_image
    ssh_public_key_names = [var.ssh_key_name]
    user_data            = file("./cloud_init.yaml")
  }
  # depends_on = [idc_sshkey.example]
}

output "instance_order" {
  value = idc_instance.example
}
