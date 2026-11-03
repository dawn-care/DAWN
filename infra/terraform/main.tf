resource "aws_vpc" "facility" {
  cidr_block           = "10.20.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                = { Name = "dawn-${var.facility}" }
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.facility.id
  cidr_block = "10.20.1.0/24"
  tags       = { Name = "dawn-${var.facility}-public" }
}

resource "aws_internet_gateway" "facility" {
  vpc_id = aws_vpc.facility.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.facility.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.facility.id
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "facility" {
  name_prefix = "dawn-${var.facility}-"
  description = "Public web traffic and administrator-only SSH"
  vpc_id      = aws_vpc.facility.id

  ingress {
    description = "Administrator SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.admin_cidr]
  }

  ingress {
    description = "HTTP redirects and certificate issuance"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS and secure WebSockets"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Package downloads and application external APIs"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "admin" {
  key_name_prefix = "dawn-${var.facility}-"
  public_key      = var.ssh_public_key
}

resource "aws_instance" "facility" {
  ami                         = var.ami_id
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.public.id
  vpc_security_group_ids      = [aws_security_group.facility.id]
  key_name                    = aws_key_pair.admin.key_name
  associate_public_ip_address = false
  user_data                   = file("${path.module}/bootstrap.sh")
  user_data_replace_on_change = true
  tags                        = { Name = "dawn-${var.facility}" }

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.storage_gb
    encrypted             = true
    delete_on_termination = false
    tags                  = { Name = "dawn-${var.facility}-data" }
  }

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  credit_specification {
    cpu_credits = "standard"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_eip" "facility" {
  domain   = "vpc"
  instance = aws_instance.facility.id

  depends_on = [aws_internet_gateway.facility, aws_route_table_association.public]
}

resource "aws_route53_record" "facility" {
  count   = var.dns_zone_id == null ? 0 : 1
  zone_id = var.dns_zone_id
  name    = var.hostname
  type    = "A"
  ttl     = 300
  records = [aws_eip.facility.public_ip]

  lifecycle {
    precondition {
      condition     = var.hostname != null && var.hostname != ""
      error_message = "Set hostname when enabling a Route 53 record."
    }
  }
}
