terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.9"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

resource "intelcloud_object_storage_bucket" "bucket1" {
  name      = "tf-demo-3"
  versioned = false
}

resource "intelcloud_object_storage_bucket_user" "user1" {
  name      = "tf-demo3-user"
  bucket_id = "${intelcloud_object_storage_bucket.bucket1.cloudaccount}-${intelcloud_object_storage_bucket.bucket1.name}"
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
# 	value = intelcloud_object_storage_bucket.bucket1
# }

output "bucket_user" {
  value = intelcloud_object_storage_bucket_user.user1
}
