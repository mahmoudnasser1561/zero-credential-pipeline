variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project" {
  type    = string
  default = "zero-credential-pipeline"
}

variable "environment" {
  type    = string
  default = "demo"
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

variable "ssm_transfer_bucket_arn" {
  type    = string
  default = "arn:aws:s3:::zero-credential-pipeline-ssm-transfer-334687118059"
}
