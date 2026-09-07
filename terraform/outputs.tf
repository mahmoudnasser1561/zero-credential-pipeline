output "vpc_id" {
  value = module.network.vpc_id
}

output "public_subnet_id" {
  value = module.network.public_subnet_id
}

output "instance_id" {
  value = module.compute.instance_id
}

output "security_group_id" {
  value = module.compute.security_group_id
}
