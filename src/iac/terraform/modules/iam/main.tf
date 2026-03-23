resource "google_project_iam_member" "bq_admin" { // Cria identidade de acesso `bq_admin`
  project = var.project_id
  role    = "roles/bigquery.admin"
  member  = "user:${var.admin_email}"
}

resource "google_project_iam_member" "bq_data_editor" { // Cria identidade de acesso `bq_editor` para os serviços
  project = var.project_id
  role    = "roles/bigquery.dataEditor"
  member  = "serviceAccount:${var.service_account_email}"
}

resource "google_service_account" "bq_looker_reader" {
  project      = var.project_id
  account_id   = "bq-looker-reader"
  display_name = "BigQuery Looker Studio Reader SA"
}

resource "google_project_iam_member" "bq_looker_reader" {
  project = var.project_id
  role    = "roles/bigquery.dataViewer"
  member  = "serviceAccount:${google_service_account.bq_looker_reader.email}"
}

resource "google_project_iam_member" "log_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${var.gke_sa_email}"
}

resource "google_project_iam_member" "metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${var.gke_sa_email}"
}

resource "google_project_iam_member" "artifact_reader" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${var.gke_sa_email}"
}

resource "google_project_iam_member" "storage_viewer" {
  project = var.project_id
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${var.gke_sa_email}"
}

resource "google_project_iam_member" "bigquery_user" {
  project = var.project_id
  role    = "roles/bigquery.dataViewer"
  member  = "serviceAccount:${var.gke_sa_email}"
}

