# ACLs stay disabled (S3 default); "public/" objects in the processed bucket are readable through the bucket policy instead.

resource "aws_s3_bucket" "raw" {
  bucket        = local.raw_bucket
  force_destroy = var.buckets_force_destroy
}

resource "aws_s3_bucket_public_access_block" "raw" {
  bucket                  = aws_s3_bucket.raw.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_cors_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  cors_rule {
    allowed_origins = ["https://${var.domain}"]
    allowed_methods = ["POST"]
    allowed_headers = ["*"]
    max_age_seconds = 3600
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "raw" {
  bucket = aws_s3_bucket.raw.id

  rule {
    id     = "expire-staging"
    status = "Enabled"
    filter {}

    expiration {
      days = 1
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }
}

resource "aws_s3_bucket_notification" "raw_uploads" {
  bucket = aws_s3_bucket.raw.id

  queue {
    queue_arn = aws_sqs_queue.uploads.arn
    events    = ["s3:ObjectCreated:*"]
  }

  depends_on = [aws_sqs_queue_policy.uploads]
}

resource "aws_s3_bucket" "processed" {
  bucket        = local.processed_bucket
  force_destroy = var.buckets_force_destroy
}

resource "aws_s3_bucket_public_access_block" "processed" {
  bucket                  = aws_s3_bucket.processed.id
  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "processed_public" {
  bucket = aws_s3_bucket.processed.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicReadPublicPrefix"
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.processed.arn}/public/*"
    }]
  })

  depends_on = [aws_s3_bucket_public_access_block.processed]
}

# Versioning + 7-day noncurrent expiry stands in for GCS soft delete.
resource "aws_s3_bucket_versioning" "processed" {
  bucket = aws_s3_bucket.processed.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "processed" {
  bucket = aws_s3_bucket.processed.id

  rule {
    id     = "soft-delete"
    status = "Enabled"
    filter {}

    noncurrent_version_expiration {
      noncurrent_days = 7
    }

    expiration {
      expired_object_delete_marker = true
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }

  depends_on = [aws_s3_bucket_versioning.processed]
}
