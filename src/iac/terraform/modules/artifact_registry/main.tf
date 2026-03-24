resource "google_artifact_registry_repository" "gke-cr" {
  provider      = google
  project       = var.project_id
  location      = var.region
  repository_id = "gke-cr"
  format        = "DOCKER"
  description   = "Docker registry para o GKE"
}