resource "google_cloud_run_v2_service" "bot" {
  name     = var.app_name
  location = var.region

  template {
    service_account = google_service_account.bot.email

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.bot.repository_id}/${var.app_name}:latest"

      env {
        name = "DISCORD_BOT_TOKEN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.discord_bot_token.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "DISCORD_GUILD_ID"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.discord_guild_id.secret_id
            version = "latest"
          }
        }
      }

      env {
        name  = "TZ"
        value = "America/Lima"
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "256Mi"
        }
      }
    }

    scaling {
      min_instance_count = 1
      max_instance_count = 1
    }
  }

  depends_on = [
    google_secret_manager_secret_iam_member.bot_token_access,
    google_secret_manager_secret_iam_member.guild_id_access,
  ]
}
