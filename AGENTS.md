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
- Web entry point: `cmd/platform_web/main.go`
- API entry point: `cmd/platform_api/main.go`
- Web routes: `platform_web/routes`
- API routes: `platform_api/routes`
- Web templates: `platform_web/templates`
- Web template partials: `platform_web/templates/partials`
- Web route templates: `platform_web/templates/routes`
- Web source assets: `platform_web/src`
- Web built assets: `platform_web/static/assets`
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

The compose service is `postgres-platformdb` and uses `infra/.env-dev`.

Expected local database env vars:

```sh
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=platformdb
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable
SESSION_SECRET=local-dev-session-secret-change-me
```

## Database Code

- Use `pkgs.Config` for runtime configuration. It is initialized once with `pkgs.Load()`.
- `pkgs/utils/database.NewPostgres()` creates the GORM Postgres connection.
- `pkgs/utils/database.NewMigrate(db)` creates a migration runner.
- `Migrate.Run()` auto-migrates the `Country`, `ProductCategory`, `Product`, `ArticleCategory`, `Article`, `Role`, and `User` models, then calls `Seed(db)`.
- `pkgs/utils/database/seeds.go` seeds supported countries, roles (`admin`, `editor`, `user`), a local admin user, article categories, and broad tech product categories.
- `cmd/platform_api/main.go` creates the Postgres connection, logs connection success, logs migration start, and runs migrations.
- All database access must stay inside the API service. The web service must not import `pkgs/utils/database`, accept `*gorm.DB`, call GORM query methods, or otherwise connect/query Postgres directly.
- Web handlers must call the API over HTTP for application data and admin mutations, then marshal/unmarshal JSON into the structs needed by templates.
- Always use `github.com/joegasewicz/identity-client` v0.4.0 or newer where possible for web-to-API JSON requests, including `Get`, `Post`, `Put`, and `Delete`.
- Always use `github.com/joegasewicz/multipart-requests` where possible for web-to-API file uploads.
- Always use `github.com/joegasewicz/form-validator` where possible for incoming web/admin form validation and typed form value extraction. API handlers that accept multipart form submissions should also validate them with `form-validator`.
- Always use `github.com/joegasewicz/entity-file-uploader` where possible for API-owned file persistence and retrieval.
- When a dependency, configuration, or repeated value is used across multiple route methods, prefer storing it as a struct member instead of recreating a local variable in every method.
- API routes must be versioned with `pkgs/utils.GetVersion`, which prepends `/api/v1`. For example: `app.GET(utils.GetVersion("/health"), health)`.
- `platform_api/routes.Register(app, db)` must declare routes directly. Do not hide route declarations behind `registerX(app, db)` helper functions.
- When creating a new API request handler, always create or use `platform_api/routes/name-of-model.go`. In that file, define a struct named after the model, then implement four pointer receiver methods on that struct: `Get`, `Post`, `Put`, and `Delete`. This is the required pattern for every model route.
- Web route files must follow the same object pattern: each `platform_web/routes/name-of-route.go` file creates a struct named after the route, and that struct implements pointer receiver methods `Get`, `Post`, `Put`, and `Delete`. Web methods must call the API for every operation involving database-backed data.
- Keep route files REST-oriented and table-backed. API route files should represent database models/tables, not random feature names or generic function buckets.
- Do not put cross-cutting business logic in route files. Shared concerns such as pagination, route versioning, URL parsing, and common route helpers belong in `pkgs/utils/routes.go` unless they are model-specific behavior inside that model's route file.

## HTTP Apps

- `cmd/platform_web/main.go` starts the Gin web app on `WEB_ADDR`, default `:8000`.
- `cmd/platform_api/main.go` starts the Gin JSON API on `API_ADDR`, default `:8001`.
- Web routes are registered in `platform_web/routes.Register(app)`.
- Web templates are loaded in `cmd/platform_web/main.go`.
- API routes are registered in `platform_api/routes.Register(app, db)`.
- Web exposes `GET /health`; API exposes `GET /api/v1/health` through `pkgs/utils.GetVersion("/health")`.
- Admin web routes live under `/admin`, not country paths.
- Admin routes:
  - `GET /admin/login`
  - `POST /admin/login`
  - `GET /admin/register`
  - `POST /admin/register`
  - `GET /admin`
  - `GET /admin/articles`
  - `GET /admin/articles/create`
  - `POST /admin/articles/create`
  - `GET /admin/articles/:id/edit`
  - `POST /admin/articles/:id/edit`
  - `POST /admin/articles/:id/delete`
- `/admin` requires a server-side session and a user role of `admin` or `editor`.
- Admin article delete uses a POST form and shows a flash message after redirect.
- The web root route `/` redirects to the detected country homepage such as `/uk` or `/us`.
- Country homepages are served by `GET /:country`.
- API routes should reflect the web route surface under `/api/v1`, for example web `GET /uk/products` maps to API `GET /api/v1/uk/products`.
- Location detection lives in `pkgs/middleware/locations.go`. It checks common CDN/proxy country headers first (`CF-IPCountry`, `X-Vercel-IP-Country`, `X-Country-Code`, `X-AppEngine-Country`), then `Accept-Language`, then defaults to `uk`.
- The web frontend should use Tailwind and call the API for application data. The API is the only HTTP app allowed to connect to the database.

## Middleware

- Prefer middleware for cross-cutting request behavior such as sessions, authentication, authorization, localization, redirects, and request context.
- Use utility functions for non-request-specific helpers.
- Server-side web sessions use `github.com/gin-contrib/sessions` with the memstore backend in `pkgs/middleware/sessions.go`.
- Role-based admin authorization lives in `pkgs/middleware/admin_auth.go`.

## Frontend Assets

- Web custom TypeScript and Sass live in `platform_web/src`.
- Vite builds `platform_web/src/main.ts` to `platform_web/static/assets/app.js` and `platform_web/static/assets/app.css`.
- Build frontend assets from `platform_web` with `npm run build`.
- `cmd/platform_web/main.go` serves built assets at `/assets`.
- `platform_web/templates/base.gohtml` renders `/assets/app.css` in the head.
- `platform_web/templates/partials/scripts.gohtml` renders `/assets/app.js` in the footer.

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
- `ArticleCategory`
  - `gorm.Model`
  - `Name string`
- `Article`
  - `gorm.Model`
  - `Author string`
  - `Title string`
  - `Slug string`
  - `Subtitle string`
  - `Body string`
  - `ImageURL string`
  - `MetaTitle string`
  - `MetaDescription string`
  - `MetaKeywords string`
  - `CanonicalURL string`
  - `IsPublished bool`
  - `PublishedAt *time.Time`
  - `ArticleCategoryID uint`
  - `ArticleCategory ArticleCategory`
  - `ProductID *uint`
  - `Product *Product`
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
go test -c ./cmd/platform_web -o /tmp/gadgetscout-platform_web.test
go test -c ./cmd/platform_api -o /tmp/gadgetscout-platform_api.test
```

`go test ./cmd/platform_api` may execute database startup and attempt to connect to Postgres. Prefer `go test -c` for a compile-only check unless the database is intentionally running.
