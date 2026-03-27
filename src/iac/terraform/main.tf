
module "bigquery" {
  source = "./modules/bigquery"

  project_id = var.project_id
  location   = var.location
  dataset_id = var.dataset_id
  tables     = var.tables
}

module "bigquery_dev" {
  source = "./modules/bigquery_dev"

  project_id     = var.project_id
  location       = var.location
  dataset_id_dev = var.dataset_id_dev
  tables         = var.tables
}

module "artifact_registry" {
  source = "./modules/artifact_registry"

  project_id = var.project_id
  region     = var.location
}

module "vpc" {
  source = "./modules/vpc"

  region       = var.location
  cluster_name = var.cluster_name
}

module "gke" {
  source = "./modules/gke"

  project_id    = var.project_id
  zone          = var.gke_zone
  cluster_name  = var.cluster_name
  gke_subnet_id = module.vpc.gke_subnet_id
  vpc_id        = module.vpc.vpc_id
}

module "iam" {
  source = "./modules/iam"

  project_id            = var.project_id
  admin_email           = var.admin_email
  service_account_email = module.bigquery.service_account_email
  gke_sa_email          = module.gke.gke_sa_email
}

module "iam_dev" {
  source = "./modules/iam_dev"

  project_id                   = var.project_id
  service_account_bq_dev_email = module.bigquery_dev.service_account_email_dev
  cs_dev_name                  = var.cs_dev_name
  service_account_cs_dev_email = module.cloud_storage_dev.service_account_cs_dev_email
}

module "cloud_storage" {
  source = "./modules/cloud_storage"

  project_id = var.project_id
  location   = var.location
  cs_name    = var.cs_name
}

module "cloud_storage_dev" {
  source = "./modules/cloud_storage_dev"

  project_id  = var.project_id
  location    = var.location
  cs_dev_name = var.cs_dev_name

}
