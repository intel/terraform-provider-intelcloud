terraform {
  required_providers {
    intelcloud = {
      source  = "intel/intelcloud"
      version = "0.0.19"
    }
  }
}


provider "intelcloud" {
  region = "us-staging-1"
  endpoints = {
    api  = "https://us-staging-1-sdk-api.eglb.intel.com"
    auth = "https://client-token.staging.api.idcservice.net"
  }
}


resource "intelcloud_iks_lb" "lb1" {
  cluster_uuid = "cl-fw6jqd3bpe"
  load_balancers {
    name   = "rk-tf-lb"
    schema = "public"
    listeners {
      port     = "81"
      protocol = "LBProtocolTCP"

      pool {
        port                = "81"
        monitor             = "https"
        load_balancing_mode = "roundRobin"
        node_group_id       = "ng-pngrk2griy"
      }
      security {
        source_ips = ["any"]
      }
    }

    security {
      source_ips = ["any"]
    }
  }
  timeouts {
    resource_timeout = "30m"
  }
}

output "iks_order" {
  value = intelcloud_iks_lb.lb1
}
