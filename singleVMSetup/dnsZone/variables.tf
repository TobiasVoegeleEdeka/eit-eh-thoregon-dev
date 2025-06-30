variable "resource_group_name" {
  description = "Name der Azure Resource Group"
  type        = string
  default     = "eit-eh-polyport-dev"
}

variable "location" {
  description = "Azure-Region für Ressourcen"
  type        = string
  default     = "francecentral"  # Angepasst an Ihre tatsächliche Region
}

variable "domain_name" {
  description = "Hauptdomain (ohne Subdomain)"
  type        = string
  default     = "edeka-inforservice.duckdns.org"
}

variable "mail_subdomain" {
  description = "Subdomain für den Mailserver"
  type        = string
  default     = "mail"
}

variable "public_ip_name" {
  description = "Name der öffentlichen IP-Ressource"
  type        = string
  default     = "mail-service-vm-pip"
}

variable "mx_preference" {
  description = "Priorität des MX-Records"
  type        = number
  default     = 10
}

variable "dns_ttl" {
  description = "Standard-TTL für DNS-Einträge"
  type        = number
  default     = 3600
}

# SPF Policy ohne dynamische IP (wird in main.tf ergänzt)
variable "spf_policy_template" {
  description = "SPF-Richtlinie ohne IP"
  type        = string
  default     = "v=spf1 ip4:%s -all"
}