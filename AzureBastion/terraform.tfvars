# Beispiel terraform.tfvars - ERSETZE DIESE WERTE!

resource_group_name = "eit-eh-polyport-dev" # Name deiner existierenden RG
location            = "WestEurope"          # Ort deiner existierenden RG

vnet_name           = "thoregon-vnet"         # Name für das neue VNet
vnet_address_space  = ["10.40.0.0/16"]      # Adressraum für das neue VNet (Beispiel!)

vm_subnet_prefix    = "10.40.1.0/24"        # Präfix für VM Subnetz (muss in vnet_address_space liegen)
bastion_subnet_prefix = "10.40.2.0/26"      # Präfix für Bastion Subnetz (muss in vnet_address_space liegen, mind /26)

# vm_subnet_name kann hier optional überschrieben werden
# vm_subnet_name = "mein-vm-subnetz"