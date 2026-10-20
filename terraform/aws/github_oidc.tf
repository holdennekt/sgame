# GitHub Actions gets short-lived credentials through OIDC, limited to this repo on var.github_deploy_ref.

resource "aws_iam_openid_connect_provider" "github" {
  count = var.create_github_oidc_provider ? 1 : 0

  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
}

data "aws_iam_openid_connect_provider" "github" {
  count = var.create_github_oidc_provider ? 0 : 1

  url = "https://token.actions.githubusercontent.com"
}

locals {
  github_oidc_provider_arn = var.create_github_oidc_provider ? aws_iam_openid_connect_provider.github[0].arn : data.aws_iam_openid_connect_provider.github[0].arn
}

resource "aws_iam_role" "github" {
  name = "sgame-github-actions"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = local.github_oidc_provider_arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_repository}:ref:${var.github_deploy_ref}"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "github" {
  name = "sgame-deploy"
  role = aws_iam_role.github.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "RegistryLogin"
        Effect   = "Allow"
        Action   = "ecr:GetAuthorizationToken"
        Resource = "*"
      },
      {
        Sid    = "RegistryPush"
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage",
        ]
        Resource = [for r in aws_ecr_repository.images : r.arn]
      },
      {
        Sid    = "RestartApp"
        Effect = "Allow"
        Action = "ssm:SendCommand"
        Resource = [
          aws_instance.sgame.arn,
          "arn:aws:ssm:${var.region}::document/AWS-RunShellScript",
        ]
      },
      {
        Sid      = "ReadCommandResult"
        Effect   = "Allow"
        Action   = "ssm:GetCommandInvocation"
        Resource = "*"
      },
    ]
  })
}
