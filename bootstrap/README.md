# Bootstrap

Applied once, manually, using local AWS credentials — not part of the pipeline.

```
cd bootstrap
terraform init
terraform apply
```

Creates: the GitHub OIDC identity provider, the IAM role GitHub Actions assumes
(trust policy scoped to this repo's `checks`/`provision`/`teardown` job identities),
the Terraform remote state bucket, and the SSM file-transfer bucket.

After applying, set the following repository variables (not secrets — an IAM role
ARN is not a credential):

```
gh api repos/mahmoudnasser1561/zero-credential-pipeline/actions/variables \
  -f name=AWS_ROLE_ARN -f value="$(terraform output -raw github_actions_role_arn)"
gh api repos/mahmoudnasser1561/zero-credential-pipeline/actions/variables \
  -f name=AWS_REGION -f value="$(terraform output -raw aws_region)"
```
