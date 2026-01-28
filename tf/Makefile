SHELL:=/bin/bash
dev/init:
	terraform init \
		-backend-config="bucket=tf-cnayp-prod" \
		-backend-config="prefix=prod/bots"

dev/plan: dev/init
	terraform plan -out create.tfplan

dev/apply: dev/plan
	terraform apply "create.tfplan"

dev/destroy: dev/init
	terraform destroy
