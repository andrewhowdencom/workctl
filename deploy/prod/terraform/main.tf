resource "google_service_account" "workctl" {
  account_id   = "workctl"
  display_name = "Workctl Service Account"
}

resource "google_secret_manager_secret_iam_member" "workctl_config_accessor" {
  secret_id = "workctl-config"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.workctl.email}"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "workctl"
  location = "europe-west10"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.workctl.email

    containers {
      image = "europe-west10-docker.pkg.dev/andrewhowdencom/workctl/workctl:${var.image_tag}"
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
}

resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_v2_service.default.location
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_dns_managed_zone" "workctl" {
  name        = "workctl-zone"
  dns_name    = "w.lahb.work."
  description = "DNS zone for workctl (w.lahb.work)"
}

resource "google_cloud_run_domain_mapping" "default" {
  location = "europe-west10"
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
