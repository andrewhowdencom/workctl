resource "google_artifact_registry_repository" "default" {
  location      = "europe-west10"
  repository_id = "workctl"
  description   = "Docker repository for workctl"
  format        = "DOCKER"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "workctl"
  location = "europe-west10"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = "europe-west10-docker.pkg.dev/andrewhowdencom/workctl/workctl:${var.image_tag}"
      ports {
        container_port = 8080
      }
    }
  }

  depends_on = [google_artifact_registry_repository.default]
}

resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_v2_service.default.location
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
