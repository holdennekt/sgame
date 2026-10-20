resource "random_password" "mongo" {
  length  = 32
  special = false
}

resource "random_password" "redis" {
  length  = 32
  special = false
}

resource "random_password" "grafana" {
  length  = 32
  special = false
}

locals {
  # Written to /opt/sgame/.env on the instance and read by docker-compose.prod.yaml.
  env_file = <<-EOT
    MONGO_ROOT_USER=root
    MONGO_ROOT_PASSWORD=${random_password.mongo.result}
    MONGO_NAME=sgame
    REDIS_USERNAME=sgame
    REDIS_PASSWORD=${random_password.redis.result}
    GRAFANA_ADMIN_PASSWORD=${random_password.grafana.result}
    STORAGE_PROVIDER=s3
    AWS_REGION=${var.region}
    RAW_BUCKET_NAME=${aws_s3_bucket.raw.bucket}
    PROCESSED_BUCKET_NAME=${aws_s3_bucket.processed.bucket}
    SQS_QUEUE_URL=${aws_sqs_queue.uploads.url}
    VALIDATOR_TYPE=claude-bedrock
    CLAUDE_MODEL_ID=${var.claude_model_id}
    IMAGE_REPOSITORY=${local.image_registry}/${local.images_repository}
  EOT
}

resource "aws_ssm_parameter" "env" {
  name  = local.env_param
  type  = "SecureString"
  value = local.env_file
}

# The deSEC token is added by hand (kept out of Terraform state), see README.
