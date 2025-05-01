terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.15"
    }
  }
}


provider "intelcloud" {
  region = "us-staging-1"
}

data "intelcloud_kubeconfig" "kubeconfig" {
  cluster_uuid = "cl-e4nimo7leq"
}

resource "local_file" "kubeconfig_output" {
  filename = "${path.module}/kubeconfig-cl.yml"
  content  = data.intelcloud_kubeconfig.kubeconfig.kubeconfig
}
