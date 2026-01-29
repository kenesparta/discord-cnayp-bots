resource "aws_secretsmanager_secret" "discord_bot_token" {
  name                    = "${var.app_name}/discord-bot-token"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "discord_bot_token" {
  secret_id     = aws_secretsmanager_secret.discord_bot_token.id
  secret_string = var.discord_bot_token
}

resource "aws_secretsmanager_secret" "discord_guild_id" {
  name                    = "${var.app_name}/discord-guild-id"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "discord_guild_id" {
  secret_id     = aws_secretsmanager_secret.discord_guild_id.id
  secret_string = var.discord_guild_id
}
