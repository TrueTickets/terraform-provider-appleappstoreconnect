# Copyright (c) HashiCorp, Inc.

variable "app_store_connect_issuer_id" {
  description = "The issuer ID from the API keys page in App Store Connect"
  type        = string
  sensitive   = true
}

variable "app_store_connect_key_id" {
  description = "The key ID from the API keys page in App Store Connect"
  type        = string
  sensitive   = true
}

variable "app_store_connect_private_key_path" {
  description = "Path to the private key (.p8) file for App Store Connect API authentication"
  type        = string
  default     = "~/.appstoreconnect/AuthKey.p8"
}
