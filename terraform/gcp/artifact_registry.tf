resource "google_artifact_registry_repository" "images" {
  repository_id = local.images_repository
  location      = var.region
  format        = "DOCKER"
  description   = "SGame backend and frontend images"

  cleanup_policy_dry_run = false

  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions {
      keep_count = 10
    }
  }

  cleanup_policies {
    id     = "delete-old"
    action = "DELETE"
    condition {
      older_than = "2592000s"
    }
  }

  depends_on = [google_project_service.apis]
}
