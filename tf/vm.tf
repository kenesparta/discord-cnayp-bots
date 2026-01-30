resource "google_compute_instance" "vm" {
  name         = var.vm_name
  machine_type = "e2-micro"
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-stable"
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"

    access_config {
      # Ephemeral public IP
    }
  }

  # Attach service account for ADC (Application Default Credentials)
  service_account {
    email  = google_service_account.bot.email
    scopes = ["https://www.googleapis.com/auth/calendar.readonly"]
  }

  metadata = {
    ssh-keys = "gcpbot:ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKO9i8DJ+UEl356RStcDI89Gp0Ibn7b0z7ef1Bij/W4b gcpbot"
  }

  tags = ["discord-bot"]

  depends_on = [google_project_service.calendar_api]
}
