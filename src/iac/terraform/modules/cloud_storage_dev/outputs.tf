output "service_account_email" {
    description = "Email referente a Service Account do Cloud Storage-Dev"
    value = google_service_account.cs_sa_dev.email
}

output "bucket_name_dev" {
    description = "Nome do Cloud Storage criado"
    value = google_storage_bucket.kafka_backup_dev.name
}