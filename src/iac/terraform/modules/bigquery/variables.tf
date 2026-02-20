variable "project_id" {
    description = "Id do Projeto no GCP"
    type = string
}

variable "location" {
    description = "Localização do BigQuery"
    type = string
}

variable "dataset_id" {
    description = "Nome do Dataset do BigQuery"
    type = string
}

variable "tables" {
  description = "Lista de tabelas do dataset"
  type = map(object({
    table_id = string
    schema   = string
  }))
}