
module "bigquery" {                     // Layer responsável por inicializar o módulo do BigQuery
    source = "./modules/bigquery"

    project_id      = var.project_id
    location        = var.location
    dataset_id      = var.dataset_id
    table_id        = var.table_id   
}

resource "google_project_iam_member" "bq_admin" {       // Cria identidade de acesso `bq_admin`
    project         = var.project_id
    role            = "roles/bigquery.admin"
    member          =  "user:${var.admin_email}"
}

resource "google_project_iam_member" "bq_editor" {      // Cria identidade de acesso `bq_editor` para os serviços
    project         = var.project_id
    role            = "roles/bigquery.dataEditor"
    member          = "serviceAccount:${google_service_account.bq_sa.email}"
}