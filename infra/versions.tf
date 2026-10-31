terraform {
  required_version = ">= 1.17"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "8.7.0"
    }
  }

  # The state bucket can't be managed by the configuration that stores its
  # state in it, so it's created by hand (see README.md).
  backend "gcs" {
    bucket = "lafeingcrokodil-diary-tfstate"
  }
}
