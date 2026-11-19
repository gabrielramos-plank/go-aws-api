variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "project" {
  description = "Name prefix for all resources. Must match infra/bootstrap and the workflows."
  type        = string
  default     = "go-aws-api"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "desired_count" {
  type    = number
  default = 1
}

variable "task_cpu" {
  description = "Fargate CPU units (256 = 0.25 vCPU, the smallest/cheapest)."
  type        = number
  default     = 256
}

variable "task_memory" {
  description = "Fargate memory in MiB."
  type        = number
  default     = 512
}

variable "db_instance_class" {
  type    = string
  default = "db.t4g.micro"
}

variable "db_name" {
  type    = string
  default = "app"
}

variable "db_username" {
  type    = string
  default = "app_admin"
}
