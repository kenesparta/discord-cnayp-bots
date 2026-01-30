resource "google_service_account" "bot" {
  account_id   = "cnayp-bot"
  display_name = "CNAYP Discord Bot"
  description  = "Service account for Discord bot to access Google Calendar API"
}

# Enable Google Calendar API
resource "google_project_service" "calendar_api" {
  service            = "calendar-json.googleapis.com"
  disable_on_destroy = false
}
