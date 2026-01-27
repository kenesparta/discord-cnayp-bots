SHELL:=/bin/bash
dev/init:
	terraform init \
		-backend-config="bucket=tf-cnayp-prod" \
		-backend-config="prefix=prod/bots"

dev/plan:
	source .env && terraform plan -out create.tfplan

dev/apply:
	source .env && terraform apply "create.tfplan"

dev/destroy: dev/init
	source .env && terraform destroy
