terraform {
  required_version = ">= 1.2"
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

# DNS Zone für DuckDNS Domain
resource "azurerm_dns_zone" "main" {
  name                = "edeka-inforservice.duckdns.org"
  resource_group_name = "eit-eh-polyport-dev"
  
  soa_record {
    email         = "hostmaster.edeka-inforservice.duckdns.org"
    expire_time   = 2419200
    minimum_ttl  = 300
    refresh_time  = 3600
    ttl           = 3600
  }
}

# A-Record für Postfix-Server
resource "azurerm_dns_a_record" "mail" {
  name                = "mail"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = "eit-eh-polyport-dev"
  ttl                 = 300
  records             = ["4.251.108.32"] #  bestehende IP
}

# MX Record
resource "azurerm_dns_mx_record" "primary" {
  name                = "@"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = "eit-eh-polyport-dev"
  ttl                 = 3600

  record {
    preference = 10  
    exchange   = "mail.edeka-inforservice.duckdns.org."
  }
}

# SPF Record
resource "azurerm_dns_txt_record" "spf" {
  name                = "@"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = "eit-eh-polyport-dev"
  ttl                 = 3600

  record {
    value = "v=spf1 ip4:4.251.108.32 -all"
  }
}

# Reverse-DNS für bestehende IP (manuelle Anpassung im Azure Portal nötig)
output "reverse_dns_instructions" {
  value = <<EOT
  Manuell im Azure Portal konfigurieren:
  1. Öffentliche IP-Ressource 'mail-service-vm' öffnen
  2. 'Konfiguration' → 'Reverse-DNS-Eintrag'
  3. Eintragen: mail.edeka-inforservice.duckdns.org
  4. Speichern
  EOT
}

output "azure_nameservers" {
  value = azurerm_dns_zone.main.name_servers
}