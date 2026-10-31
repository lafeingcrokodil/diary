variable "project_id" {
  description = "ID of the GCP project that hosts the app."
  type        = string
  default     = "lafeingcrokodil-diary"
}

variable "region" {
  description = "Region for the app's resources."
  type        = string
  default     = "europe-west9"
}

variable "billing_account" {
  description = "ID of the billing account that pays for the project, e.g. 012345-6789AB-CDEF01."
  type        = string
}

variable "monthly_budget" {
  description = "Monthly budget in the billing account's currency. Alerts are sent as spending approaches it."
  type        = number
  default     = 5

  validation {
    condition     = var.monthly_budget > 0 && floor(var.monthly_budget) == var.monthly_budget
    error_message = "The monthly budget must be a positive whole number."
  }
}

variable "allowed_emails" {
  description = "Email addresses of the Google accounts that are allowed to use the app."
  type        = list(string)

  validation {
    condition     = length(var.allowed_emails) > 0 && alltrue([for email in var.allowed_emails : !strcontains(email, ",")])
    error_message = "At least one email address is required, and the addresses can't contain commas."
  }
}

variable "domain" {
  description = "Custom domain that the app is served on."
  type        = string
  default     = "diary.lafeingcrokodil.com"
}
