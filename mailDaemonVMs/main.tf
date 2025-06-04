terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# --- DATENQUELLEN ---
data "azurerm_resource_group" "rg" {
  name = var.resource_group_name
}

# --- NETZWERK-RESSOURCEN ---
resource "azurerm_virtual_network" "vnet" {
  name                = var.vnet_name
  address_space       = var.vnet_address_space
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags
}

resource "azurerm_subnet" "vm_subnet" {
  name                 = "vm-subnet"
  resource_group_name  = data.azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.vm_subnet_prefix]
}

resource "azurerm_subnet" "bastion_subnet" {
  name                 = "AzureBastionSubnet"
  resource_group_name  = data.azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.bastion_subnet_prefix]
}

resource "azurerm_network_security_group" "vm_nsg" {
  name                = "vm-application-nsg"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags
  

  security_rule {
    name                       = "AllowSSHFromAdmin"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = var.allow_admin_ipv4_cidr
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowSSHFromAnsibleVM"
    priority                   = 101
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = var.allow_ansible_vm_ipv4_cidr
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowSSH_IPv6"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = var.allow_ssh_ipv6_cidr
    destination_address_prefix = "*"
  }

  

  # Neue Regel: Erlaubt eingehenden Mail-Verkehr von anderen Mail-Servern
  security_rule {
    name                       = "AllowSMTPInbound"
    priority                   = 300
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "25"
    source_address_prefix      = "Internet"
    destination_address_prefix = "*"
  }

  # Neue Regel: Erlaubt Ihrer API, E-Mails zum Versand einzuliefern
  security_rule {
    name                       = "AllowSubmissionFromVnet"
    priority                   = 301
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "587"
    source_address_prefix      = "VirtualNetwork" # Zugriff nur aus Ihrem VNet
    destination_address_prefix = "*"
  }
# Regel fuer Lets Encrypt ( offene , automatisierte Zertifizierungstelle) 
  security_rule {
    name                       = "AllowLetsEncryptHTTPChallenge"
    priority                   = 150 
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "Internet"
    destination_address_prefix = "*"
  }

}

# --- RESSOURCEN FÜR API-VM ---

resource "azurerm_public_ip" "api_vm_pip" {
  name                = "${var.api_vm_name}-pip"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_network_interface" "api_vm_nic" {
  name                = "${var.api_vm_name}-nic"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.vm_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.api_vm_pip.id
  }
}

resource "azurerm_network_interface_security_group_association" "api_nic_nsg_assoc" {
  network_interface_id      = azurerm_network_interface.api_vm_nic.id
  network_security_group_id = azurerm_network_security_group.vm_nsg.id
}

resource "azurerm_linux_virtual_machine" "api_vm" {
  name                  = var.api_vm_name
  resource_group_name   = data.azurerm_resource_group.rg.name
  location              = data.azurerm_resource_group.rg.location
  size                  = var.vm_size
  admin_username        = var.admin_username
  network_interface_ids = [azurerm_network_interface.api_vm_nic.id]
  tags                  = var.tags

  admin_ssh_key {
    username   = var.admin_username
    public_key = file(var.admin_public_key_path)
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }
}

# --- RESSOURCEN FÜR DB-VM ---

resource "azurerm_public_ip" "db_vm_pip" {
  name                = "${var.db_vm_name}-pip"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_network_interface" "db_vm_nic" {
  name                = "${var.db_vm_name}-nic"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.vm_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.db_vm_pip.id
  }
}

resource "azurerm_network_interface_security_group_association" "db_nic_nsg_assoc" {
  network_interface_id      = azurerm_network_interface.db_vm_nic.id
  network_security_group_id = azurerm_network_security_group.vm_nsg.id
}

resource "azurerm_linux_virtual_machine" "db_vm" {
  name                  = var.db_vm_name
  resource_group_name   = data.azurerm_resource_group.rg.name
  location              = data.azurerm_resource_group.rg.location
  size                  = var.vm_size
  admin_username        = var.admin_username
  network_interface_ids = [azurerm_network_interface.db_vm_nic.id]
  tags                  = var.tags

  admin_ssh_key {
    username   = var.admin_username
    public_key = file(var.admin_public_key_path)
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }
}

# --- RESSOURCEN FÜR POSTFIX-VM ---

resource "azurerm_public_ip" "postfix_vm_pip" {
  name                = "${var.postfix_vm_name}-pip"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_network_interface" "postfix_vm_nic" {
  name                = "${var.postfix_vm_name}-nic"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.vm_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.postfix_vm_pip.id
  }
}

resource "azurerm_network_interface_security_group_association" "postfix_nic_nsg_assoc" {
  network_interface_id      = azurerm_network_interface.postfix_vm_nic.id
  network_security_group_id = azurerm_network_security_group.vm_nsg.id
}

resource "azurerm_linux_virtual_machine" "postfix_vm" {
  name                  = var.postfix_vm_name
  resource_group_name   = data.azurerm_resource_group.rg.name
  location              = data.azurerm_resource_group.rg.location
  size                  = var.vm_size
  admin_username        = var.admin_username
  network_interface_ids = [azurerm_network_interface.postfix_vm_nic.id]
  tags                  = var.tags

  admin_ssh_key {
    username   = var.admin_username
    public_key = file(var.admin_public_key_path)
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }
}

# --- AUSGABEN ---
output "api_vm_public_ip" {
  description = "Public IP address of the API VM."
  value       = azurerm_public_ip.api_vm_pip.ip_address
}

output "api_vm_private_ip" {
  description = "Private IP address of the API VM."
  value       = azurerm_network_interface.api_vm_nic.private_ip_address
}

output "db_vm_public_ip" {
  description = "Public IP address of the DB VM."
  value       = azurerm_public_ip.db_vm_pip.ip_address
}

output "db_vm_private_ip" {
  description = "Private IP address of the DB VM."
  value       = azurerm_network_interface.db_vm_nic.private_ip_address
}

output "postfix_vm_public_ip" {
  description = "Public IP address of the Postfix VM."
  value       = azurerm_public_ip.postfix_vm_pip.ip_address
}

output "postfix_vm_private_ip" {
  description = "Private IP address of the Postfix VM."
  value       = azurerm_network_interface.postfix_vm_nic.private_ip_address
}