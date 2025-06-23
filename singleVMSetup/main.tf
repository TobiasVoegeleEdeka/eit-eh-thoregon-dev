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

# --- NETZWERK-RESSOURCEN (ALS DATENQUELLE LESEN) ---
data "azurerm_virtual_network" "vnet" {
  name                = var.vnet_name
  resource_group_name = data.azurerm_resource_group.rg.name
}

data "azurerm_subnet" "vm_subnet" {
  name                 = "vm-subnet"
  resource_group_name  = data.azurerm_resource_group.rg.name
  virtual_network_name = data.azurerm_virtual_network.vnet.name
}

# --- FIREWALL (NSG) ---
resource "azurerm_network_security_group" "app_vm_nsg" {
  name                = "${var.vm_name}-nsg"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags

  security_rule {
    name                       = "AllowSSH"
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

  security_rule {
    name                       = "AllowSubmission" # Für API/Clients zum Einliefern
    priority                   = 301
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "587"
    source_address_prefix      = "Internet" # Muss für externe Clients offen sein
    destination_address_prefix = "*"
  }
}

# --- RESSOURCEN FÜR DIE EINE VM ---
resource "azurerm_public_ip" "vm_pip" {
  name                = "${var.vm_name}-pip"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
  domain_name_label   = lower(var.vm_name)
}

resource "azurerm_network_interface" "vm_nic" {
  name                = "${var.vm_name}-nic"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = data.azurerm_subnet.vm_subnet.id # Greift jetzt auf die Datenquelle zu
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_pip.id
  }
}

resource "azurerm_network_interface_security_group_association" "vm_nic_nsg_assoc" {
  network_interface_id      = azurerm_network_interface.vm_nic.id
  network_security_group_id = azurerm_network_security_group.app_vm_nsg.id
}

resource "azurerm_linux_virtual_machine" "all_in_one_vm" {
  name                  = var.vm_name
  resource_group_name   = data.azurerm_resource_group.rg.name
  location              = data.azurerm_resource_group.rg.location
  size                  = "Standard_B2s" # Etwas größer für mehrere Container
  admin_username        = var.admin_username
  network_interface_ids = [azurerm_network_interface.vm_nic.id]
  tags                  = var.tags

  custom_data = base64encode(file("${path.module}/setup.yaml"))

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
output "vm_public_ip" {
  description = "Public IP address of the main VM."
  value       = azurerm_public_ip.vm_pip.ip_address
}

output "vm_fqdn" {
  description = "Fully Qualified Domain Name of the main VM."
  value       = azurerm_public_ip.vm_pip.fqdn
}