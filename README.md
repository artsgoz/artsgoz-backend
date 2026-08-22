# Artsgoz Backend

Go API built with Fiber.

## Run locally

Provide `DATABASE_URL` directly, or configure `GCP_PROJECT_ID`, `GCP_SECRET_ID`,
and optionally `GCP_SECRET_VERSION` in `.env.local` to load it from Secret Manager.

```sh
go run ./cmd/api
```

The API listens on `:3000` by default. Override it with `HTTP_ADDRESS`.

## Checks

```sh
go test ./...
go vet ./...
bash scripts/check-deps.sh
```

See [`docs/architecture.md`](docs/architecture.md) for package boundaries and conventions.
