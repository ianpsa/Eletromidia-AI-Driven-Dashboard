
module "bigquery" { 
    source = "./modules/bigquery"

    project_id      = var.project_id
    location        = var.location
    dataset_id      = var.dataset_id
    tables          = var.tables
}

module "bigquery_dev" {
  source = "./modules/bigquery_dev"

  project_id = var.project_id
  location = var.location
  dataset_id_dev = var.dataset_id_dev
  tables = var.tables
}

module "iam" {
    source = "./modules/iam"

    project_id              = var.project_id
    admin_email             = var.admin_email
    service_account_email   = module.bigquery.service_account_email
}

module "cloud_storage" {
    source = "./modules/cloud_storage"

    project_id = var.project_id
    location   = var.location
    cs_name    = var.cs_name
}