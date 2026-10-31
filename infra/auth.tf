# Firebase Authentication, which is based on Identity Platform. Google sign-in
# itself is turned on in the Firebase console (see README.md), because it needs
# an OAuth client, which Terraform can't create.
#
# This configuration can only be created once per project, and can't be
# deleted.
resource "google_identity_platform_config" "default" {
  # Domains from which the sign-in flow can be started.
  authorized_domains = [
    var.domain,
    "${var.project_id}.web.app",
    "${var.project_id}.firebaseapp.com",
    "localhost",
  ]

  depends_on = [
    google_firebase_project.default,
    google_project_service.services["identitytoolkit.googleapis.com"],
  ]
}

# The front end's registration with Firebase, which gives it the configuration
# it needs to use Firebase Authentication.
resource "google_firebase_web_app" "default" {
  provider = google-beta

  display_name = "Diary"

  depends_on = [google_firebase_project.default]
}

data "google_firebase_web_app_config" "default" {
  provider = google-beta

  web_app_id = google_firebase_web_app.default.app_id
}
