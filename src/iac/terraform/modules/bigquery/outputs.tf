output "service_account_email" {
    description = "Email da Service Account criada para o BigQuery"
    value = google_service_account.bq_sa.email
}

output "dataset_id" {
    description = "Nome do Dataset do BigQuery"
    value = google_bigquery_dataset.eletromidia_bq.dataset_id
}

output "table_ids" {
    description = "Nome da Tabela do BigQuery"
    value = { for k, v in google_bigquery_table.tables : k => v.table_id }
}