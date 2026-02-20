output "bucket_name" {
  description = "Nome do bucket criado"
  value       = google_storage_bucket.kafka_backup.name
}

output "service_account_email" {
  description = "Email da Service Account do Cloud Storage"
  value       = google_service_account.cs_sa.email
}