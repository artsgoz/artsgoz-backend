# Artsgoz Backend

Go API built with Gin, GORM, and PostgreSQL.

## Run locally

Provide `DB_HOST`, `DB_USER`, and `DB_NAME`. `DB_PORT` defaults to `5432` and
`DB_SSL_MODE` defaults to `disable`. `DB_PASSWORD` may be empty.

```sh
go run ./cmd/api
```

The API listens on port `3000` by default. Override it with `PORT`.
Gin runs in `release` mode by default; use `GIN_MODE=debug` locally when needed.

## Checks

```sh
go test ./...
go vet ./...
bash scripts/check-deps.sh
```

See [`docs/architecture.md`](docs/architecture.md) for package boundaries and conventions.
