variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-east1"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "discord-cnayp-bot"
}

variable "discord_bot_token" {
  description = "Discord bot token"
  type        = string
  sensitive   = true
}

variable "discord_guild_id" {
  description = "Discord guild (server) ID"
  type        = string
}

variable "ssh_public_key" {
  description = "SSH public key for instance access (format: username:key)"
  type        = string
}
