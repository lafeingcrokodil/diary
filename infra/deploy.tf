# Lets GitHub Actions deploy the app without a long-lived key: workflow runs
# exchange GitHub's OIDC token for short-lived credentials of the deployer
# service account, through workload identity federation.

locals {
  # The repository's numeric ID doesn't change if the repository is renamed,
  # and can't be taken over by a new repository with the same name.
  github_repository_id = "1440716135"
}

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github"
  description               = "GitHub Actions workflows."

  depends_on = [google_project_service.services["iam.googleapis.com"]]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_mapping = {
    "google.subject"          = "assertion.sub"
    "attribute.repository_id" = "assertion.repository_id"
  }

  # Only accept tokens from workflow runs for the main branch of this
  # repository. Pull requests (including those from forks) get no access.
  attribute_condition = "assertion.repository_id == '${local.github_repository_id}' && assertion.ref == 'refs/heads/main'"
}

resource "google_service_account" "app_deployer" {
  account_id  = "app-deployer"
  description = "Deploys the app from GitHub Actions."

  depends_on = [google_project_service.services["iam.googleapis.com"]]
}

# Workflow runs accepted by the provider can act as the deployer.
resource "google_service_account_iam_member" "app_deployer_github" {
  service_account_id = google_service_account.app_deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository_id/${local.github_repository_id}"
}

# The deployer can push images to the repository...
resource "google_artifact_registry_repository_iam_member" "app_deployer" {
  location   = google_artifact_registry_repository.main.location
  repository = google_artifact_registry_repository.main.name
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.app_deployer.member
}

# ...deploy new revisions of the service...
resource "google_cloud_run_v2_service_iam_member" "app_deployer" {
  location = google_cloud_run_v2_service.app.location
  name     = google_cloud_run_v2_service.app.name
  role     = "roles/run.developer"
  member   = google_service_account.app_deployer.member
}

# ...and run them as the runtime service account.
resource "google_service_account_iam_member" "app_deployer_runtime" {
  service_account_id = google_service_account.app_runtime.name
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.app_deployer.member
}
