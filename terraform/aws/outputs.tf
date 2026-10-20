output "account_id" {
  value = local.account_id
}

output "instance_id" {
  value = aws_instance.sgame.id
}

output "instance_public_ip" {
  value       = aws_instance.sgame.public_ip
  description = "Ephemeral; the deSEC record is updated at every boot."
}

output "raw_bucket" {
  value = aws_s3_bucket.raw.bucket
}

output "processed_bucket" {
  value = aws_s3_bucket.processed.bucket
}

output "uploads_queue_url" {
  value = aws_sqs_queue.uploads.url
}

output "image_repository" {
  value = "${local.image_registry}/${local.images_repository}"
}

output "github_role_arn" {
  value = aws_iam_role.github.arn
}

output "github_setup_commands" {
  description = "Run these once from the repo to point GitHub Actions at AWS (DEPLOY_TARGET=aws also turns the GCP workflow off)."
  value       = <<-EOT
    gh secret set AWS_DEPLOY_ROLE_ARN --body '${aws_iam_role.github.arn}'
    gh variable set AWS_REGION --body '${var.region}'
    gh variable set AWS_INSTANCE_ID --body '${aws_instance.sgame.id}'
    gh variable set AWS_IMAGE_REPOSITORY --body '${local.image_registry}/${local.images_repository}'
    gh variable set DEPLOY_TARGET --body 'aws'
  EOT
}
