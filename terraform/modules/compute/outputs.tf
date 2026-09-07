output "instance_id" {
  value = aws_instance.this.id
}

output "instance_role_arn" {
  value = aws_iam_role.instance.arn
}

output "security_group_id" {
  value = aws_security_group.instance.id
}
