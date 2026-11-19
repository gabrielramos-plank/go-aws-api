variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project" {
  description = "Name prefix for all resources. Must match `project` in infra/app and the workflows."
  type        = string
  default     = "go-aws-api"
}

variable "github_repo" {
  description = "GitHub repository allowed to assume the CI roles, as \"owner/name\"."
  type        = string
}

variable "frontend_github_repo" {
  description = "Optional FE repository (\"owner/name\") allowed to deploy to the FE bucket. Empty = no role."
  type        = string
  default     = ""
}
