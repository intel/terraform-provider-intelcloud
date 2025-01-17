# ITAC Terraform Provider (Terraform Plugin Framework)

This ITAC provider plugin brings the power of Hashicorp's Terraform to Intel Tiber AI Cloud (ITAC). It allows developers to model and manage their ITAC Resources through HCL IaaC (Infrastructure as a Code).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 0.13
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
```shell
git clone https://github.com/intel/terraform-provider-intel-cloud
```

2. Enter the repository directory
```shell
cd terraform-provider-intel-cloud
```

3. Build the provider using the Go `install` command:
```shell
go install
```


## Terraform configuration

To quickly get started using the ITAC provider for Terraform, configure the provider as shown below. Full provider documentation with details on all options available is located on the [Terraform Registry site](https://registry.terraform.io/providers/intel/intel-cloud/latest/docs).

```hcl
terraform {
  required_providers {
    intel-cloud = {
      source = "intel/intel-cloud"
      version = "0.0.2"
    }
  }
}

provider "intel-cloud" {
  # Configuration options
}