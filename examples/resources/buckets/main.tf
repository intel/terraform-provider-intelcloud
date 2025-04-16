terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.11"
    }
  }
}


provider "intelcloud" {
  region = var.idc_region
}

resource "intelcloud_object_storage_bucket" "bucket1" {
  name      = "tf-bucket-demo"
  versioned = var.versioned
}

output "bucket_order" {
  value = intelcloud_object_storage_bucket.bucket1
}
