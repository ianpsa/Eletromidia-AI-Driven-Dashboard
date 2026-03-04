resource "google_storage_bucket" "kafka_backup_dev" {
    name = var.cs_dev_name
    project     = var.project_id
    location    = var.location
    force_destroy = true
    uniform_bucket_level_access = true

    versioning {
    enabled = false
    }

    lifecycle_rule {
      condition {
        age = 30
      }
      action {
        type = "SetStorageClass"
        storage_class = "NEARLINE"
      }
    }

    lifecycle_rule {
      condition {
        age = 360
      }
      action {
        type = "Delete"
      }
    }
}

resource "google_service_account" "cs_sa_dev" {
    project = var.project_id
    account_id = "cs-dev-kafka-consumer"
    display_name = "Cloud Storage-Dev Kafka Backup SA"
}
