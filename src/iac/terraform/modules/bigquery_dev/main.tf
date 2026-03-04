resource "google_bigquery_dataset" "eletromidia_bq_dev" {
  project    = var.project_id
  dataset_id = var.dataset_id_dev
  location   = var.location
}

resource "google_bigquery_table" "tables" {
  for_each = var.tables

  project    = var.project_id
  dataset_id = google_bigquery_dataset.eletromidia_bq_dev.dataset_id
  table_id   = each.value.table_id
  schema     = file("${path.module}/${each.value.schema}")

  deletion_protection = false
}

resource "google_service_account" "bq_sa_dev" {
  project      = var.project_id
  account_id   = "bq-dev-data-editor"
  display_name = "BigQuery Development Enviroment Data Editor Service Account"
}