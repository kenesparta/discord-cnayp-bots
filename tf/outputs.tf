output "vm_name" {
  description = "Name of the VM instance"
  value       = google_compute_instance.vm.name
}

output "vm_external_ip" {
  description = "External IP address of the VM"
  value       = google_compute_instance.vm.network_interface[0].access_config[0].nat_ip
}

output "vm_internal_ip" {
  description = "Internal IP address of the VM"
  value       = google_compute_instance.vm.network_interface[0].network_ip
}

output "vm_zone" {
  description = "Zone where the VM is located"
  value       = google_compute_instance.vm.zone
}

output "service_account_email" {
  description = "Service account email (use this to share private calendars)"
  value       = google_service_account.bot.email
}
