terraform {
  required_providers {
    idc = {
      source = "cloud.intel.com/services/idc"
    }
  }
}

provider "idc" {
   region = "staging-1"
}

data "idc_kubernetes_clusters" "default" {}

output "test_cluster" {
	value = data.idc_kubernetes_clusters.default
}
