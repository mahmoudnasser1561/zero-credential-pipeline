terraform {
  backend "s3" {
    bucket       = "zero-credential-pipeline-tfstate-334687118059"
    key          = "zero-credential-pipeline/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
