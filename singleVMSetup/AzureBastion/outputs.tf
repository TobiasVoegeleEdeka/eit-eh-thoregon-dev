output "vnet_id" {
  description = "Die ID des erstellten VNets."
  value       = azurerm_virtual_network.vnet.id
}

output "vm_subnet_id" {
  description = "Die ID des erstellten VM-Subnetzes."
  value       = azurerm_subnet.vm_subnet.id
}

output "bastion_subnet_id" {
  description = "Die ID des erstellten AzureBastionSubnet."
  value       = azurerm_subnet.bastion_subnet.id
}