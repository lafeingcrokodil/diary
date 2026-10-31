variable "project_id" {
  description = "ID of the GCP project that hosts the app."
  type        = string
  default     = "lafeingcrokodil-diary"
}

variable "region" {
  description = "Region for the app's resources."
  type        = string
  default     = "europe-west10"
}
