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

resource "idc_object_storage_bucket" "bucket1" {
  name = "tf-demo99"
  versioned = false
}

output "bucket_order" {
	value = idc_object_storage_bucket.bucket1
}
