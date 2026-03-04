resource "google_project_iam_member" "bq_dev_data_editor" {
  project = var.project_id
  role    = "roles/bigquery.dataEditor"
  member  = "serviceAccount:${var.service_account_bq_dev_email}"
}

resource "google_storage_bucket_iam_member" "cs_dev_sa" {
  bucket = var.cs_dev_name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${var.service_account_cs_dev_email}"
}