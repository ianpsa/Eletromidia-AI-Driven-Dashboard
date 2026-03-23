variable "project_id" {
  description = "Id do Projeto do GCP"
  type        = string
}

variable "admin_email" {
  description = "Email do usuário admin do BigQuery"
  type        = string
}

variable "service_account_email" {
  description = "Email da Service Account criada pelo BigQuery"
  type        = string
}

variable "gke_sa_email" {
  type = string
  description = "Email da Service Account do GKE"
}