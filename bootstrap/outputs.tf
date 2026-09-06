output "github_actions_role_arn" {
  description = "IAM role ARN GitHub Actions assumes via OIDC. Set this as the repo variable AWS_ROLE_ARN."
  value       = aws_iam_role.github_actions.arn
}

output "tf_state_bucket" {
  description = "S3 bucket name for Terraform remote state (terraform/backend.tf)"
  value       = aws_s3_bucket.tf_state.id
}

output "ssm_transfer_bucket" {
  description = "S3 bucket name for SSM file transfer (ansible.cfg)"
  value       = aws_s3_bucket.ssm_transfer.id
}

output "aws_region" {
  description = "AWS region all resources are created in"
  value       = var.aws_region
}
