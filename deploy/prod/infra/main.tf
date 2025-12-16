resource "google_service_account" "workctl" {
  account_id   = "workctl"
  display_name = "Workctl Service Account"
}

import {
  id = "projects/andrewhowdencom/serviceAccounts/workctl@andrewhowdencom.iam.gserviceaccount.com"
  to = google_service_account.workctl
}

resource "google_secret_manager_secret_iam_member" "workctl_config_accessor" {
  secret_id = "workctl-config"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.workctl.email}"
}

import {
  id = "projects/andrewhowdencom/secrets/workctl-config/roles/secretmanager.secretAccessor/serviceAccount:workctl@andrewhowdencom.iam.gserviceaccount.com"
  to = google_secret_manager_secret_iam_member.workctl_config_accessor
}

resource "google_dns_managed_zone" "workctl" {
  name        = "workctl-zone"
  dns_name    = "w.lahb.work."
  description = "DNS zone for workctl (w.lahb.work)"
}

import {
  id = "projects/andrewhowdencom/managedZones/workctl-zone"
  to = google_dns_managed_zone.workctl
}
