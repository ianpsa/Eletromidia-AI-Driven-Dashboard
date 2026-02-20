resource "google_storage_bucket" "kafka_backup" {
    name        = var.cs_name
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

resource "google_storage_managed_folder" "topics" {
    bucket = google_storage_bucket.kafka_backup.name
    name = "topics/"
}

resource "google_service_account" "cs_sa" {
  project   = var.project_id
  account_id = "cs-kafka-consumer"
  display_name = "Cloud Storage Kafka Backup SA"
}


resource "google_storage_bucket_iam_member" "cs_sa" {
  bucket = google_storage_bucket.kafka_backup.name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${google_service_account.cs_sa.email}"
}
