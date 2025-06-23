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

# --- Datenquelle zum Nachschlagen der existierenden Ressourcengruppe ---
data "azurerm_resource_group" "rg" {
  # Verwendet den Namen aus der Variable (oder dem Default-Wert)
  name = var.resource_group_name
}

# --- NEU: Virtuelles Netzwerk (VNet) wird jetzt erstellt ---
resource "azurerm_virtual_network" "vnet" {
  name                = var.vnet_name           # Name aus Variable
  address_space       = var.vnet_address_space  # Adressraum aus Variable
  location            = data.azurerm_resource_group.rg.location # Ort von der RG übernehmen
  resource_group_name = data.azurerm_resource_group.rg.name   # Name der existierenden RG
}

# --- Subnetz für normale VMs (Referenziert jetzt die NEUE VNet Ressource) ---
resource "azurerm_subnet" "vm_subnet" {
  name                 = var.vm_subnet_name
  resource_group_name  = data.azurerm_resource_group.rg.name
  # Wichtig: Referenziert jetzt die neu erstellte VNet-Ressource
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.vm_subnet_prefix]
}

# --- Dediziertes Subnetz für Azure Bastion (Referenziert jetzt die NEUE VNet Ressource) ---
resource "azurerm_subnet" "bastion_subnet" {
  name                 = "AzureBastionSubnet" # Fester Name!
  resource_group_name  = data.azurerm_resource_group.rg.name

  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.bastion_subnet_prefix]
}

# --- HIER FÜGST DU SPÄTER DEN CODE FÜR BASTION PIP, BASTION HOST, VM NIC, VM NSG, VM SELBST EIN ---
# Z.B.:
# resource "azurerm_public_ip" "thoregon_bastion_pip" { ... location = data.azurerm_resource_group.rg.location ... resource_group_name = data.azurerm_resource_group.rg.name ... }
# resource "azurerm_bastion_host" "thoregon_bastion" { ... subnet_id = azurerm_subnet.bastion_subnet.id ... public_ip_address_id = azurerm_public_ip.thoregon_bastion_pip.id ... }
# resource "azurerm_network_interface" "vm_nic" { ... subnet_id = azurerm_subnet.vm_subnet.id ... } # Keine Public IP für VM mehr
# resource "azurerm_network_security_group" "vm_nsg" { ... } # Ohne eingehende SSH Regel von außen
# resource "azurerm_network_interface_security_group_association" "vm_nic_nsg_assoc" { ... }
# resource "azurerm_linux_virtual_machine" "vm" { ... network_interface_ids = [azurerm_network_interface.vm_nic.id] ... }