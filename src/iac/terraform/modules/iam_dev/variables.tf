variable "project_id" {
  description = "Id do Projeto utilizado"
  type        = string
}

variable "service_account_bq_dev_email" {
  description = "Email da Service Account do BigQuery-Dev"
  type        = string
}

variable "cs_dev_name" {
  description = "Nome do Cloud Storage-Dev"
  type        = string
}

variable "service_account_cs_dev_email" {
  description = "Service Account do Cloud Storage-Dev"
  type        = string
}