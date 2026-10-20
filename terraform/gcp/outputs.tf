output "project_id" {
  value = var.project_id
}

output "vm_name" {
  value = google_compute_instance.sgame.name
}

output "vm_zone" {
  value = var.zone
}

output "vm_external_ip" {
  value       = google_compute_instance.sgame.network_interface[0].access_config[0].nat_ip
  description = "Ephemeral; the deSEC record is updated at every boot."
}

output "raw_bucket" {
  value = google_storage_bucket.raw.name
}

output "processed_bucket" {
  value = google_storage_bucket.processed.name
}

output "image_repository" {
  value = "${var.region}-docker.pkg.dev/${var.project_id}/${local.images_repository}"
}

output "github_workload_identity_provider" {
  value = google_iam_workload_identity_pool_provider.github.name
}

output "github_service_account" {
  value = google_service_account.github.email
}

output "github_setup_commands" {
  description = "Run these once from the repo to point GitHub Actions at the new project."
  value       = <<-EOT
    gh secret set GCP_WORKLOAD_IDENTITY_PROVIDER --body '${google_iam_workload_identity_pool_provider.github.name}'
    gh secret set GCP_SERVICE_ACCOUNT --body '${google_service_account.github.email}'
    gh variable set GCP_PROJECT_ID --body '${var.project_id}'
    gh variable set GCP_VM_NAME --body '${google_compute_instance.sgame.name}'
    gh variable set GCP_VM_ZONE --body '${var.zone}'
    gh variable set GCP_REGION --body '${var.region}'
  EOT
}
