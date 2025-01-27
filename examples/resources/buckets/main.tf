terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.6"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

resource "intelcloud_object_storage_bucket" "bucket1" {
  name      = "tf-demo99"
  versioned = false
}

output "bucket_order" {
  value = intelcloud_object_storage_bucket.bucket1
}
