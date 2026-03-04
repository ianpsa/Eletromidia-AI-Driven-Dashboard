
resource "google_bigquery_dataset" "eletromidia_bq" { // Cria o dataset no BigQuery
  project    = var.project_id
  dataset_id = var.dataset_id
  location   = var.location
}

resource "google_bigquery_table" "tables" {
  for_each = var.tables

  project    = var.project_id
  dataset_id = google_bigquery_dataset.eletromidia_bq.dataset_id
  table_id   = each.value.table_id
  schema     = file("${path.module}/${each.value.schema}")

  deletion_protection = false
}

resource "google_service_account" "bq_sa" {
  project      = var.project_id
  account_id   = "bq-data-editor"
  display_name = "BigQuery Data Editor Service Account"
}