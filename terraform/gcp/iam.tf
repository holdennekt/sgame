# --- Runtime identity of the VM ----------------------------------------------

resource "google_service_account" "vm" {
  account_id   = "sgame-vm"
  display_name = "SGame VM runtime"

  depends_on = [google_project_service.apis]
}

resource "google_project_iam_member" "vm_vertex" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = google_service_account.vm.member
}

# Lets the Go storage client sign URLs and POST policies through IAM signBlob (no key file).
resource "google_service_account_iam_member" "vm_signs_for_itself" {
  service_account_id = google_service_account.vm.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = google_service_account.vm.member
}

resource "google_storage_bucket_iam_member" "vm_objects" {
  for_each = {
    raw       = google_storage_bucket.raw.name
    processed = google_storage_bucket.processed.name
  }

  bucket = each.value
  role   = "roles/storage.objectAdmin"
  member = google_service_account.vm.member
}

# storage.buckets.get, needed by the startup check in NewGCSStorage.
resource "google_storage_bucket_iam_member" "vm_bucket_reader" {
  for_each = {
    raw       = google_storage_bucket.raw.name
    processed = google_storage_bucket.processed.name
  }

  bucket = each.value
  role   = "roles/storage.legacyBucketReader"
  member = google_service_account.vm.member
}

resource "google_pubsub_subscription_iam_member" "vm_subscriber" {
  subscription = google_pubsub_subscription.uploads.name
  role         = "roles/pubsub.subscriber"
  member       = google_service_account.vm.member
}

resource "google_artifact_registry_repository_iam_member" "vm_reader" {
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.name
  role       = "roles/artifactregistry.reader"
  member     = google_service_account.vm.member
}

resource "google_secret_manager_secret_iam_member" "vm_secrets" {
  for_each = {
    env   = google_secret_manager_secret.env.id
    desec = google_secret_manager_secret.desec_token.id
  }

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.vm.member
}

# --- CI identity (GitHub Actions through Workload Identity Federation) ------------

resource "google_service_account" "github" {
  account_id   = "github-actions-sa"
  display_name = "GitHub Actions deploy"

  depends_on = [google_project_service.apis]
}

resource "google_artifact_registry_repository_iam_member" "github_writer" {
  location   = google_artifact_registry_repository.images.location
  repository = google_artifact_registry_repository.images.name
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.github.member
}

resource "google_project_iam_member" "github_compute_viewer" {
  project = var.project_id
  role    = "roles/compute.viewer"
  member  = google_service_account.github.member
}

resource "google_service_account_iam_member" "github_acts_as_vm" {
  service_account_id = google_service_account.vm.name
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.github.member
}

resource "google_compute_instance_iam_member" "github_os_admin_login" {
  project       = var.project_id
  zone          = var.zone
  instance_name = google_compute_instance.sgame.name
  role          = "roles/compute.osAdminLogin"
  member        = google_service_account.github.member
}

resource "google_iap_tunnel_instance_iam_member" "github_tunnel" {
  project  = var.project_id
  zone     = var.zone
  instance = google_compute_instance.sgame.name
  role     = "roles/iap.tunnelResourceAccessor"
  member   = google_service_account.github.member
}

# --- Humans who may SSH in (IAP + OS Login, with sudo) --------------------------------

resource "google_compute_instance_iam_member" "admin_os_admin_login" {
  for_each = local.admin_members

  project       = var.project_id
  zone          = var.zone
  instance_name = google_compute_instance.sgame.name
  role          = "roles/compute.osAdminLogin"
  member        = each.value
}

resource "google_iap_tunnel_instance_iam_member" "admin_tunnel" {
  for_each = local.admin_members

  project  = var.project_id
  zone     = var.zone
  instance = google_compute_instance.sgame.name
  role     = "roles/iap.tunnelResourceAccessor"
  member   = each.value
}

resource "google_service_account_iam_member" "admin_acts_as_vm" {
  for_each = local.admin_members

  service_account_id = google_service_account.vm.name
  role               = "roles/iam.serviceAccountUser"
  member             = each.value
}

resource "google_project_iam_member" "admin_compute_viewer" {
  for_each = local.admin_members

  project = var.project_id
  role    = "roles/compute.viewer"
  member  = each.value
}
