resource "aws_vpc" "sgame" {
  cidr_block           = "10.10.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "sgame"
  }
}

resource "aws_internet_gateway" "sgame" {
  vpc_id = aws_vpc.sgame.id

  tags = {
    Name = "sgame"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.sgame.id
  cidr_block              = "10.10.0.0/24"
  availability_zone       = local.az
  map_public_ip_on_launch = true

  tags = {
    Name = "sgame-public"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.sgame.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.sgame.id
  }

  tags = {
    Name = "sgame-public"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# Free gateway endpoint: S3 traffic from the instance stays inside AWS.
resource "aws_vpc_endpoint" "s3" {
  vpc_id            = aws_vpc.sgame.id
  service_name      = "com.amazonaws.${var.region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = [aws_route_table.public.id]
}

# No SSH port: shell access goes through SSM Session Manager.
resource "aws_security_group" "sgame" {
  name        = "sgame"
  description = "SGame web traffic"
  vpc_id      = aws_vpc.sgame.id
}

resource "aws_vpc_security_group_ingress_rule" "web" {
  for_each = toset(["80", "443"])

  security_group_id = aws_security_group.sgame.id
  description       = "web ${each.value}"
  ip_protocol       = "tcp"
  from_port         = tonumber(each.value)
  to_port           = tonumber(each.value)
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.sgame.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}
