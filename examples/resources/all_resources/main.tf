terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.15"
    }
  }
}


provider "intelcloud" {
  region = "us-region-1"
}

locals {
  name = "apollo11"
  tags = {
    environment = "Demo"
  }
}

data "intelcloud_machine_images" "image" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

resource "intelcloud_sshkey" "sshkey-1" {
  metadata = {
    name = "${local.name}-sshkey"
  }
  spec = {
    ssh_public_key = file(var.ssh_pubkey_path)
  }
}

resource "intelcloud_instance" "myinstance-1" {
  name = "${local.name}-instance"
  spec = {
    instance_type = var.instance_types[var.instance_type]
    machine_image = data.intelcloud_machine_images.image.result.name
    interface_specs = [{
      name = var.instance_interface_spec.name
      vnet = var.instance_interface_spec.vnet
    }]
    ssh_public_key_names = [intelcloud_sshkey.sshkey-1.metadata.name]
  }
  depends_on = [intelcloud_sshkey.sshkey-1]
}

resource "intelcloud_filesystem" "fsvol-1" {
  name = "${local.name}-filevol"

  spec = {
    size_in_tb = var.filesystem_size_in_gb
  }
}

resource "intelcloud_object_storage_bucket" "bucket1" {
  name      = "${local.name}-bucket"
  versioned = false
}

resource "intelcloud_object_storage_bucket_user" "user1" {
  name      = "${intelcloud_object_storage_bucket.bucket1.name}-user"
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

resource "intelcloud_iks_cluster" "cluster1" {
  name               = "${local.name}-iks"
  kubernetes_version = "1.27"

  storage = {
    size_in_tb = 30
  }
}

resource "intelcloud_iks_node_group" "ng1" {
  cluster_uuid         = intelcloud_iks_cluster.cluster1.id
  name                 = "${local.name}-ng"
  node_count           = 1
  node_type            = var.instance_types[var.instance_type]
  userdata_url         = ""
  ssh_public_key_names = [intelcloud_sshkey.sshkey-1.metadata.name]
}

resource "intelcloud_iks_lb" "lb1" {
  cluster_uuid = intelcloud_iks_cluster.cluster1.id
  load_balancers = [
    {
      name     = "${local.name}-lb-pub2"
      port     = 80
      vip_type = "public"
    }
  ]
  depends_on = [intelcloud_iks_node_group.ng1]
}
