data "google_storage_project_service_account" "gcs" {
  depends_on = [google_project_service.apis]
}

resource "google_pubsub_topic" "uploads" {
  name = local.uploads_topic

  depends_on = [google_project_service.apis]
}

resource "google_pubsub_topic_iam_member" "gcs_publisher" {
  topic  = google_pubsub_topic.uploads.name
  role   = "roles/pubsub.publisher"
  member = data.google_storage_project_service_account.gcs.member
}

resource "google_pubsub_subscription" "uploads" {
  name  = local.uploads_subscription
  topic = google_pubsub_topic.uploads.id

  ack_deadline_seconds       = 60
  message_retention_duration = "604800s"

  expiration_policy {
    ttl = ""
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
}
