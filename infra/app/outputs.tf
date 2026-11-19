output "app_url" {
  description = "FE and API share this origin; the API lives under /api."
  value       = "https://${aws_cloudfront_distribution.main.domain_name}"
}

output "api_items_url" {
  value = "https://${aws_cloudfront_distribution.main.domain_name}/api/v1/items"
}

output "cloudfront_distribution_id" {
  value = aws_cloudfront_distribution.main.id
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  value = aws_ecs_service.api.name
}

output "db_endpoint" {
  value = aws_db_instance.main.address
}
