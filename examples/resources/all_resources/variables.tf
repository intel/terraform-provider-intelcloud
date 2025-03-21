variable "ssh_pubkey_path" {
  type = string
}

variable "filesystem_description" {
  type    = string
  default = "demo filesystem"
}


variable "filesystem_size_in_gb" {
  type = number
}

variable "filesystem_type" {
  type    = string
  default = "ComputeGeneral"
}

variable "idc_region" {
  type    = string
  default = "region-2"
}

variable "idc_availability_zone" {
  type    = string
  default = "us-region-2a"
}

variable "os_image" {
  type    = string
  default = "ubuntu-2204-jammy-v20230122"
}

variable "instance_interface_spec" {
  type = map(any)
  default = {
    "name" = "eth0"
    "vnet" = "us-region-2a-default"
  }
}

variable "instance_types" {
  type = map(any)
  default = {
    "vm-small" = "vm-spr-sml"
    "vm-large" = "vm-spr-lrg"
  }
}

variable "instance_type" {
  type = string
}

variable "instance_count" {
  type    = number
  default = 1
}