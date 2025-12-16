resource "google_service_account" "workctl" {
  account_id   = "workctl"
  display_name = "Workctl Service Account"
}


RESOURCE "google_secret_manager_secret_iam_member" "workctl_config_accessor" {
  secret_id = "workctl-config"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.workctl.email}"
}


RESOURCE "google_dns_managed_zone" "workctl" {
  name        = "workctl-zone"
  dns_name    = "w.lahb.work."
  description = "DNS zone for workctl (w.lahb.work)"
}


# Terraform Executor Service Account
resource "google_service_account" "workctl_gh_actions" {
  account_id   = "workctl-gh-actions"
  display_name = "Workctl GitHub Actions Executor"
  description  = "Service Account for GitHub Actions to execute Terraform"
}

# Project-level permissions
resource "google_project_iam_member" "workctl_gh_actions_project_roles" {
  for_each = toset([
    "roles/run.admin",
    "roles/storage.admin",
  ])
  project = "andrewhowdencom"
  role    = each.key
  member  = "serviceAccount:${google_service_account.workctl_gh_actions.email}"
}

# Allow impersonating/using the workctl Service Account
resource "google_service_account_iam_member" "workctl_gh_actions_impersonate_workctl" {
  service_account_id = google_service_account.workctl.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.workctl_gh_actions.email}"
}

# Allow managing DNS records in the workctl zone
resource "google_dns_managed_zone_iam_member" "workctl_gh_actions_dns_admin" {
  managed_zone = google_dns_managed_zone.workctl.name
  role         = "roles/dns.admin"
  member       = "serviceAccount:${google_service_account.workctl_gh_actions.email}"
}

# Allow CI Runner to impersonate this Service Account
resource "google_service_account_iam_member" "ci_runner_impersonate_workctl_gh_actions" {
  service_account_id = google_service_account.workctl_gh_actions.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = "serviceAccount:github-actions-runner@andrewhowdencom.iam.gserviceaccount.com"
}
