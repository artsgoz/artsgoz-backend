# Architecture

Artsgoz is a modular Go service organized as flat feature packages. The layout
matches the convention used by the Iconroof API: each business feature owns its
model, repository, service, HTTP handler, and routes.

## Layout

```text
cmd/
├── api/                  application composition root
└── migrate/              schema migration command
internal/
├── server/               Gin engine and HTTP middleware
├── user/
│   ├── model.go          GORM model and JSON representation
│   ├── repository.go     repository interface and GORM implementation
│   ├── service.go        application and business behavior
│   ├── handler.go        Gin request/response translation
│   ├── routes.go         route registration
│   └── service_test.go   service tests with repository stubs
└── platform/
	├── config/           environment and Secret Manager loading
	└── postgres/         GORM client and pool lifecycle
migrations/               ordered SQL migrations
```

## Feature flow

```text
Gin handler → Service → Repository interface ← GORM repository
```

All feature code shares one Go package, but responsibilities remain separated by
file. This favors discoverability and straightforward wiring over strict layer
isolation.

## Wiring

`cmd/api/main.go` is the composition root. It constructs each dependency
explicitly:

```go
repository := user.NewGormRepository(database.DB)
service := user.NewService(repository)
handler := user.NewHandler(service)
user.RegisterRoutes(api, handler)
```

Feature packages must not create database connections or read environment
variables. Those responsibilities stay in `internal/platform` and `cmd`.

## Request flow

For user registration:

1. The handler binds the HTTP request.
2. The service validates and normalizes input, hashes the password, and builds
   the user model.
3. The repository persists the model with GORM.
4. PostgreSQL uniqueness errors become `ErrEmailAlreadyExists`.
5. The handler maps expected errors to stable HTTP responses.
6. Gin's default middleware logs requests and recovers from panics.

The database unique constraint is authoritative; the service does not perform a
race-prone check-before-insert query.

## Conventions

- Keep one package per feature and split responsibilities by filename.
- Define repository interfaces next to their GORM implementations.
- Keep HTTP request structs private to `handler.go`.
- Keep SQL/GORM operations in `repository.go`.
- Keep validation and business behavior in `service.go`.
- Register routes in `routes.go`; wire dependencies in `cmd/api`.
- Wrap infrastructure failures with operation context using `%w`.
- Never expose password hashes in JSON.
- Add service tests using small repository stubs.
- Continue using SQL migrations; do not call GORM `AutoMigrate` at runtime.

## Configuration

Configuration is resolved in this order:

1. Process environment
2. `.env.local` for missing values
3. GCP Secret Manager when required `DB_*` fields are incomplete

The server uses `PORT` and defaults to `3000`. Gin uses `GIN_MODE` and defaults
to `release`.
