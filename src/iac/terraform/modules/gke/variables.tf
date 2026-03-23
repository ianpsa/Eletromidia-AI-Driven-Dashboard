variable "project_id" {
  type = string
  description = "ID do Projeto utilizado pelo GKE"
}

variable "cluster_name" {
    type = string
    description = "Nome do Cluster GKE"
}

variable "zone" {
  type = string
  description = "Região utilizada pelo Cluster GKE"
}

variable "gke_subnet_id" {
  type = string
  description = "Subnet reservada ao GKE"
}

variable "vpc_id" {
  type = string
  description = "Id da Virtual Private Cloud do Projeto"
}

