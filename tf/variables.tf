variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_profile" {
  description = "AWS CLI profile name"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "discord-cnayp-bot"
}

variable "github_repo" {
  description = "GitHub repository in format owner/repo"
  type        = string
}

variable "discord_bot_token" {
  description = "Discord bot token"
  type        = string
  sensitive   = true
}

variable "discord_guild_id" {
  description = "Discord guild ID"
  type        = string
  sensitive   = true
}
