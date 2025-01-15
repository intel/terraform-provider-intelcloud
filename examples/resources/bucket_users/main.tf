terraform {
  required_providers {
    intelcloud = {
      source = "hashicorps/intelcloud"
    }
  }
}

provider "idc" {
  region = "staging-1"
}

resource "idc_object_storage_bucket" "bucket1" {
  name      = "tf-demo-3"
  versioned = false
}

resource "idc_object_storage_bucket_user" "user1" {
  name      = "tf-demo3-user"
  bucket_id = "${idc_object_storage_bucket.bucket1.cloudaccount}-${idc_object_storage_bucket.bucket1.name}"
  allow_actions = [
    "GetBucketLocation",
    "GetBucketPolicy",
    "ListBucket",
    "ListBucketMultipartUploads",
    "ListMultipartUploadParts",
    "GetBucketTagging",
  ]
  allow_policies = {
    path_prefix = "/"
    policies = [
      "ReadBucket",
      "WriteBucket",
      "DeleteBucket",
    ]
  }
}

# output "bucket_order" {
# 	value = idc_object_storage_bucket.bucket1
# }

output "bucket_user" {
  value = idc_object_storage_bucket_user.user1
}
