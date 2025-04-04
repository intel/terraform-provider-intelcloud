# Intel Tiber AI Cloud Terraform Provider 

This Intel Tiber AI Cloud provider plugin brings the power of Hashicorp's Terraform to Intel Tiber AI Cloud. It allows developers to model and manage their AI Cloud Resources through HCL IaaC (Infrastructure as a Code).

Useful links:
- [Intel Tiber AI Cloud Documentation](https://console.cloud.intel.com/docs/index.html)
- [Intel Tiber AI Cloud Provider Documentation](https://registry.terraform.io/providers/intel/intelcloud/latest/docs)
- [Terraform Documentation](https://www.terraform.io/docs/language/index.html)
<!-- [Terraform Provider Development](DEVELOPMENT.md) -->

The provider lets you declaratively define the configuration for your Intel Cloud platform.


## Contents

### Intel Tiber AI Cloud Provider for Terraform
  - [Requirements](#requirements)
  - [Using the Intel Tiber AI Cloud provider](#using-the-Intel-Tiber-AI-Cloud-provider)


### Requirements
-	[Terraform](https://www.terraform.io/downloads.html) 0.13 or higher
-	[Go](https://golang.org/doc/install) v1.21 or higher (to build the provider plugin)
- [Intel Cloud Platform](https://ai.cloud.intel.com/)

### Using the Intel Tiber AI Cloud provider

#### Intel Tiber AI Cloud Login Credentials
For creating resources on AI Cloud, it requires auth credentials. More specifically, currently it requires following `three` environment variables to be configured.

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
      version = "0.0.9"
    }
  }
}

provider "intelcloud" {
  # Configuration options
}
