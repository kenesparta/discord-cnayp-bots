resource "google_secret_manager_secret" "discord_bot_token" {
  secret_id = "${var.app_name}-discord-token"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "discord_bot_token" {
  secret      = google_secret_manager_secret.discord_bot_token.id
  secret_data = var.discord_bot_token
}

resource "google_secret_manager_secret" "discord_guild_id" {
  secret_id = "${var.app_name}-guild-id"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "discord_guild_id" {
  secret      = google_secret_manager_secret.discord_guild_id.id
  secret_data = var.discord_guild_id
}
