# Artsgoz Backend

Go API built with Fiber v3, Firebase Auth, and PostgreSQL.

## Run locally

Ensure credentials or environment variables (`GCP_PROJECT_ID`, `GCP_SECRET_ID`, or `firebase-service-account.json`) are configured.

```sh
go run cmd/main.go
```

The API listens on port `3000` by default. Override it with `PORT`.

## Database Migrations (PostgreSQL)

If using PostgreSQL migrations via pgx:

```sh
go run ./cmd/migrate up
```

## Checks

```sh
go test ./...
go vet ./...
```
