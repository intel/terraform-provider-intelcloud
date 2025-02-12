## Pre-reqs for running through this example:

- [ITAC terraform setup](../../DEVELOPMENT.md)

## Usage examples

1. Creating a bucket in IDC
```shell
cd examples/resources/buckets/
terraform init
terraform apply
```

2. Deleting a bucket created in the above step
```shell
cd examples/resources/buckets/
terraform destroy
```