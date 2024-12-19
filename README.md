# IDC Terraform Provider (Terraform Plugin Framework)

This IDC provider plugin brings the power of Hashicorp's Terraform to Intel Developer Cloud (IDC). It allows developers to model and manage their IDC Resources through HCL IaaC (Infrastructure as a Code).

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

## Trying out the IDC Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

For local develoment, update the terraform config to point it to the local copy of the provider plugin.

Edit `~/.terraformrc` file and add following config block

```
provider_installation {

  dev_overrides {
      "cloud.intel.com/services/idc" = "<$GOPATH>/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## IDC Login Credentials
For creating resources on IDC, it requires auth credentials. More specifically, currently it requires following `two` environment variables to be configured.

```
export IDC_CLOUDACCOUNT=<cloudaccount>
export IDC_APITOKEN=<JWT Token >
```

You can optionally, download and setup the following CLI tool to fetch it automatically.

[IRR Binary Download](https://github.com/intel-innersource/applications.web.saas.optimization-registry.api/releases/tag/v0.23.5)

```
response=$(irr_darwin idc login --interactive --json)
IDC_CLOUDACCOUNT=$(echo $response | jq -r ".account_id")
IDC_APITOKEN=$(echo $response | jq -r ".tokens.access_token")
```

## Next Steps
