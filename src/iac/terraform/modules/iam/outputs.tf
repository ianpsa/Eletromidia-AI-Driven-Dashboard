output "bq_looker_reader_email" {
    description = "Email da Service Account para a leitura do Looker"
    value = google_service_account.bq_looker_reader.email
}