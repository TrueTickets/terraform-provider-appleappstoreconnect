# Copyright (c) HashiCorp, Inc.

variable "app_store_connect_issuer_id" {
  type        = string
  sensitive   = true
  description = "The issuer ID from the API keys page in App Store Connect"
}

variable "app_store_connect_key_id" {
  type        = string
  sensitive   = true
  description = "The key ID from the API keys page in App Store Connect"
}

variable "app_store_connect_private_key" {
  type        = string
  sensitive   = true
  description = "The private key for App Store Connect API authentication"
}
