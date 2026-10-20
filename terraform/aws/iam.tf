resource "aws_iam_role" "vm" {
  name = "sgame-vm"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "vm_ssm_core" {
  role       = aws_iam_role.vm.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy" "vm" {
  name = "sgame-vm"
  role = aws_iam_role.vm.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Objects"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:AbortMultipartUpload",
          "s3:GetObjectTagging",
          "s3:PutObjectTagging",
        ]
        Resource = ["${aws_s3_bucket.raw.arn}/*", "${aws_s3_bucket.processed.arn}/*"]
      },
      {
        # HeadBucket (startup check) and 404s instead of 403s on missing keys.
        Sid      = "Buckets"
        Effect   = "Allow"
        Action   = "s3:ListBucket"
        Resource = [aws_s3_bucket.raw.arn, aws_s3_bucket.processed.arn]
      },
      {
        Sid    = "UploadQueue"
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:ChangeMessageVisibility",
          "sqs:GetQueueAttributes",
        ]
        Resource = aws_sqs_queue.uploads.arn
      },
      {
        Sid      = "RegistryLogin"
        Effect   = "Allow"
        Action   = "ecr:GetAuthorizationToken"
        Resource = "*"
      },
      {
        Sid    = "RegistryPull"
        Effect = "Allow"
        Action = [
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchCheckLayerAvailability",
        ]
        Resource = [for r in aws_ecr_repository.images : r.arn]
      },
      {
        Sid      = "Parameters"
        Effect   = "Allow"
        Action   = "ssm:GetParameter"
        Resource = "arn:aws:ssm:${var.region}:${local.account_id}:parameter${local.ssm_prefix}/*"
      },
      {
        Sid      = "Claude"
        Effect   = "Allow"
        Action   = "bedrock-mantle:CreateInference"
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_instance_profile" "vm" {
  name = "sgame-vm"
  role = aws_iam_role.vm.name
}

resource "aws_iam_role" "dlm" {
  name = "sgame-dlm"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "dlm.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "dlm" {
  role       = aws_iam_role.dlm.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSDataLifecycleManagerServiceRole"
}
