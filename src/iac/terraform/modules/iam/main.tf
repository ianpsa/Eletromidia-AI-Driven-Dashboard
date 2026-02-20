resource "google_project_iam_member" "bq_admin" {       // Cria identidade de acesso `bq_admin`
    project         = var.project_id
    role            = "roles/bigquery.admin"
    member          =  "user:${var.admin_email}"
}

resource "google_project_iam_member" "bq_data_editor" {      // Cria identidade de acesso `bq_editor` para os serviços
    project         = var.project_id
    role            = "roles/bigquery.dataEditor"
    member          = "serviceAccount:${var.service_account_email}"
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