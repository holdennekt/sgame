data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_ebs_volume" "data" {
  availability_zone = local.az
  size              = var.data_volume_size_gb
  type              = "gp3"
  encrypted         = true

  tags = {
    Name     = "sgame-data"
    Snapshot = "sgame-daily"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_dlm_lifecycle_policy" "daily_snapshots" {
  description        = "sgame daily data snapshots"
  execution_role_arn = aws_iam_role.dlm.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]
    target_tags = {
      Snapshot = "sgame-daily"
    }

    schedule {
      name      = "daily"
      copy_tags = true

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["02:00"]
      }

      retain_rule {
        count = 7
      }
    }
  }
}

# Runs user data on every boot (cloud-init's default is first boot only), like a GCE startup-script.
data "cloudinit_config" "sgame" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/cloud-config"
    content = yamlencode({
      cloud_final_modules = [["scripts-user", "always"]]
    })
  }

  part {
    content_type = "text/x-shellscript"
    filename     = "sgame-startup.sh"
    content = templatefile("${path.module}/startup.sh.tftpl", {
      domain            = var.domain
      repo_url          = "https://github.com/${var.github_repository}.git"
      region            = var.region
      env_param         = local.env_param
      desec_param       = local.desec_param
      letsencrypt_email = var.letsencrypt_email
      data_volume_id    = replace(aws_ebs_volume.data.id, "-", "")
    })
  }
}

resource "aws_instance" "sgame" {
  ami                         = data.aws_ami.ubuntu.id
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.public.id
  vpc_security_group_ids      = [aws_security_group.sgame.id]
  iam_instance_profile        = aws_iam_instance_profile.vm.name
  associate_public_ip_address = true # ephemeral: start-app-aws.sh re-points the deSEC record at every boot
  disable_api_termination     = var.termination_protection
  user_data                   = data.cloudinit_config.sgame.rendered

  # Hop limit 2 lets the containers reach IMDSv2 through the Docker bridge for the instance role credentials.
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  root_block_device {
    volume_size = var.boot_volume_size_gb
    volume_type = "gp3"
    encrypted   = true
  }

  tags = {
    Name = "sgame"
  }

  lifecycle {
    ignore_changes = [ami]
  }

  depends_on = [aws_ssm_parameter.env]
}

resource "aws_volume_attachment" "data" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.sgame.id
}
