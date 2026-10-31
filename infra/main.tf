provider "google" {
  project = var.project_id
  region  = var.region

  # Bill API calls made with user credentials to this project, rather than to
  # the project that the credentials were created in. Some APIs (e.g. billing
  # budgets) reject calls without a quota project.
  user_project_override = true
  billing_project       = var.project_id
}

provider "google-beta" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
  billing_project       = var.project_id
}

locals {
  services = [
    "artifactregistry.googleapis.com", # Container images
    "billingbudgets.googleapis.com",   # Budget alerts
    "firebase.googleapis.com",         # Firebase
    "firebasehosting.googleapis.com",  # Firebase Hosting
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

# Used to look up the project number. It's separate from the other APIs, so
# that the lookup only waits for this API: depends_on applies to all instances
# of a resource, so depending on the list above would defer the lookup to the
# apply phase whenever an API is added to it.
resource "google_project_service" "cloudresourcemanager" {
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}
