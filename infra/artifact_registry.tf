resource "google_artifact_registry_repository" "main" {
  repository_id = "main"
  location      = var.region
  format        = "DOCKER"
  description   = "Container images."

  # Delete image versions older than 30 days, but always keep the five most
  # recent versions of each image, so that storage stays within the free tier
  # while recent versions remain available for rollbacks. Keep policies take
  # precedence over delete policies.
  cleanup_policies {
    id     = "delete-old"
    action = "DELETE"
    condition {
      tag_state  = "ANY"
      older_than = "${30 * 24 * 60 * 60}s"
    }
  }

  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions {
      keep_count = 5
    }
  }

  # Scanning is charged per image.
  vulnerability_scanning_config {
    enablement_config = "DISABLED"
  }

  depends_on = [google_project_service.services["artifactregistry.googleapis.com"]]
}
