variable "idc_region" {
  type = string
  default = "us-region-2"
}

variable "machine_image" {
  type    = string
  default = "ubuntu-2204-jammy-v20230122"
}

variable "ssh_public_key_names" {
  type = string
  default = "your-public-key-name"
}

variable "instance_type" {
  type    = string
  default = "vm-spr-sml"
}