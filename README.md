# Artsgoz Backend

Go API built with Fiber v3, Firebase Auth, and PostgreSQL.

## Run locally

Provide `DB_HOST`, `DB_USER`, and `DB_NAME`. `DB_PORT` defaults to `5432` and `DB_SSL_MODE` defaults to `disable`. `DB_PASSWORD` may be empty.

```sh
go run ./cmd/main.go
```

The API listens on port `3000` by default. Override it with `PORT`.

## Checks

```sh
go test ./...
go vet ./...
```

See [docs/architecture.md](https://github.com/artsgoz/artsgoz-backend/blob/main/docs/architecture.md) for package boundaries and conventions.
