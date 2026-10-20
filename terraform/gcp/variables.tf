variable "project_id" {
  type        = string
  description = "Existing GCP project that will host everything (billing must be enabled)."
}

variable "region" {
  type    = string
  default = "europe-central2"
}

variable "zone" {
  type    = string
  default = "europe-central2-c"
}

variable "domain" {
  type        = string
  default     = "sgame.dedyn.io"
  description = "Public domain, managed at deSEC. Used for TLS, CORS and the app URL."
}

variable "letsencrypt_email" {
  type        = string
  default     = ""
  description = "Contact email for Let's Encrypt expiry notices. Empty registers without an email."
}

variable "github_repository" {
  type    = string
  default = "holdennekt/sgame"
}

variable "github_deploy_ref" {
  type        = string
  default     = "refs/heads/main"
  description = "Only workflow runs on this ref can impersonate the CI service account."
}

variable "machine_type" {
  type    = string
  default = "e2-standard-2"
}

variable "boot_disk_size_gb" {
  type    = number
  default = 30
}

variable "data_disk_size_gb" {
  type        = number
  default     = 50
  description = "Persistent disk for Docker data (MongoDB, Redis, Prometheus, Grafana) and TLS certs."
}

variable "bucket_location" {
  type    = string
  default = "EU"
}

variable "raw_bucket_name" {
  type        = string
  default     = null
  description = "Defaults to <project_id>-attachments-raw. Bucket names are global, so the old names may be unavailable."
}

variable "processed_bucket_name" {
  type        = string
  default     = null
  description = "Defaults to <project_id>-attachments-processed."
}

variable "buckets_force_destroy" {
  type    = bool
  default = false
}

variable "vertex_location" {
  type        = string
  default     = "global"
  description = "Vertex AI location for the answer validator (GCP_LOCATION)."
}

variable "claude_model_id" {
  type        = string
  default     = "claude-haiku-5-5"
  description = "Claude model for the answer validator (CLAUDE_MODEL_ID); Vertex IDs have no prefix."
}

variable "admin_members" {
  type        = list(string)
  default     = []
  description = "IAM members (e.g. user:you@example.com) allowed to SSH into the VM through IAP with sudo."
}

variable "deletion_protection" {
  type    = bool
  default = true
}
