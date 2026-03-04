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

variable "dataset_id_dev" {
    description = "Nome do Dataset no BigQuery"
    type = string
}

variable "tables" {
    description = "Nomes das tabelas do BigQuery"
    type = map(object({
        table_id = string
        schema = string
    }))
}   

variable "admin_email" {
    description = "Email do usuário que terá admin no BigQuery"
    type = string
}

variable "cs_name" {
    description = "Nome do Bucket Cloud Storage"
    type = string
}

