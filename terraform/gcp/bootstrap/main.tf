# One-off: creates the remote-state bucket; this config keeps its own tiny state locally (git-ignored).

terraform {
  required_version = ">= 1.9"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.7"
    }
  }
}

variable "project_id" {
  type = string
}

variable "location" {
  type    = string
  default = "EU"
}

provider "google" {
  project = var.project_id
}

resource "google_project_service" "storage" {
  service            = "storage.googleapis.com"
  disable_on_destroy = false
}

resource "google_storage_bucket" "tfstate" {
  name                        = "${var.project_id}-tfstate"
  location                    = var.location
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = false

  versioning {
    enabled = true
  }

  lifecycle_rule {
    action {
      type = "Delete"
    }
    condition {
      num_newer_versions = 10
      with_state         = "ARCHIVED"
    }
  }

  depends_on = [google_project_service.storage]
}

output "backend_hcl" {
  value = <<-EOT
    bucket = "${google_storage_bucket.tfstate.name}"
    prefix = "sgame"
  EOT
}
