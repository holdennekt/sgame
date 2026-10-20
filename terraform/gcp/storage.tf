# Both buckets keep object ACLs enabled: the backend sets public-read ACLs on "public/" objects.

resource "google_storage_bucket" "raw" {
  name                        = local.raw_bucket
  location                    = var.bucket_location
  storage_class               = "STANDARD"
  uniform_bucket_level_access = false
  public_access_prevention    = "inherited"
  force_destroy               = var.buckets_force_destroy

  soft_delete_policy {
    retention_duration_seconds = 604800
  }

  cors {
    origin          = ["https://${var.domain}"]
    method          = ["POST"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      age = 1
    }
  }

  depends_on = [google_project_service.apis]
}

resource "google_storage_bucket" "processed" {
  name                        = local.processed_bucket
  location                    = var.bucket_location
  storage_class               = "STANDARD"
  uniform_bucket_level_access = false
  public_access_prevention    = "inherited"
  force_destroy               = var.buckets_force_destroy

  soft_delete_policy {
    retention_duration_seconds = 604800
  }

  depends_on = [google_project_service.apis]
}

resource "google_storage_notification" "raw_uploads" {
  bucket         = google_storage_bucket.raw.name
  topic          = google_pubsub_topic.uploads.id
  payload_format = "JSON_API_V1"
  event_types    = ["OBJECT_FINALIZE"]

  depends_on = [google_pubsub_topic_iam_member.gcs_publisher]
}
