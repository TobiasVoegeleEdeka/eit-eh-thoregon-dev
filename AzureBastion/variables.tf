variable "resource_group_name" {
  description = "Name der BEREITS EXISTIERENDEN Ressourcengruppe."
  type        = string
  default     = "eit-eh-polyport-dev"
}

# Entfernt: existing_vnet_name, da wir das VNet jetzt erstellen

variable "location" {
  description = "Azure Region der existierenden Ressourcengruppe."
  type        = string
  default     = "WestEurope" 
}

variable "vnet_name" {
  description = "Name für das NEU zu erstellende Virtual Network (VNet)."
  type        = string
  default     = "thoregon-vnet" 
}

variable "vnet_address_space" {
  description = "Adressraum für das NEU zu erstellende VNet (z.B. [\"10.40.0.0/16\"])."
  type        = list(string)
  # Kein Default-Wert, muss z.B. in terraform.tfvars angegeben werden
}

variable "vm_subnet_name" {
  description = "Name für das zu erstellende Subnetz für normale VMs."
  type        = string
  default     = "vm-subnet"
}

variable "vm_subnet_prefix" {
  description = "Adresspräfix für das VM-Subnetz (muss im VNet-Adressraum liegen)."
  type        = string
  # Kein Default-Wert, muss angegeben werden (z.B. in terraform.tfvars)
}

variable "bastion_subnet_prefix" {
  description = "Adresspräfix für das AzureBastionSubnet (muss im VNet liegen, mind. /26 empfohlen)."
  type        = string
  # Kein Default-Wert, muss angegeben werden (z.B. in terraform.tfvars)
}