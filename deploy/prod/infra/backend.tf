terraform {
  backend "gcs" {
    bucket = "andrewhowdencom-infrastructure-state"
    prefix = "workctl-infra"
  }
}
