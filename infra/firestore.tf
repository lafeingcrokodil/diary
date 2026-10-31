# The app's data. "(default)" is the database that client libraries use unless
# told otherwise.
resource "google_firestore_database" "default" {
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  # The database holds data that can't be recreated, so neither Terraform nor
  # anyone else should be able to delete it by accident.
  delete_protection_state = "DELETE_PROTECTION_ENABLED"
  deletion_policy         = "ABANDON"

  depends_on = [google_project_service.services["firestore.googleapis.com"]]
}

# Lets the app read and write the database.
resource "google_project_iam_member" "app_runtime_firestore" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = google_service_account.app_runtime.member
}
