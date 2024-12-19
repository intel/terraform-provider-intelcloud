terraform {
  required_providers {
    idc = {
      source = "hashicorps/idc"
    }
  }
}

provider "idc" {
   region = "us-region-2"
}

data "idc_filesystems" "example" {}

output "test_storages" {
	value = data.idc_filesystems.example
}

