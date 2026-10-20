resource "google_compute_network" "sgame" {
  name                    = "sgame"
  auto_create_subnetworks = false

  depends_on = [google_project_service.apis]
}

resource "google_compute_subnetwork" "sgame" {
  name                     = "sgame-${var.region}"
  region                   = var.region
  network                  = google_compute_network.sgame.id
  ip_cidr_range            = "10.10.0.0/24"
  private_ip_google_access = true
}

resource "google_compute_firewall" "web" {
  name          = "sgame-allow-web"
  network       = google_compute_network.sgame.id
  direction     = "INGRESS"
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["sgame"]

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }
}

resource "google_compute_firewall" "iap_ssh" {
  name          = "sgame-allow-iap-ssh"
  network       = google_compute_network.sgame.id
  direction     = "INGRESS"
  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["sgame"]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}
