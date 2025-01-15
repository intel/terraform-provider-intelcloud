# ITAC Terraform Provider (Terraform Plugin Framework)

This ITAC provider plugin brings the power of Hashicorp's Terraform to Intel Tiber AI Cloud (ITAC). It allows developers to model and manage their ITAC Resources through HCL IaaC (Infrastructure as a Code).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

Currently, this plugin is not published to terraform registry and is available to be used in Local Dev mode locally. 

## Trying out the ITAC Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

For local develoment, update the terraform config to point it to the local copy of the provider plugin.

Edit `~/.terraformrc` file and add following config block

```
provider_installation {

  dev_overrides {
      "cloud.intel.com/services/itac" = "<$GOPATH>/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## ITAC Login Credentials
For creating resources on ITAC, it requires auth credentials. More specifically, currently it requires following `three` environment variables to be configured.

```
export ITAC_CLOUDACCOUNT=<cloudaccount>
export ITAC_CLIENT_ID=<Client ID >
export ITAC_CLIENT_SECRET=<Client secret>
```

## Next Steps


