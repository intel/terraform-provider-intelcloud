terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.8"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
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
  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for cloud-init to finish...'",
      "timeout=1200;",
      "while ! grep -q 'finish: modules-final: SUCCESS:' /var/log/cloud-init.log; do sleep 30;",
      "timeout=$((timeout - 30));",
      "if [ $timeout -le 0 ]; then echo 'Timeout waiting for cloud-init' >&2; exit 1; fi; done",
      "echo 'cloud-init finished.'"
    ]
    connection {
      type        = "ssh"
      user        = "ubuntu"
      private_key = file(var.ssh_privatekey_path)
      host        = intelcloud_instance.example.interfaces[0].address

      # Configure SSH to use the jump server
      bastion_host        = self.ssh_proxy.address
      bastion_user        = "guest"
      bastion_private_key = file(var.ssh_privatekey_path)
    }
  }
}

output "instance_order" {
  value = intelcloud_instance.example
}
