locals {
  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

data "aws_ssm_parameter" "al2023" {
  name = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
}

data "aws_iam_policy_document" "instance_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "instance" {
  name               = "${var.project}-instance-role"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role.json

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

data "aws_iam_policy_document" "ssm_transfer_bucket" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:ListBucket",
    ]
    resources = [
      var.ssm_transfer_bucket_arn,
      "${var.ssm_transfer_bucket_arn}/*",
    ]
  }
}

resource "aws_iam_role_policy" "ssm_transfer_bucket" {
  name   = "${var.project}-ssm-transfer-bucket"
  role   = aws_iam_role.instance.id
  policy = data.aws_iam_policy_document.ssm_transfer_bucket.json
}

resource "aws_iam_instance_profile" "instance" {
  name = "${var.project}-instance-profile"
  role = aws_iam_role.instance.name

  tags = local.tags
}

# tfsec:ignore:aws-ec2-no-public-egress-sgr -- 443 egress required for SSM endpoints, package repos, GitHub releases
resource "aws_security_group" "instance" {
  name_prefix = "${var.project}-instance-"
  description = "Zero-ingress instance SG; egress limited to 443 for SSM/package endpoints"
  vpc_id      = var.vpc_id

  egress {
    description = "HTTPS to SSM endpoints, package repos, GitHub releases"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, { Name = "${var.project}-instance-sg" })

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_instance" "this" {
  ami                    = data.aws_ssm_parameter.al2023.value
  instance_type          = var.instance_type
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [aws_security_group.instance.id]
  iam_instance_profile   = aws_iam_instance_profile.instance.name

  metadata_options {
    http_tokens   = "required"
    http_endpoint = "enabled"
  }

  root_block_device {
    encrypted = true
  }

  tags = merge(local.tags, { Name = "${var.project}-instance" })
}
