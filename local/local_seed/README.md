# Local API Seed

Creates 30 products and 30 articles by calling the API only.

Prerequisites:

```sh
cd infra
make docker-compose-local
go run ./cmd/api
```

Run:

```sh
python3 local/local_seed/main.py
```

Optional environment variables:

```sh
GADGETSCOUT_API_URL=http://localhost:8001
GADGETSCOUT_API_PREFIX=/api/v1
GADGETSCOUT_SEED_COUNTRY=uk
GADGETSCOUT_SEED_RUN_ID=myrun
```

Article images are uploaded from `local/local_seed/images` through the API upload endpoint.
