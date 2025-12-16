data "google_service_account" "workctl" {
  account_id = "workctl"
}

data "google_dns_managed_zone" "workctl" {
  name = "workctl-zone"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "workctl"
  location = "europe-west1"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = data.google_service_account.workctl.email

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
  managed_zone = data.google_dns_managed_zone.workctl.name
  rrdatas      = [for r in google_cloud_run_domain_mapping.default.status[0].resource_records : r.rrdata if r.type == "CNAME"]
}


