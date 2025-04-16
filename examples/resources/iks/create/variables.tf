variable "idc_region" {
  type = string
  default = "us-region-2"
}

variable "idc_availability_zone" {
  type = string
  default = "us-region-2a"
}

variable "kubernetes_version" {
  type    = string
  default = "1.30"
}

variable "ssh_public_key_names" {
  type = list(string)
  default = ["your-public-key-name"]
}

variable "node_type" {
  type    = string
  default = "vm-spr-sml"
}

variable "node_count" {
  type    = number
  default = 1
}

variable "size_in_tb" {
  type = number
  default = 30
}
