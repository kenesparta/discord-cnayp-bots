SHELL:=/bin/bash
fmt:
	@terraform fmt --recursive

dev/init: fmt
	@terraform init \
		-backend-config="bucket=tf-cnayp-prod" \
		-backend-config="prefix=prod/bots"

dev/plan: dev/init
	@terraform plan -out create.tfplan

dev/apply:
	@terraform apply "create.tfplan"

dev/destroy: dev/init
	@terraform destroy