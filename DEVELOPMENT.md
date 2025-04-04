# Intel Tiber AI Cloud Terraform Provider

This Intel Tiber AI Cloud provider plugin brings the power of Hashicorp's Terraform to Intel Tiber AI Cloud. It allows developers to model and manage their AI Cloud Resources through HCL IaaC (Infrastructure as a Code).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 0.13
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
```shell
git clone https://github.com/intel/terraform-provider-intelcloud
```

2. Enter the repository directory
```shell
cd terraform-provider-intelcloud
```

3. Build the provider using the Go `install` command:
```shell
go install
```

## Intel Tiber AI Cloud Login Credentials
For creating resources on AI Cloud, it requires auth credentials. More specifically, currently it requires following `three` environment variables to be configured.

```
export ITAC_CLOUDACCOUNT=<cloudaccount>
export ITAC_CLIENT_ID=<Client ID >
export ITAC_CLIENT_SECRET=<Client secret>
```


## Terraform configuration

To quickly get started using the Intel Tiber AI Cloud provider for Terraform, configure the provider as shown below. Full provider documentation with details on all options available is located on the [Terraform Registry site](https://registry.terraform.io/providers/intel/intelcloud/latest/docs).

```hcl
terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.9"
    }
  }
}

provider "intelcloud" {
  # Configuration options
}