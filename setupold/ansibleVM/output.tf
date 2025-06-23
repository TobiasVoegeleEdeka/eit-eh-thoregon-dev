output "public_ip_address" {
  value = azurerm_public_ip.pip.ip_address
}
output "vm_name" {
  value = azurerm_linux_virtual_machine.vm.name
}

output "public_ip_azure_fqdn" {
  description = "Der von Azure generierte FQDN für die öffentliche IP."
  value       = azurerm_public_ip.pip.fqdn
}