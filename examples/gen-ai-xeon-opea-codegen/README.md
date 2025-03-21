## Pre-reqs for running through this example:

- [ITAC terraform setup](../../DEVELOPMENT.md)

## Usage examples

1. Deploying  gen-ai codegen
```shell
cd examples/gen-ai-xeon-opea-codegen
```
Change the `cloud_init.yaml` file to add your hugging face token (you can create an account on  [Hugging face](https://huggingface.co/settings/tokens) and get a token.
```shell
  - echo 'export HUGGINGFACEHUB_API_TOKEN=<HUGGINGFACE_API_TOKEN>' | sudo tee -a /etc/profile.d/opea.sh
```
Change the `terraform.tfvars` file to add your `ssh_keyname` , `ssh_pubkey_path` 
```shell
idc_region            = "us-region-2"
ssh_key_name          = "<SSH_KEY>"
instance_name         = "genai-chatqna-demo3"
ssh_pubkey_path       = "<PATH_TO_SSH_PUBKEY>"
instance_type         = "vm-large"
filesystem_name       = "shri-fs6"
filesystem_size_in_tb = 1
```
```shell
terraform init
terraform apply
```
Verify that instance has been deployed. Follow IDC instructions to ssh into the instance.

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