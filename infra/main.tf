provider "google" {
  project = var.project_id
  region  = var.region

  # Bill API calls made with user credentials to this project, rather than to
  # the project that the credentials were created in. Some APIs (e.g. billing
  # budgets) reject calls without a quota project.
  user_project_override = true
  billing_project       = var.project_id
}

locals {
  services = [
    "artifactregistry.googleapis.com", # Container images
    "billingbudgets.googleapis.com",   # Budget alerts
    "iam.googleapis.com",              # Service accounts, workload identity pools
    "iamcredentials.googleapis.com",   # Short-lived credentials for GitHub Actions
    "run.googleapis.com",              # Cloud Run
    "sts.googleapis.com",              # Token exchange for workload identity federation
  ]
}

resource "google_project_service" "services" {
  for_each = toset(local.services)

  service = each.value

  # Keep the APIs enabled if this resource is removed, since other resources
  # may still depend on them.
  disable_on_destroy = false
}
