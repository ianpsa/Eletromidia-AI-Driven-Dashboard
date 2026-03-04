variable "project_id" {
    description = "Id do Projeto no GCP - Dev Enviroment"
    type = string
}

variable "dataset_id_dev" {
    description = "Nome do Dataset do BigQuery - Dev Enviroment"
    type = string
}

variable "location" {
    description = "Localização do projeto"
}

variable "tables" {
    description = "Lista de tabelas do Dataset"
    type = map(object({
        table_id = string
        schema = string
    }))
}