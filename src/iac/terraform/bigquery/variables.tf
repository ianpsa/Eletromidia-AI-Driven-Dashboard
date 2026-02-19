variable "project_id" {
    description = "ID do Projeto do GCP"
    type = string
}

variable "location" {
    description = "Localização do BigQuery"
    type = string
    default = "US"
}

variable "dataset_id" {
    description = "Nome do Dataset no BigQuery"
    type = string
}

variable "table_id" {
    description = "Nome da tabela do BigQuery"
    type = string
}   

variable "admin_email" {
    description = "Email do usuário que terá admin no BigQuery"
    type = string
}