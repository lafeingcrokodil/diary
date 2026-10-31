data "google_project" "project" {
  # The project number is looked up through this API.
  depends_on = [google_project_service.services["cloudresourcemanager.googleapis.com"]]
}

resource "google_billing_budget" "monthly" {
  billing_account = var.billing_account
  display_name    = "${var.project_id} monthly budget"

  budget_filter {
    projects = ["projects/${data.google_project.project.number}"]
  }

  amount {
    # Without a currency code, the amount is in the billing account's currency.
    specified_amount {
      units = tostring(var.monthly_budget)
    }
  }

  # Email the billing account's administrators and users as actual spending
  # passes each threshold, and when the forecast for the month exceeds the
  # budget.
  threshold_rules {
    threshold_percent = 0.5
  }

  threshold_rules {
    threshold_percent = 0.9
  }

  threshold_rules {
    threshold_percent = 1.0
  }

  threshold_rules {
    threshold_percent = 1.0
    spend_basis       = "FORECASTED_SPEND"
  }

  depends_on = [google_project_service.services["billingbudgets.googleapis.com"]]
}
