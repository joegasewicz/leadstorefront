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

WOODPECKER_HOST=https://ci.leadstorefront.com
WOODPECKER_ADMIN=joegasewicz
WOODPECKER_GITHUB=true
WOODPECKER_GITHUB_CLIENT=
WOODPECKER_GITHUB_SECRET=
WOODPECKER_AGENT_SECRET=
```

## Woodpecker CI

Woodpecker runs on the app droplet behind Caddy at `https://ci.leadstorefront.com`.
Create a GitHub OAuth app before starting the CI stack:

- Homepage URL: `https://ci.leadstorefront.com`
- Authorization callback URL: `https://ci.leadstorefront.com/authorize`

Set `WOODPECKER_GITHUB_CLIENT`, `WOODPECKER_GITHUB_SECRET`, and a random
`WOODPECKER_AGENT_SECRET` in the root `.env` file used by deployment. Generate
the agent secret with:

```sh
openssl rand -hex 32
```

After updating `/root/.env` on the app droplet, restart Woodpecker:

```sh
docker compose -f /root/docker-compose.app.yaml up -d woodpecker-server woodpecker-agent
docker logs --tail=100 root-woodpecker-server-1
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
