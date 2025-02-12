## Pre-reqs for running through this example:

- [ITAC terraform setup](../../DEVELOPMENT.md)

## Usage examples

1. Creating a filesystem in IDC
```shell
cd examples/resources/filesystems/
terraform init
terraform apply
```

2. Deleting a filesystem created in the above step
```shell
cd examples/resources/filesystems/
terraform destroy
```

3. Modifying a filesystem created in step 1
```shell
cd examples/resources/filesystems/
Edit the main.tf file to edit any values, for eg. modify 'size_in_tb = 5'
terraform apply
```