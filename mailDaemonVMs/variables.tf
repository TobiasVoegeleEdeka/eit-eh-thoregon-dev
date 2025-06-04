# --- GRUPPE & STANDORT ---
variable "resource_group_name" {
  description = "Name der BEREITS EXISTIERENDEN Ressourcengruppe."
  type        = string
  default     = "eit-eh-polyport-dev"
}

variable "location" {
  description = "Azure Region der Ressourcengruppe (wird automatisch aus der RG gelesen, dieser Wert hier ist ein Fallback)."
  type        = string
  default     = "West Europe"
}

# --- NETZWERK ---
variable "vnet_name" {
  description = "Name für das NEU zu erstellende Virtual Network (VNet)."
  type        = string
  default     = "polyport-vnet"
}

variable "vnet_address_space" {
  description = "Adressraum für das VNet (z.B. [\"10.50.0.0/16\"]). Muss in terraform.tfvars angegeben werden."
  type        = list(string)
}

variable "vm_subnet_prefix" {
  description = "Adresspräfix für das VM-Subnetz (z.B. \"10.50.1.0/24\"). Muss in terraform.tfvars angegeben werden."
  type        = string
}

variable "bastion_subnet_prefix" {
  description = "Adresspräfix für das AzureBastionSubnet (z.B. \"10.50.0.0/26\"). Mindestens /26 empfohlen. Muss in terraform.tfvars angegeben werden."
  type        = string
}


# --- VMs ---
variable "admin_username" {
  description = "Administrator-Benutzername für die VMs."
  type        = string
  default     = "azureadmin"
}

variable "admin_public_key_path" {
  description = "Pfad zum öffentlichen SSH-Schlüssel für den Admin-Benutzer (z.B. \"~/.ssh/id_rsa.pub\"). Muss in terraform.tfvars angegeben werden."
  type        = string
}

variable "api_vm_name" {
  description = "Name für die Postfix API VM."
  type        = string
  default     = "postfix-api-vm"
}

variable "db_vm_name" {
  description = "Name für die Postgres DB VM."
  type        = string
  default     = "postgres-db-vm"
}

variable "postfix_vm_name" {
  description = "Name für die Postfix Mailserver VM."
  type        = string
  default     = "postfix-mail-vm"
}

variable "vm_size" {
  description = "Größe der VMs (z.B. 'Standard_B2s')."
  type        = string
  default     = "Standard_B1s"
}

# --- VARIABLEN FÜR SSH-ZUGRIFF UND TAGS ---

variable "allow_admin_ipv4_cidr" {
  description = "Administrative IPv4 CIDR für direkten SSH-Zugriff."
  type        = string
  default     = "91.21.30.141/32"
}

variable "allow_ansible_vm_ipv4_cidr" {
  description = "Oeffentliche IPv4 CIDR der Ansible Control VM für automatisierten SSH-Zugriff."
  type        = string
  default     = "4.233.100.91/32"
}
variable "allow_ssh_ipv6_cidr" {
  description = "IPv6 CIDR-Block für erlaubten SSH-Zugriff. Beispiel: Ihre öffentliche IPv6 /128."
  type        = string
  default     = "2001:4860:7:1410::f4/128"
}

variable "tags" {
  description = "Einheitliche Tags für alle Ressourcen."
  type        = map(string)
  default = {
    environment = "dev"
    project     = "polyport"
    created_by  = "terraform"
  }
}