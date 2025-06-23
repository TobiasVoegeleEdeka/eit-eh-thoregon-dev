# terraform.tfvars

vnet_address_space    = ["10.50.0.0/16"]
vm_subnet_prefix      = "10.50.1.0/24"
bastion_subnet_prefix = "10.50.0.0/26"
admin_public_key_path = "~/.ssh/id_rsa.pub"