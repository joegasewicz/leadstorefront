# LeadStorefront Infra

Two DigitalOcean droplets are used:

- `leadstorefront-app` runs Caddy, the web service, and the API service.
- `leadstorefront-db` runs Postgres.

## Provisioning

1. Add the `~/.ssh/leadstorefront.pub` public key to DigitalOcean.
2. Create a root `.env` file with the required Terraform, Docker, and app variables.
3. Run Terraform and Ansible from this directory.

```sh
make terraform-plan
make terraform-apply
make generate-inventory
make ansible-run
```

## Required Root `.env` Values

```sh
DO_PAT=
DOCKER_USERNAME=
DOCKER_PASSWORD=

POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
POSTGRES_HOST=
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable

WEB_DOMAIN=leadstorefront.com
WEB_ADDR=:8000
API_DOMAIN=leadstorefront.com
API_ADDR=:8001
SESSION_SECRET=

API_IMAGE=josefdigital/platform:platform_api-v0.0.1
WEB_IMAGE=josefdigital/platform:platform_web-v0.0.1

```

## Access Remote Postgres Via Tunnel

After Terraform has created the DB droplet and `ansible/inventory.ini` has been generated, use the DB floating IP from the inventory:

```sh
ssh -f -N -i ~/.ssh/leadstorefront -o IdentitiesOnly=yes \
  -L 65432:127.0.0.1:5432 root@DB_FLOATING_IP
```

Then connect locally on port `65432`.

## Local Postgres

The existing local development database remains:

```sh
make docker-compose-local
```
