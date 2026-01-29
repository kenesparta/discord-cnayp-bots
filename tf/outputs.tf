output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = aws_ecr_repository.bot.repository_url
}

output "apprunner_service_url" {
  description = "App Runner service URL"
  value       = aws_apprunner_service.bot.service_url
}

output "apprunner_service_arn" {
  description = "App Runner service ARN"
  value       = aws_apprunner_service.bot.arn
}

output "github_actions_role_arn" {
  description = "IAM role ARN for GitHub Actions"
  value       = aws_iam_role.github_actions.arn
}
