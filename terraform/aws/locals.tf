locals {
  account_id = data.aws_caller_identity.current.account_id
  az         = coalesce(var.availability_zone, data.aws_availability_zones.available.names[0])

  raw_bucket       = coalesce(var.raw_bucket_name, "sgame-${local.account_id}-attachments-raw")
  processed_bucket = coalesce(var.processed_bucket_name, "sgame-${local.account_id}-attachments-processed")

  images_repository = "sgame-images"
  image_names       = toset(["sgame-backend", "sgame-frontend"])
  image_registry    = "${local.account_id}.dkr.ecr.${var.region}.amazonaws.com"

  ssm_prefix  = "/sgame"
  env_param   = "${local.ssm_prefix}/env"
  desec_param = "${local.ssm_prefix}/DESEC_TOKEN"
}
