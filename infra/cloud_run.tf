# The identity that the app runs as. It only gets the roles that the app needs,
# which so far are none.
resource "google_service_account" "app_runtime" {
  account_id  = "app-runtime"
  description = "Runs the Cloud Run service."

  depends_on = [google_project_service.services["iam.googleapis.com"]]
}

resource "google_cloud_run_v2_service" "app" {
  name     = "app"
  location = var.region

  # The service holds no data and can be recreated from this configuration and
  # an image, so deletion protection would only get in the way of changes that
  # replace it (e.g. moving it to another region).
  deletion_protection = false

  # Anyone can reach the service without authenticating with Google Cloud, so
  # any access control is up to the app.
  ingress              = "INGRESS_TRAFFIC_ALL"
  invoker_iam_disabled = true

  template {
    service_account = google_service_account.app_runtime.email

    # Scale to zero when idle, and cap the number of instances, which also
    # caps the cost.
    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      # A placeholder until the first deployment: deployments update the
      # image outside of Terraform (see lifecycle below).
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      resources {
        limits = {
          cpu    = "1"
          memory = "256Mi"
        }
        # Only allocate CPU while handling requests (request-based billing).
        cpu_idle = true
      }

      startup_probe {
        http_get {
          path = "/health"
        }
      }
    }
  }

  lifecycle {
    # Deployments set the image (and the client that deployed it), so
    # Terraform shouldn't revert those.
    ignore_changes = [
      client,
      client_version,
      template[0].containers[0].image,
    ]
  }

  depends_on = [google_project_service.services["run.googleapis.com"]]
}
