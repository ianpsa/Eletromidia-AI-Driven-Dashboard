output "service_account_email_dev" {
    description = "Email equivalente a Service Account de Editor do BigQuery-Dev"
    value = google_service_account.bq_sa_dev.email
}

output "dataset_id_dev" {
    description = "Nome do Dataset do BigQuery-Dev"
    value = google_bigquery_dataset.eletromidia_bq_dev.dataset_id
}

output "table_ids" {
    description = "Id das tabelas criadas no BigQuery-Dev"
    value = { for k, v in google_bigquery_table.tables : k => v.table_id}
}