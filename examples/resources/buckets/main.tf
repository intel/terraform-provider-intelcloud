terraform {
  required_providers {
    intel-cloud = {
      source = "intel/intel-cloud"
      version = "0.0.1"
    }
  }
}


provider "intel-cloud" {
  region = "us-region-1"
}

resource "idc_object_storage_bucket" "bucket1" {
  name      = "tf-demo99"
  versioned = false
}

output "bucket_order" {
  value = idc_object_storage_bucket.bucket1
}
