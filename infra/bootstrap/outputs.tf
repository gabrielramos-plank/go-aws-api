# Copy these into GitHub: Settings → Secrets and variables → Actions → Variables.

output "TF_STATE_BUCKET" {
  value = aws_s3_bucket.tfstate.bucket
}

output "AWS_DEPLOY_ROLE_ARN" {
  value = aws_iam_role.github_deploy.arn
}

output "AWS_TERRAFORM_ROLE_ARN" {
  value = aws_iam_role.github_terraform.arn
}

output "FRONTEND_BUCKET" {
  value = aws_s3_bucket.frontend.bucket
}

output "AWS_FRONTEND_ROLE_ARN" {
  value = one(aws_iam_role.github_frontend[*].arn)
}

output "ecr_repository_url" {
  value = aws_ecr_repository.api.repository_url
}
