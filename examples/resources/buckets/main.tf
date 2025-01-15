terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}

provider "idc" {
  region = "us-region-2"
}

resource "idc_object_storage_bucket" "bucket1" {
  name      = "tf-demo99"
  versioned = false
}

output "bucket_order" {
  value = idc_object_storage_bucket.bucket1
}
