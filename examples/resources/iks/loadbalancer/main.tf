terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.18"
    }
  }
}


provider "intelcloud" {
  region = "us-region-2"
}

resource "intelcloud_iks_lb" "lb1" {
  cluster_uuid = "cl-vpymfpt3zu"
  load_balancers = [{
    name     = "tf-lb-a"
    port     = 80
    vip_type = "private"
    },
    {
      name     = "tf-lb-3"
      port     = 443
      vip_type = "public"
  }]
}

output "iks_order" {
  value = intelcloud_iks_lb.lb1
}
