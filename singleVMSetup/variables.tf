variable "resource_group_name" {
  description = "Name der Ressourcengruppe."
  type        = string
  default     = "eit-eh-polyport-dev"
}

variable "vnet_name" {
  description = "Name für das Virtual Network."
  type        = string
  default     = "polyport-vnet"
}

variable "vnet_address_space" {
  description = "Adressraum für das VNet."
  type        = list(string)
  default     = ["10.50.0.0/16"]
}

variable "vm_subnet_prefix" {
  description = "Adresspräfix für das VM-Subnetz."
  type        = string
  default     = "10.50.1.0/24"
}

variable "admin_username" {
  description = "Administrator-Benutzername für die VM."
  type        = string
  default     = "azureadmin"
}

variable "admin_public_key_path" {
  description = "Pfad zum öffentlichen SSH-Schlüssel."
  type        = string
  # Muss in terraform.tfvars oder als -var übergeben werden
}

variable "vm_name" {
  description = "Name für die All-in-One VM."
  type        = string
  default     = "mail-service-vm"
}

variable "allow_admin_ipv4_cidr" {
  description = "Ihre IPv4 CIDR für direkten SSH-Zugriff."
  type        = string
  default     = "0.0.0.0/0" # ÄNDERN Sie dies zu Ihrer IP für mehr Sicherheit!
}

variable "tags" {
  description = "Einheitliche Tags für alle Ressourcen."
  type        = map(string)
  default = {
    environment = "dev",
    project     = "polyport",
    created_by  = "terraform"
  }
}