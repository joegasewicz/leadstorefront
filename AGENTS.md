# Gadget Scout Agent Context

## Product Purpose

Gadget Scout lists the best gadget deals in the world. The product should help English-speaking shoppers discover strong-value deals on consumer technology and compare them across supported countries.

Initial market coverage:

- United States
- United Kingdom
- Australia
- South Africa
- New Zealand
- Canada
- Ireland
- Singapore

When adding product, content, pricing, scraping, or localization features, keep country-specific retailers, currencies, availability, shipping, and deal quality in mind. The web frontend should be clear, fast, and deal-focused rather than a marketing landing page.

Country-specific pages use lowercase two-letter URL codes:

- `/us`
- `/uk`
- `/au`
- `/za`
- `/nz`
- `/ca`
- `/ie`
- `/sg`

The default country is `uk`. If the visitor country cannot be detected, redirect to `/uk`.

## Project Layout

- Go module: `gadgetscout`
- Web entry point: `cmd/web/main.go`
- API entry point: `cmd/api/main.go`
- Web routes: `web/routes`
- API routes: `api/routes`
- Web templates: `web/templates`
- Web template partials: `web/templates/partials`
- Web route templates: `web/templates/routes`
- Runtime config: `pkgs/config.go`
- Shared models: `pkgs/models`
- Shared middleware: `pkgs/middleware`
- Shared utilities: `pkgs/utils`
- Database utilities: `pkgs/utils/database`
- Local infrastructure: `infra`
- Flutter mobile app: `mobile`

## Local Postgres

Local Postgres is managed from the `infra` directory.

```sh
cd infra
make docker-compose-local
```

The compose service is `postgres-gadgetscout` and uses `infra/.env-dev`.

Expected local database env vars:

```sh
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=gadgetscout
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable
```

## Database Code

- Use `pkgs.Config` for runtime configuration. It is initialized once with `pkgs.Load()`.
- `pkgs/utils/database.NewPostgres()` creates the GORM Postgres connection.
- `pkgs/utils/database.NewMigrate(db)` creates a migration runner.
- `Migrate.Run()` auto-migrates the `Country`, `ProductCategory`, `Product`, `Role`, and `User` models, then calls `Seed(db)`.
- `pkgs/utils/database/seeds.go` seeds supported countries.
- `cmd/api/main.go` creates the Postgres connection, logs connection success, logs migration start, and runs migrations.

## HTTP Apps

- `cmd/web/main.go` starts the Gin web app on `WEB_ADDR`, default `:8000`.
- `cmd/api/main.go` starts the Gin JSON API on `API_ADDR`, default `:8001`.
- Web routes are registered in `web/routes.Register(app)`.
- Web templates are loaded in `cmd/web/main.go`.
- API routes are registered in `api/routes.Register(app)`.
- Both services expose `GET /health`.
- The web root route `/` redirects to the detected country homepage such as `/uk` or `/us`.
- Country homepages are served by `GET /:country`.
- Location detection lives in `pkgs/middleware/locations.go`. It checks common CDN/proxy country headers first (`CF-IPCountry`, `X-Vercel-IP-Country`, `X-Country-Code`, `X-AppEngine-Country`), then `Accept-Language`, then defaults to `uk`.
- The web frontend should use Tailwind and call the API for application data.

## Model Conventions

Models live in `pkgs/models` and must always embed `gorm.Model`.

Current models:

- `Country`
  - `gorm.Model`
  - `Code string`
  - `Name string`
  - `Currency string`
- `ProductCategory`
  - `gorm.Model`
  - `Name string`
- `Product`
  - `gorm.Model`
  - `Name string`
  - `Slug string`
  - `Description string`
  - `Brand string`
  - `ModelNumber string`
  - `ImageURL string`
  - `ProductURL string`
  - `AffiliateURL string`
  - `RetailerName string`
  - `RetailerURL string`
  - `Source string`
  - `ExternalID string`
  - `Currency string`
  - `CurrentPriceCents int64`
  - `OriginalPriceCents int64`
  - `ShippingCostCents *int64`
  - `DiscountPercent int`
  - `CouponCode string`
  - `DealScore int`
  - `Rating float32`
  - `ReviewCount int`
  - `IsAvailable bool`
  - `IsFeatured bool`
  - `StartsAt *time.Time`
  - `EndsAt *time.Time`
  - `LastCheckedAt *time.Time`
  - `CountryID uint`
  - `Country Country`
  - `CategoryID uint`
  - `Category ProductCategory`
- `Role`
  - `gorm.Model`
  - `Name string`
- `User`
  - `gorm.Model`
  - `Email string`
  - `Password string`
  - `RoleID uint`
  - `Role Role`

## Verification

Useful compile checks:

```sh
go test ./pkgs/models
go test ./pkgs/utils/database
go test -c ./cmd/web -o /tmp/gadgetscout-web.test
go test -c ./cmd/api -o /tmp/gadgetscout-api.test
```

`go test ./cmd/api` may execute database startup and attempt to connect to Postgres. Prefer `go test -c` for a compile-only check unless the database is intentionally running.
