output "image_repository" {
  description = "Where to push the app's container images."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.main.repository_id}"
}

output "service_url" {
  description = "URL of the Cloud Run service."
  value       = google_cloud_run_v2_service.app.uri
}

output "workload_identity_provider" {
  description = "Workload identity provider for GitHub Actions (the GCP_WORKLOAD_IDENTITY_PROVIDER repository variable)."
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "deployer_service_account" {
  description = "Service account that GitHub Actions deploys as (the GCP_DEPLOYER_SERVICE_ACCOUNT repository variable)."
  value       = google_service_account.app_deployer.email
}

output "custom_domain_dns_updates" {
  description = "DNS records to add for the custom domain, so that Firebase Hosting can verify it and serve it."
  value       = google_firebase_hosting_custom_domain.default.required_dns_updates
}

output "firebase_web_config" {
  description = "Firebase configuration for the front end. None of these values are secret: they identify the project, and are visible to anyone who loads the app."
  value = {
    apiKey     = data.google_firebase_web_app_config.default.api_key
    authDomain = var.domain
    projectId  = var.project_id
    appId      = google_firebase_web_app.default.app_id
  }
}
