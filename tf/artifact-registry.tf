resource "google_artifact_registry_repository" "bot" {
  location      = var.region
  repository_id = var.app_name
  description   = "Docker repository for Discord CNCF bot"
  format        = "DOCKER"

  cleanup_policies {
    id     = "keep-latest"
    action = "KEEP"

    most_recent_versions {
      keep_count = 3
    }
  }
}
