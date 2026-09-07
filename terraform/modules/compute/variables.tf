variable "project" {
  type = string
}

variable "environment" {
  type    = string
  default = "demo"
}

variable "vpc_id" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "ssm_transfer_bucket_arn" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.small"
}
