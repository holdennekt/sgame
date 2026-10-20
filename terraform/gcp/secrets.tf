resource "random_password" "mongo" {
  length  = 32
  special = false
}

resource "random_password" "redis" {
  length  = 32
  special = false
}

resource "random_password" "grafana" {
  length  = 32
  special = false
}

locals {
  # Written to /opt/sgame/.env on the VM and read by docker-compose.prod.yaml.
  env_file = <<-EOT
    MONGO_ROOT_USER=root
    MONGO_ROOT_PASSWORD=${random_password.mongo.result}
    MONGO_NAME=sgame
    REDIS_USERNAME=sgame
    REDIS_PASSWORD=${random_password.redis.result}
    GRAFANA_ADMIN_PASSWORD=${random_password.grafana.result}
    GCP_PROJECT_ID=${var.project_id}
    GCP_LOCATION=${var.vertex_location}
    RAW_BUCKET_NAME=${local.raw_bucket}
    PROCESSED_BUCKET_NAME=${local.processed_bucket}
    PUBSUB_SUBSCRIPTION_ID=${local.uploads_subscription}
    STORAGE_PROVIDER=gcs
    VALIDATOR_TYPE=claude-vertex
    CLAUDE_MODEL_ID=${var.claude_model_id}
    IMAGE_REPOSITORY=${var.region}-docker.pkg.dev/${var.project_id}/${local.images_repository}
  EOT
}

resource "google_secret_manager_secret" "env" {
  secret_id = "sgame-env"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "env" {
  secret      = google_secret_manager_secret.env.id
  secret_data = local.env_file
}

# Value is added by hand to stay out of state: printf '%s' "$TOKEN" | gcloud secrets versions add DESEC_TOKEN --data-file=-
resource "google_secret_manager_secret" "desec_token" {
  secret_id = "DESEC_TOKEN"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}
