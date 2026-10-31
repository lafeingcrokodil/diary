# Adds Firebase to the project, so that Firebase services can be used with it.
resource "google_firebase_project" "default" {
  provider = google-beta

  depends_on = [google_project_service.services["firebase.googleapis.com"]]
}

# Firebase Hosting serves the app on the custom domain, by forwarding every
# request to the Cloud Run service. (Cloud Run itself can only serve custom
# domains through a paid load balancer, or through domain mapping, which
# isn't available in every region.)
resource "google_firebase_hosting_site" "default" {
  provider = google-beta

  site_id = var.project_id

  depends_on = [
    google_firebase_project.default,
    google_project_service.services["firebasehosting.googleapis.com"],
  ]
}

resource "google_firebase_hosting_version" "default" {
  provider = google-beta

  site_id = google_firebase_hosting_site.default.site_id

  config {
    rewrites {
      glob = "**"
      run {
        service_id = google_cloud_run_v2_service.app.name
        region     = google_cloud_run_v2_service.app.location
      }
    }
  }
}

resource "google_firebase_hosting_release" "default" {
  provider = google-beta

  site_id      = google_firebase_hosting_site.default.site_id
  version_name = google_firebase_hosting_version.default.name
  message      = "Forward all requests to Cloud Run"
}

resource "google_firebase_hosting_custom_domain" "default" {
  provider = google-beta

  site_id       = google_firebase_hosting_site.default.site_id
  custom_domain = var.domain

  # The DNS records are added by hand (see README.md), so don't wait for them.
  wait_dns_verification = false
}
