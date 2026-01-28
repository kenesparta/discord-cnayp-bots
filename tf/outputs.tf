output "instance_external_ip" {
  description = "Compute Engine instance external IP"
  value       = google_compute_address.bot.address
}

output "instance_name" {
  description = "Compute Engine instance name"
  value       = google_compute_instance.bot.name
}

output "instance_zone" {
  description = "Compute Engine instance zone"
  value       = google_compute_instance.bot.zone
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.bot.repository_id}"
}

output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.bot.email
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "gcloud compute ssh ${google_compute_instance.bot.name} --zone=${google_compute_instance.bot.zone} --project=${var.project_id}"
}

output "deploy_command" {
  description = "Command to deploy/update the bot"
  value       = "gcloud compute ssh ${google_compute_instance.bot.name} --zone=${google_compute_instance.bot.zone} --project=${var.project_id} --command='sudo systemctl restart konlet-startup'"
}
