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
│   ├── admin_service.go  profile business behavior
│   ├── admin_handler.go  Gin request/response translation
│   ├── routes.go         route registration
│   └── admin_service_test.go   service tests with repository stubs
└── platform/
	├── config/           environment and local dotenv loading
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
firebaseClient, err := auth.InitFirebase(cfg.FirebaseCredentials, cfg.FirebaseProjectID)
if err != nil { return err }
service := auth.NewService(repository, firebaseClient)
auth.Register(api, service)
user.Register(api, database.DB, auth.Middleware(service))
```

Feature packages must not create database connections or read environment
variables. Those responsibilities stay in `internal/platform` and `cmd`.

## Request flow

For Firebase login:

1. The frontend signs in with Firebase Email/Password or Google.
2. The handler binds the Firebase ID token and optional username.
3. The service verifies the token, including revocation, through Firebase Admin.
4. The service retrieves the profile by Firebase UID or safely links a legacy
   profile using a verified email; first sign-in creates a member profile.
5. PostgreSQL enforces identity uniqueness. Conditional updates prevent
   overwriting a Firebase UID during concurrent linking.
6. Protected routes verify Firebase tokens and load roles from PostgreSQL.

See [Firebase authentication](firebase-auth.md) for the frontend contract and
migration steps. Firebase owns credentials; PostgreSQL stores application profiles.

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

Configuration is read from the process environment after optionally loading
`.env.local`. Existing environment variables take precedence over the file.

The server uses `PORT` and defaults to `3000`. Gin uses `GIN_MODE` and defaults
to `release`.
