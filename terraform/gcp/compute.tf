data "google_compute_image" "debian" {
  family  = "debian-12"
  project = "debian-cloud"
}

resource "google_compute_disk" "data" {
  name = "sgame-data"
  type = "pd-balanced"
  zone = var.zone
  size = var.data_disk_size_gb

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [google_project_service.apis]
}

resource "google_compute_resource_policy" "daily_snapshots" {
  name   = "sgame-daily-snapshots"
  region = var.region

  snapshot_schedule_policy {
    schedule {
      daily_schedule {
        days_in_cycle = 1
        start_time    = "02:00"
      }
    }

    retention_policy {
      max_retention_days    = 7
      on_source_disk_delete = "KEEP_AUTO_SNAPSHOTS"
    }

    snapshot_properties {
      storage_locations = [lower(var.bucket_location)]
    }
  }
}

resource "google_compute_disk_resource_policy_attachment" "data" {
  name = google_compute_resource_policy.daily_snapshots.name
  disk = google_compute_disk.data.name
  zone = var.zone
}

resource "google_compute_instance" "sgame" {
  name                      = "sgame"
  zone                      = var.zone
  machine_type              = var.machine_type
  tags                      = ["sgame"]
  allow_stopping_for_update = true
  deletion_protection       = var.deletion_protection

  boot_disk {
    initialize_params {
      image = data.google_compute_image.debian.self_link
      size  = var.boot_disk_size_gb
      type  = "pd-balanced"
    }
  }

  attached_disk {
    source      = google_compute_disk.data.id
    device_name = "sgame-data"
    mode        = "READ_WRITE"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.sgame.id

    # Ephemeral public IP: start-app.sh re-points the deSEC record at every boot.
    access_config {}
  }

  service_account {
    email  = google_service_account.vm.email
    scopes = ["cloud-platform"]
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  metadata = {
    enable-oslogin = "TRUE"
    startup-script = templatefile("${path.module}/startup.sh.tftpl", {
      domain            = var.domain
      repo_url          = "https://github.com/${var.github_repository}.git"
      env_secret        = google_secret_manager_secret.env.secret_id
      desec_secret      = google_secret_manager_secret.desec_token.secret_id
      letsencrypt_email = var.letsencrypt_email
    })
  }

  lifecycle {
    ignore_changes = [boot_disk[0].initialize_params[0].image]
  }

  depends_on = [
    google_secret_manager_secret_version.env,
    google_secret_manager_secret_iam_member.vm_secrets,
    google_artifact_registry_repository_iam_member.vm_reader,
  ]
}
