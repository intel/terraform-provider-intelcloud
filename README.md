# ITAC Terraform Provider 

This ITAC provider plugin brings the power of Hashicorp's Terraform to Intel Tiber AI Cloud (ITAC). It allows developers to model and manage their ITAC Resources through HCL IaaC (Infrastructure as a Code).

Useful links:
- [ITAC Documentation](https://staging.console.idcservice.net/docs/index.html)
- [ITAC Provider Documentation](https://registry.terraform.io/providers/intel/intelcloud/latest/docs)
- [Terraform Documentation](https://www.terraform.io/docs/language/index.html)
<!-- [Terraform Provider Development](DEVELOPMENT.md) -->

The provider lets you declaratively define the configuration for your Intel Cloud platform.


## Contents

### ITAC Provider for Terraform
  - [Requirements](#requirements)
  - [Using the ITAC provider](#using-the-itac-provider)


### Requirements
-	[Terraform](https://www.terraform.io/downloads.html) 0.13 or higher
-	[Go](https://golang.org/doc/install) v1.21 or higher (to build the provider plugin)
- [Intel Cloud Platform](https://ai.cloud.intel.com/)

### Using the ITAC provider

#### ITAC Login Credentials
For creating resources on ITAC, it requires auth credentials. More specifically, currently it requires following `three` environment variables to be configured.

The values for these environment variables can be created by the user using creds as a service.

```
export ITAC_CLOUDACCOUNT=<cloudaccount>
export ITAC_CLIENT_ID=<Client ID >
export ITAC_CLIENT_SECRET=<Client secret>
```


To quickly get started using the Intel provider for Terraform, configure the provider as shown below. Full provider documentation with details on all options available is located on the [Terraform Registry site](https://registry.terraform.io/providers/intel/intelcloud/latest/docs).

```hcl
terraform {
  required_providers {
    intelcloud = {
      source = "intel/intelcloud"
      version = "0.0.8"
    }
  }
}

provider "intelcloud" {
  # Configuration options
}
