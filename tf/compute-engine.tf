# Static external IP
resource "google_compute_address" "bot" {
  name   = "${var.app_name}-ip"
  region = var.region
}

# Compute Engine instance with Container-Optimized OS
resource "google_compute_instance" "bot" {
  name         = var.app_name
  machine_type = "e2-micro"
  zone         = "${var.region}-b"

  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-stable"
      size  = 10 # GB, minimal for free tier
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.bot.address
    }
  }

  service_account {
    email  = google_service_account.bot.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    # Startup script to run the bot container
    startup-script = <<-EOF
      #!/bin/bash
      set -e

      # Container-Optimized OS has Docker pre-installed
      # Authenticate to Artifact Registry
      docker-credential-gcr configure-docker --registries=${var.region}-docker.pkg.dev

      # Fetch secrets from Secret Manager
      DISCORD_BOT_TOKEN=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/discord-bot-token" -H "Metadata-Flavor: Google" 2>/dev/null || \
        gcloud secrets versions access latest --secret="${var.app_name}-discord-token" 2>/dev/null || echo "")
      DISCORD_GUILD_ID=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/discord-guild-id" -H "Metadata-Flavor: Google" 2>/dev/null || \
        gcloud secrets versions access latest --secret="${var.app_name}-guild-id" 2>/dev/null || echo "")

      # Stop existing container if running
      docker stop ${var.app_name} 2>/dev/null || true
      docker rm ${var.app_name} 2>/dev/null || true

      # Pull latest image
      docker pull ${var.region}-docker.pkg.dev/${var.project_id}/${var.app_name}/${var.app_name}:latest

      # Run the bot
      docker run -d \
        --name ${var.app_name} \
        --restart unless-stopped \
        -e DISCORD_BOT_TOKEN="$DISCORD_BOT_TOKEN" \
        -e DISCORD_GUILD_ID="$DISCORD_GUILD_ID" \
        -e TZ=America/Lima \
        ${var.region}-docker.pkg.dev/${var.project_id}/${var.app_name}/${var.app_name}:latest
    EOF
  }

  tags = ["discord-bot"]

  # Allow the instance to be stopped/started for updates
  allow_stopping_for_update = true

  depends_on = [
    google_secret_manager_secret_iam_member.bot_token_access,
    google_secret_manager_secret_iam_member.guild_id_access,
    google_project_iam_member.bot_artifact_reader,
  ]
}

# Grant Artifact Registry read access to the service account
resource "google_project_iam_member" "bot_artifact_reader" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.bot.email}"
}

# Firewall rule for SSH (optional, for manual access)
resource "google_compute_firewall" "ssh" {
  name    = "${var.app_name}-allow-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"] # Restrict to your IP for better security
  target_tags   = ["discord-bot"]
}
