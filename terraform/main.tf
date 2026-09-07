module "network" {
  source = "./modules/network"

  project     = var.project
  environment = var.environment
}

module "compute" {
  source = "./modules/compute"

  project                 = var.project
  environment             = var.environment
  vpc_id                  = module.network.vpc_id
  subnet_id               = module.network.public_subnet_id
  ssm_transfer_bucket_arn = var.ssm_transfer_bucket_arn
  instance_type           = var.instance_type
}
