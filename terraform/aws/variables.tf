variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "availability_zone" {
  type        = string
  default     = null
  description = "Defaults to the first available zone of the region. The instance and its data volume must share it."
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
  description = "Only workflow runs on this ref can assume the CI role."
}

variable "create_github_oidc_provider" {
  type        = bool
  default     = true
  description = "An account holds one GitHub OIDC provider; set false if it already exists."
}

variable "instance_type" {
  type        = string
  default     = "t3.large"
  description = "2 vCPU / 8 GiB, the same size as the GCP e2-standard-2."
}

variable "boot_volume_size_gb" {
  type    = number
  default = 30
}

variable "data_volume_size_gb" {
  type        = number
  default     = 50
  description = "EBS volume for Docker data (MongoDB, Redis, Prometheus, Grafana) and TLS certs."
}

variable "raw_bucket_name" {
  type        = string
  default     = null
  description = "Defaults to sgame-<account_id>-attachments-raw. Bucket names are global."
}

variable "processed_bucket_name" {
  type        = string
  default     = null
  description = "Defaults to sgame-<account_id>-attachments-processed."
}

variable "buckets_force_destroy" {
  type    = bool
  default = false
}

variable "claude_model_id" {
  type        = string
  default     = "anthropic.claude-haiku-5-5"
  description = "Claude model for the answer validator (CLAUDE_MODEL_ID); Bedrock IDs carry the anthropic. prefix."
}

variable "termination_protection" {
  type    = bool
  default = true
}
