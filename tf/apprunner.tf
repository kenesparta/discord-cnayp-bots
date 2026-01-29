resource "aws_apprunner_service" "bot" {
  service_name = var.app_name

  source_configuration {
    authentication_configuration {
      access_role_arn = aws_iam_role.apprunner_ecr_access.arn
    }

    image_repository {
      image_configuration {
        port = "8080"
        runtime_environment_secrets = {
          DISCORD_BOT_TOKEN = aws_secretsmanager_secret.discord_bot_token.arn
          DISCORD_GUILD_ID  = aws_secretsmanager_secret.discord_guild_id.arn
        }
      }
      image_identifier      = "${aws_ecr_repository.bot.repository_url}:latest"
      image_repository_type = "ECR"
    }

    auto_deployments_enabled = false
  }

  instance_configuration {
    cpu               = "256"
    memory            = "512"
    instance_role_arn = aws_iam_role.apprunner_instance.arn
  }

  health_check_configuration {
    protocol            = "TCP"
    interval            = 10
    timeout             = 5
    healthy_threshold   = 1
    unhealthy_threshold = 5
  }

  depends_on = [
    aws_iam_role_policy_attachment.apprunner_ecr_access
  ]
}
