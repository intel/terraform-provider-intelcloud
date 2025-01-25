# ITAC Terraform Provider

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

## ITAC Login Credentials
For creating resources on ITAC, it requires auth credentials. More specifically, currently it requires following `three` environment variables to be configured.

```
export ITAC_CLOUDACCOUNT=<cloudaccount>
export ITAC_CLIENT_ID=<Client ID >
export ITAC_CLIENT_SECRET=<Client secret>
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

## Usage examples

1. Creating a filesystem in IDC
```shell
cd examples/resources/filesystems/
terraform init
terraform apply
```

2. Creating a bucket in IDC
```shell
cd examples/resources/buckets/
terraform init
terraform apply
```

3. Creating an instance in IDC
```shell
cd examples/resources/buckets/
terraform init
terraform apply
```

4. Deploying  gen-ai chatqna
```shell
cd examples/gen-ai-xeon-opea-chatqna
```
Change the `cloud_init.yaml` file to add your hugging face token (you can create an account on  [Hugging face](https://huggingface.co/settings/tokens) and get a token.
```shell
  - echo 'export HUGGINGFACEHUB_API_TOKEN=<HUGGINGFACE_API_TOKEN>' | sudo tee -a /etc/profile.d/opea.sh
```
Change the `terraform.tfvars` file to add your `ssh_keyname` , `ssh_pubkey_path` and `ssh_user_email`
```shell
idc_region            = "us-region-2"
ssh_key_name          = "<SSH_KEY>"
instance_name         = "genai-chatqna-demo3"
ssh_pubkey_path       = "<PATH_TO_SSH_PUBKEY>"
ssh_user_email        = "<USER_EMAIL>"
instance_type         = "vm-large"
filesystem_name       = "shri-fs6"
filesystem_size_in_tb = 1
```
```shell
terraform init
terraform apply
```
Verify that instance has been deployed. Follow IDC instructions to ssh into the instance.
Edit the `/etc/profile.d/opea.sh` file and change the below values and save the file.
```shell
#!/bin/bash
export public_ip=<INSTANCE_IP>
export host_ip=>INSTANCE_IP
...
...
```
Source the file
```shell
source /etc/profile.d/opea.sh
```

Re-deploy the gen-ai docker containers
```shell
cd /opt/GenAIExamples/CodeGen/docker_compose/intel/cpu/xeon
docker compose down
docker compose up
```

Run the following command for port-forwarding from the instance to your local machine. Run this on your local machine
```shell
ssh -J guest@<JUMP_VM_IP_IDC>  -L 5174:localhost:5174 -L 7778:localhost:7778 -L 9000:localhost:9000 -L 8028:localhost:8028 -L 6379:localhost:6379 -L 8001:localhost:8001 -L 6006:localhost:6006 -L 6000:localhost:6000 -L 7000:localhost:7000 -L 8399:localhost:8399 -L 9399:localhost:9399 -L 6007:localhost:6007 -L 8888:localhost:8888  ubuntu@<INSTANCE_IP>
```

Go to your local browser and run the following:
```
localhost:5174
```
The browser should open the code-gen UI.