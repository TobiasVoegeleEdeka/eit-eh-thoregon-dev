variable "location" {
  description = "Azure Region"
  type        = string
  default     = "WestEurope" 
}

variable "resource_group_name" {
  description = " Azure Ressourcengruppe"
  type        = string
  default     = "eit-eh-polyport-dev"
}

variable "existing_vnet_name" {
  description = "virtuelles Netzwerk (VNet)"
  type        = string
  default     = "thoregon-vnet"
}

variable "existing_subnet_name" {
  description = "subnet"
  type        = string
  default     = "vm-subnet" 
}

variable "admin_username" {
  description = "Admin-Benutzername für die VM"
  type        = string
  default     = "ansiblemin"
}

variable "ssh_public_key_path" {
  description = "SSH Pfad"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "vm_name" {
  description = "VM Name"
  type        = string
  default     = "AnsibleControlCenter"
}

variable "vm_size" {
  description = "Größe der VM"
  type        = string
  default     = "Standard_B1s"
}


variable "allow_ssh_ipv4_cidr" {
  description = "IPv4 CIDR-Block für erlaubten SSH-Zugriff."
  type        = string
  default     = "91.21.30.141/32"
}

variable "allow_ssh_ipv6_cidr" {
  description = "IPv6 CIDR-Block für erlaubten SSH-Zugriff."
  type        = string
  default     = "2001:4860:7:1410::f4/128"
}