# Artsgoz Backend

Go API built with Gin.

## Run locally

Provide `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and optionally
`DB_SSL_MODE`. Alternatively, configure `GCP_PROJECT_ID`, `GCP_SECRET_ID`, and
`GCP_SECRET_VERSION` to load those fields from Secret Manager.

```sh
go run ./cmd/api
```

The API listens on port `3000` by default. Override it with `PORT`.

## Checks

```sh
go test ./...
go vet ./...
bash scripts/check-deps.sh
```

See [`docs/architecture.md`](docs/architecture.md) for package boundaries and conventions.
