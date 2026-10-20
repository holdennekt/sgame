resource "aws_sqs_queue" "uploads_dlq" {
  name                      = "sgame-attachment-uploads-dlq"
  message_retention_seconds = 1209600
  sqs_managed_sse_enabled   = true
}

resource "aws_sqs_queue" "uploads" {
  name                       = "sgame-attachment-uploads"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 604800
  receive_wait_time_seconds  = 20
  sqs_managed_sse_enabled    = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.uploads_dlq.arn
    maxReceiveCount     = 10
  })
}

resource "aws_sqs_queue_policy" "uploads" {
  queue_url = aws_sqs_queue.uploads.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowRawBucketNotifications"
      Effect    = "Allow"
      Principal = { Service = "s3.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.uploads.arn
      Condition = {
        ArnEquals    = { "aws:SourceArn" = aws_s3_bucket.raw.arn }
        StringEquals = { "aws:SourceAccount" = local.account_id }
      }
    }]
  })
}
