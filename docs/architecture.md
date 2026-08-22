# Architecture

Artsgoz is a modular Go service organized by business feature. It uses Clean
Architecture dependency direction without requiring tactical DDD patterns.

## Layout

```text
cmd/
├── api/                 API composition root
└── migrate/             schema migration command
internal/
├── server/              Fiber setup and cross-cutting HTTP middleware
├── user/
│   ├── domain/          business model and invariants
│   ├── usecase/         application operations and consumer-owned ports
│   ├── controller/      HTTP request/response translation
│   ├── repository/      PostgreSQL adapter
│   └── module.go        private feature wiring and public route registration
└── platform/
    ├── config/          configuration and secret loading
    ├── logging/         process logger
    └── postgres/        connection pool
migrations/              ordered database migrations
```

## Dependency rule

Dependencies point toward business behavior:

```text
controller ──> usecase ──> domain
repository ──> usecase ──> domain
module/bootstrap ──> concrete packages for wiring
```

- `domain` imports only the standard library. It does not know HTTP, SQL,
  configuration, logging, hashing libraries, or application error codes.
- `usecase` owns the interfaces it consumes. It does not import Fiber, pgx, a
  controller, or a repository implementation.
- `controller` translates HTTP DTOs and application errors. HTTP models do not
  become domain models.
- `repository` translates database behavior into the use-case contract.
- `module.go` is the feature composition boundary. Its public API exposes route
  registration, not handlers or repositories.
- `cmd` and `module.go` may depend on concrete implementations because their
  responsibility is dependency wiring.

## Request flow

Registration follows this path:

1. The controller binds the JSON request.
2. The use case validates and normalizes application input.
3. Injected services hash the password and provide the ID and current time.
4. The domain constructor creates a valid user.
5. The PostgreSQL adapter inserts it and translates a unique violation into
   `usecase.ErrEmailAlreadyExists`.
6. The controller maps known application errors to stable HTTP responses.
7. The server error handler logs unexpected failures and returns a safe 500.

The database unique constraint is authoritative. There is no check-then-insert
query because it cannot prevent concurrent registrations.

## Conventions

- Package names are short, singular, and describe responsibility.
- Prefer one file per use case once a package contains multiple operations.
- Keep interfaces small and declare them where they are consumed.
- Do not introduce an interface solely to mirror a concrete type.
- Avoid generic `utils` packages; place code with the capability that owns it.
- Wrap unexpected errors with operation context using `%w`.
- Keep public error messages separate from logged internal errors.
- Introduce transaction ports only for use cases that need atomic multi-step work.
- New behavior requires use-case tests. Add adapter integration tests for SQL.

## Configuration and lifecycle

Configuration is resolved in this order:

1. Process environment
2. `.env.local` for missing values
3. GCP Secret Manager when `DATABASE_URL` is still missing

The API validates configuration before opening PostgreSQL, uses a bounded startup
context, emits structured JSON logs, and shuts down on `SIGINT` or `SIGTERM`.
