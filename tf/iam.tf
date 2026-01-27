resource "google_service_account" "bot" {
  account_id   = var.app_name
  display_name = "Discord CNCF Bot Service Account"
}

resource "google_secret_manager_secret_iam_member" "bot_token_access" {
  secret_id = google_secret_manager_secret.discord_bot_token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.bot.email}"
}

resource "google_secret_manager_secret_iam_member" "guild_id_access" {
  secret_id = google_secret_manager_secret.discord_guild_id.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.bot.email}"
}
