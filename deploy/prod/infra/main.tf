resource "google_service_account" "workctl" {
  account_id   = "workctl"
  display_name = "Workctl Service Account"
}


resource "google_secret_manager_secret_iam_member" "workctl_config_accessor" {
  secret_id = "workctl-config"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.workctl.email}"
}


resource "google_dns_managed_zone" "workctl" {
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
    "roles/secretmanager.admin",
    "roles/dns.admin",
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

# Allow GitHub Actions to authenticate via Workload Identity Federation
resource "google_service_account_iam_member" "gh_actions_wif_user" {
  service_account_id = google_service_account.workctl_gh_actions.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/projects/422614898574/locations/global/workloadIdentityPools/github/attribute.repository/andrewhowdencom/workctl"
}

# --- Application Resources Consolidation ---

resource "google_cloud_run_v2_service" "default" {
  name     = "workctl"
  location = "europe-west1"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.workctl.email

    containers {
      image = "europe-west1-docker.pkg.dev/andrewhowdencom/workctl/workctl:${var.image_tag}"
      ports {
        container_port = 8080
      }
      volume_mounts {
        name       = "config-volume"
        mount_path = "/etc/workctl"
      }
    }
    volumes {
      name = "config-volume"
      secret {
        secret = "workctl-config"
        items {
          version = "latest"
          path    = "config.yaml"
        }
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }
}




resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_v2_service.default.location
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}




resource "google_cloud_run_domain_mapping" "default" {
  location = "europe-west1"
  name     = "w.lahb.work"

  metadata {
    namespace = "andrewhowdencom"
  }

  spec {
    route_name = "workctl"
  }
}




resource "google_dns_record_set" "cname" {
  name         = "w.lahb.work."
  type         = "CNAME"
  ttl          = 300
  managed_zone = google_dns_managed_zone.workctl.name
  rrdatas      = [for r in google_cloud_run_domain_mapping.default.status[0].resource_records : r.rrdata if r.type == "CNAME"]
}


