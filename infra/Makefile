# Paths
TERRAFORM_DIR=terraform
ANSIBLE_DIR=ansible

# Default target: plan, apply, inventory, ansible
all: terraform-plan terraform-apply generate-inventory ansible-run

# Run terraform plan.sh
terraform-plan:
	cd $(TERRAFORM_DIR) && ./plan.sh

# Run terraform apply.sh
terraform-apply:
	cd $(TERRAFORM_DIR) && ./apply.sh

# Generate inventory.ini from terraform outputs
generate-inventory:
	cd $(TERRAFORM_DIR) && terraform output -json | jq -r '\
	  "[db]\nleadstorefront-db ansible_host=\(.db_ip.value) ansible_user=root ansible_ssh_private_key_file=~/.ssh/leadstorefront\n\n\
	   [app]\nleadstorefront-app ansible_host=\(.app_ip.value) ansible_user=root ansible_ssh_private_key_file=~/.ssh/leadstorefront"' \
	  > ../$(ANSIBLE_DIR)/inventory.ini
	@echo "Ansible inventory generated."

# Run ansible playbook
ansible-run:
	set -a; . ../.env; set +a; cd $(ANSIBLE_DIR) && ansible-playbook -i inventory.ini playbook.yaml

ansible-run-app:
	set -a; . ../.env; set +a; cd $(ANSIBLE_DIR) && ansible-playbook -i inventory.ini playbook.yaml --limit app

ansible-run-db:
	set -a; . ../.env; set +a; cd $(ANSIBLE_DIR) && ansible-playbook -i inventory.ini playbook.yaml --limit db

# Destroy everything via destroy.sh
destroy:
	cd $(TERRAFORM_DIR) && ./destroy.sh

executable_permissions:
	chmod +x terraform/*.sh

docker-compose-local:
	docker compose -f docker-compose.local.yaml up postgres-platformdb

.PHONY: all terraform-plan terraform-apply generate-inventory ansible-run ansible-run-app ansible-run-db destroy executable_permissions docker-compose-local
