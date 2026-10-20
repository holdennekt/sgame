# One-off: creates the remote-state bucket; this config keeps its own tiny state locally (git-ignored).

terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

variable "region" {
  type    = string
  default = "eu-central-1"
}

provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "tfstate" {
  bucket        = "sgame-${data.aws_caller_identity.current.account_id}-tfstate"
  force_destroy = false
}

resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id

  rule {
    id     = "keep-10-versions"
    status = "Enabled"
    filter {}

    noncurrent_version_expiration {
      noncurrent_days           = 1
      newer_noncurrent_versions = 10
    }
  }

  depends_on = [aws_s3_bucket_versioning.tfstate]
}

output "backend_hcl" {
  value = <<-EOT
    bucket       = "${aws_s3_bucket.tfstate.bucket}"
    key          = "sgame/terraform.tfstate"
    region       = "${var.region}"
    use_lockfile = true
    encrypt      = true
  EOT
}
