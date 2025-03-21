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

locals {
  name = "apollo11"
  tags = {
    environment = "Demo"
  }
}

data "idc_machine_images" "image" {
  most_recent = true
  filters = [
    {
      name   = "name"
      values = ["ubuntu-2204-jammy"]
    }
  ]
}

resource "idc_sshkey" "sshkey-1" {
  metadata = {
    name = "${local.name}-sshkey"
  }
  spec = {
    ssh_public_key = file(var.ssh_pubkey_path)
  }
}

resource "idc_instance" "myinstance-1" {
  async             = false
  name              = "${local.name}-instance"
  availability_zone = var.idc_availability_zone
  spec = {
    instance_type = var.instance_types[var.instance_type]
    machine_image = data.idc_machine_images.image.result.name
    interface_specs = [{
      name = var.instance_interface_spec.name
      vnet = var.instance_interface_spec.vnet
    }]
    ssh_public_key_names = [idc_sshkey.sshkey-1.metadata.name]
  }
  depends_on = [idc_sshkey.sshkey-1]
}

resource "idc_filesystem" "fsvol-1" {
  name              = "${local.name}-filevol"
  description       = var.filesystem_description
  availability_zone = var.idc_availability_zone
  spec = {
    size_in_gb      = var.filesystem_size_in_gb
    filesystem_type = var.filesystem_type
  }
}

resource "idc_object_storage_bucket" "bucket1" {
  name      = "${local.name}-bucket"
  versioned = false
}

resource "idc_object_storage_bucket_user" "user1" {
  name      = "${idc_object_storage_bucket.bucket1.name}-user"
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

resource "idc_iks_cluster" "cluster1" {
  async              = false
  name               = "${local.name}-iks"
  availability_zone  = var.idc_availability_zone
  kubernetes_version = "1.27"

  storage = {
    size_in_gb = 30
  }
}

resource "idc_iks_node_group" "ng1" {
  cluster_uuid         = idc_iks_cluster.cluster1.id
  name                 = "${local.name}-ng"
  node_count           = 1
  node_type            = var.instance_types[var.instance_type]
  userdata_url         = ""
  ssh_public_key_names = [idc_sshkey.sshkey-1.metadata.name]
  interfaces = [{
    name = var.idc_availability_zone
    vnet = var.instance_interface_spec.vnet
  }]
}

resource "idc_iks_lb" "lb1" {
  cluster_uuid = idc_iks_cluster.cluster1.id
  load_balancers = [
    {
      name     = "${local.name}-lb-pub2"
      port     = 80
      vip_type = "public"
    }
  ]
  depends_on = [idc_iks_node_group.ng1]
}
