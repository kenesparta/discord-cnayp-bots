SHELL:=/bin/bash

fmt:
	@terraform fmt --recursive

dev/init: fmt
	@terraform init \
		-backend-config="bucket=tf.cnayp.kenesparta.dev" \
		-backend-config="key=dns/prod/cnayp-bot" \
		-backend-config="region=us-east-1"

dev/plan: dev/init
	@terraform plan -out create.tfplan

dev/apply: dev/plan
	@terraform apply create.tfplan

dev/destroy: dev/init
	@terraform destroy

# Step 1: Apply ECR and IAM first (no App Runner)
dev/apply-base: dev/init
	@terraform apply \
		-target=aws_ecr_repository.bot \
		-target=aws_ecr_lifecycle_policy.bot \
		-target=aws_iam_role.github_actions \
		-target=aws_iam_role_policy.github_actions_ecr \
		-target=aws_iam_role.apprunner_ecr_access \
		-target=aws_iam_role_policy_attachment.apprunner_ecr_access \
		-target=aws_iam_role.apprunner_instance \
		-target=aws_secretsmanager_secret.discord_bot_token \
		-target=aws_secretsmanager_secret_version.discord_bot_token \
		-target=aws_secretsmanager_secret.discord_guild_id \
		-target=aws_secretsmanager_secret_version.discord_guild_id

# Step 2: Push initial image (run from repo root)
dev/push-initial:
	@ECR_URL=$$(terraform output -raw ecr_repository_url) && \
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $$ECR_URL && \
	cd .. && docker build -t $$ECR_URL:latest . && \
	docker push $$ECR_URL:latest

# Step 3: Apply App Runner after image exists
dev/apply-apprunner: dev/init
	@terraform apply \
		-target=aws_apprunner_service.bot \
		-target=aws_iam_role_policy.github_actions_apprunner \
		-target=aws_iam_role_policy.apprunner_secrets