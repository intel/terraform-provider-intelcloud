# Intel Tiber AI Cloud Terraform Provider 
### This example is to demonstrate to the users, how an IKS clusters with multiple nodegroups can be deployed (with nodegroup details being configurable)

### Pre-requisites
- Please see this page for getting setup correctly before running this example [Setting up provider locally](../../../../README.md)

## Explanation of the example [main.tf](main.tf)

### Creating the resource
```
locals {
  ...
  ...
  # list of nodegroups user wants to create
  node_group_list = {
    "ng1" = {
      node_type  = "vm-spr-sml"
      node_count = 1
      ssh_keys   = ["rk-win-key", "rk-tf-key"]
    }
    "ng2" = {
      node_type  = "vm-spr-sml"
      node_count = 1
      ssh_keys   = ["rk-win-key"]
    }
  }
  ...
  ...
}
```

The above block is used to define a local variable for represnting a list of nodegroups, in the main.tf file, this will be looped through and values will picked up for each nodegroup, namely `ng1` and `ng2` in this case.

When the user runs `terraform apply` in the current directory, the user should see the following:

```
Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # intelcloud_iks_cluster.cluster will be created
  + resource "intelcloud_iks_cluster" "cluster" {
      + cloudaccount       = (known after apply)
      + cluster_status     = (known after apply)
      + id                 = (known after apply)
      + kubernetes_version = "1.30"
      + name               = (known after apply)
      + network            = (known after apply)
      + storage            = {
          + size_in_tb       = 30
          + state            = (known after apply)
          + storage_provider = (known after apply)
        }
      + upgrade_available  = (known after apply)

      + timeouts {
          + resource_timeout = "30m"
        }
    }

  # intelcloud_iks_node_group.node_group["ng1"] will be created
  + resource "intelcloud_iks_node_group" "node_group" {
      + cluster_uuid         = (known after apply)
      + id                   = (known after apply)
      + imiid                = (known after apply)
      + name                 = "ng1"
      + node_count           = 1
      + node_type            = "vm-spr-sml"
      + ssh_public_key_names = [
          + "rk-tf-key",
          + "rk-win-key",
        ]
      + state                = (known after apply)
      + userdata_url         = (known after apply)
      + vnets                = (known after apply)
    }

  # intelcloud_iks_node_group.node_group["ng2"] will be created
  + resource "intelcloud_iks_node_group" "node_group" {
      + cluster_uuid         = (known after apply)
      + id                   = (known after apply)
      + imiid                = (known after apply)
      + name                 = "ng2"
      + node_count           = 1
      + node_type            = "vm-spr-sml"
      + ssh_public_key_names = [
          + "rk-win-key",
        ]
      + state                = (known after apply)
      + userdata_url         = (known after apply)
      + vnets                = (known after apply)
    }

  # random_pet.prefix will be created
  + resource "random_pet" "prefix" {
      + id        = (known after apply)
      + length    = 2
      + separator = "-"
    }

Plan: 4 to add, 0 to change, 0 to destroy.
```

As you can see, each nodegroup is treated as a separate resource to be created.

### Modifying the resource
- Assume `terraform apply` is executed by the user and completes successfully.
- The user can change the nodegroup count and run `terraform apply` to modify the nodegroup
```
locals {
  ...
  ...
  # list of nodegroups user wants to create
  node_group_list = {
    "ng1" = {
      node_type  = "vm-spr-sml"
      node_count = 2
      ssh_keys   = ["rk-win-key", "rk-tf-key"]
    }
    "ng2" = {
      node_type  = "vm-spr-sml"
      node_count = 3
      ssh_keys   = ["rk-win-key"]
    }
  }
  ...
  ...
}
```
- After `terraform apply` the node count in nodegroups `ng1` and `ng2` will updated to 2 and 3 respectively 

### Modification with deletion included
- If after that user edits the local variable to have just one nodegroup in the list, for example
```
locals {
  ...
  ...
  # list of nodegroups user wants to create
  node_group_list = {
    "ng1" = {
      node_type  = "vm-spr-sml"
      node_count = 1
      ssh_keys   = ["rk-win-key", "rk-tf-key"]
    }
  ...
  ...
}
```

- If after this the user runs `terraform apply`  then nodegroup `ng2` will be deleted as the provider will assume it as such


### Destroying the resource
- The resource can be destroyed (once created successfully) by simply running `terraform destroy`
