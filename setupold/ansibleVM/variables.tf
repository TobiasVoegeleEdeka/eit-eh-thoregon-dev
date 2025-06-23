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

variable "additional_inbound_nsg_rules" {
  description = "Eine Liste von zusätzlichen eingehenden Sicherheitsregeln für die Network Security Group (z.B. für SMTP, HTTP, etc.). SSH-Regeln werden separat über ihre eigenen Variablen gehandhabt."
  type = list(object({
    name                       = string
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
  default = [
    # SMTP-Regel für Port 25
    {
      name                       = "AllowSMTPInbound"
      priority                   = 130 
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "25"
      source_address_prefix      = "Internet" 
      destination_address_prefix = "*"
    },
    # # SMTP Submission-Regel für Port 587
    # {
    #   name                       = "AllowSMTPOutnbound"
    #   priority                   = 100 
    #   direction                  = "Outbound"
    #   access                     = "Allow"
    #   protocol                   = "Tcp"
    #   source_port_range          = "*"
    #   destination_port_range     = "587"
    #   source_address_prefix      = "Internet" 
    #   destination_address_prefix = "*"
    # }
   
  ]
}