output "image_repository" {
  description = "Where to push the app's container images."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.main.repository_id}"
}

output "service_url" {
  description = "URL of the Cloud Run service."
  value       = google_cloud_run_v2_service.app.uri
}
