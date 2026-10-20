locals {
  raw_bucket       = coalesce(var.raw_bucket_name, "${var.project_id}-attachments-raw")
  processed_bucket = coalesce(var.processed_bucket_name, "${var.project_id}-attachments-processed")

  uploads_topic        = "sgame-attachment-uploads"
  uploads_subscription = "sgame-attachment-uploads-sub"
  images_repository    = "sgame-images"

  admin_members = toset(var.admin_members)
}
