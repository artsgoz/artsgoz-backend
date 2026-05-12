# artsgoz-backend — Architecture Guide

A Go (Fiber + pgx) backend organized using **Domain-Driven Design (DDD)** + **Clean Architecture**. This document explains the *why*, the *what*, and the *how* — so a new contributor can land their first PR confidently.

---

## Table of contents

1. [DDD & Clean Architecture in 5 minutes](#1-ddd--clean-architecture-in-5-minutes)
2. [Project layout](#2-project-layout)
3. [The four layers](#3-the-four-layers)
4. [Dependency rule (the one rule)](#4-dependency-rule-the-one-rule)
5. [Request lifecycle](#5-request-lifecycle)
6. [How to add a new feature (step by step)](#6-how-to-add-a-new-feature-step-by-step)
7. [How to add a new bounded context (module)](#7-how-to-add-a-new-bounded-context-module)
8. [Conventions & rules](#8-conventions--rules)
9. [Local development setup](#9-local-development-setup)
10. [FAQ](#10-faq)

---

## 1. DDD & Clean Architecture in 5 minutes

### What problem are we solving?

When code grows, it tends to tangle. Business rules get mixed with HTTP parsing, SQL queries leak into handlers, and you can't change the database without rewriting half the app. **DDD** and **Clean Architecture** are two complementary ideas that prevent that.

### Domain-Driven Design (DDD)

DDD says: **model the software around the language and rules of the business**, not around the database tables.

Key DDD vocabulary you'll see in this repo:

| Term | Meaning | Example in this repo |
|---|---|---|
| **Bounded Context** | A self-contained part of the business with its own language and rules | `user`, `artwork`, `order` |
| **Entity** | A thing with an identity (an `id`) that changes over time | `User`, `Artwork` |
| **Aggregate Root** | The entity that owns and protects a cluster of related data | `User` owns its password, email, profile |
| **Value Object** | A thing defined by its *value*, not identity (immutable) | `Money`, `EmailAddress` (if we add them) |
| **Repository** | An interface for loading/saving aggregates — *defined in the domain*, *implemented in infrastructure* | `UserRepository` |
| **Use Case / Application Service** | A single business operation that orchestrates entities and repositories | `RegisterUser` |
| **Domain Event** | A fact that happened in the domain (often emitted by aggregates) | `UserRegistered` (future) |

### Clean Architecture

Clean Architecture says: **dependencies must point inward, toward the business rules.**

```
┌──────────────────────────────────────────────────────────┐
│  Interface layer (HTTP, gRPC, CLI)        ← outer-most   │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Application layer (use cases, DTOs)               │  │
│  │  ┌──────────────────────────────────────────────┐  │  │
│  │  │  Domain layer (entities, repo interfaces)    │  │  │
│  │  │              ← inner-most, knows nothing     │  │  │
│  │  │                else                          │  │  │
│  │  └──────────────────────────────────────────────┘  │  │
│  │            ↑ implements                            │  │
│  │  Infrastructure layer (Postgres, Redis, GCS, ...)  │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

**The domain has zero imports from outer layers.** Infrastructure depends *on the domain* (to implement its interfaces), never the other way around. This is called **dependency inversion** and it's why you can swap Postgres for Firestore without touching business logic.

### Why bother?

- **Testability**: business logic runs without spinning up Postgres or Fiber.
- **Swappability**: change DB, swap Fiber for Echo, add a CLI — none of it touches `domain/` or `application/`.
- **Clarity for new contributors**: each folder has *one* responsibility. You always know where new code goes.
- **Scales with the team**: each bounded context can be owned by one team end-to-end.

---

## 2. Project layout

```
artsgoz-backend/
├── cmd/
│   └── main.go                     # Composition root — wires everything together
├── config/
│   └── config.go                   # Loads env from GCP Secret Manager + .env.local
├── docs/
│   └── architecture.md             # ← you are here
├── internal/                       # All app code lives under internal/ so it
│   │                               #   can't be imported by external repos
│   ├── platform/                   # Shared kernel — reusable across modules
│   │   ├── apperr/                 # Typed application errors + Fiber error handler
│   │   ├── httpserver/             # Fiber app factory pre-wired with apperr
│   │   ├── postgres/               # pgxpool factory
│   │   └── validator/              # go-playground/validator singleton
│   │
│   └── modules/                    # Bounded contexts — one folder per business area
│       └── user/                   # Example bounded context: user
│           ├── domain/             # ① Entities, value objects, repository interfaces
│           │   ├── user.go
│           │   └── user_repository.go
│           ├── application/        # ② Use cases + input/output DTOs
│           │   ├── dto.go
│           │   └── register_user.go
│           ├── infrastructure/     # ③ Implementations of domain interfaces
│           │   └── persistence/
│           │       └── postgres/
│           │           └── user_repository.go
│           ├── interface/          # ④ HTTP / gRPC / CLI entry points
│           │   └── rest/
│           │       ├── handler.go
│           │       └── routes.go
│           └── module.go           # Wiring helper consumed by cmd/main.go
│
├── migrations/                     # SQL migration files
│   ├── 0001_create_users.up.sql
│   └── 0001_create_users.down.sql
│
├── pkg/                            # Reusable, project-agnostic utilities
│   └── utils/
│       └── secretmanager.go        # Thin GCP Secret Manager client
│
├── test/                           # End-to-end & integration tests (future)
├── go.mod
└── go.sum
```

### Why `internal/`?

Go enforces that anything under `internal/` cannot be imported by code outside the parent module. This makes accidental coupling impossible — perfect for application code that should never leak out.

### Why `pkg/` separate from `internal/`?

`pkg/` is for code that *could* legitimately be reused by another repo (e.g. our GCP Secret Manager wrapper). Almost nothing belongs here. When in doubt, put it under `internal/`.

---

## 3. The four layers

### ① Domain layer — `internal/modules/<ctx>/domain/`

**The heart of the business. Imports nothing app-specific (no Fiber, no pgx, no JSON tags).**

Contains:
- **Entities** as Go structs with constructor functions that enforce invariants.
- **Repository interfaces** (`UserRepository`) — *what* the domain needs from persistence, not *how*.
- (Future) Value objects, domain events, domain services.

Example — `internal/modules/user/domain/user.go`:

```go
type User struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
}

func NewUser(email, plaintextPassword string) (*User, error) {
    if email == "" {
        return nil, apperr.Validation("email is required", nil)
    }
    if len(plaintextPassword) < 8 {
        return nil, apperr.Validation("password must be at least 8 characters", nil)
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
    if err != nil {
        return nil, apperr.Internal("hash password", err)
    }
    return &User{
        ID:           uuid.NewString(),
        Email:        email,
        PasswordHash: string(hash),
        CreatedAt:    time.Now().UTC(),
    }, nil
}
```

> The constructor enforces invariants. Once `NewUser` returns a `*User`, it is guaranteed valid. Never construct entities by hand outside this layer.

Repository interface — `internal/modules/user/domain/user_repository.go`:

```go
type UserRepository interface {
    Save(ctx context.Context, u *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
}
```

### ② Application layer — `internal/modules/<ctx>/application/`

**Orchestrates use cases. One file per use case. Knows about the domain but nothing about HTTP or SQL.**

A use case = one business operation. It does:
1. Validate input (DTO with `validate:` tags).
2. Load aggregates via repository interfaces.
3. Apply domain logic (call entity methods, build new entities).
4. Persist via repositories.
5. Return an output DTO.

Example — `internal/modules/user/application/register_user.go`:

```go
type RegisterUser struct {
    users domain.UserRepository
}

func NewRegisterUser(users domain.UserRepository) *RegisterUser {
    return &RegisterUser{users: users}
}

func (uc *RegisterUser) Execute(ctx context.Context, in RegisterUserInput) (RegisterUserOutput, error) {
    if err := validator.Struct(in); err != nil {
        return RegisterUserOutput{}, err
    }
    existing, err := uc.users.FindByEmail(ctx, in.Email)
    if err != nil {
        var appErr *apperr.Error
        if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
            return RegisterUserOutput{}, err
        }
    }
    if existing != nil {
        return RegisterUserOutput{}, apperr.Conflict("email already registered")
    }
    user, err := domain.NewUser(in.Email, in.Password)
    if err != nil {
        return RegisterUserOutput{}, err
    }
    if err := uc.users.Save(ctx, user); err != nil {
        return RegisterUserOutput{}, err
    }
    return RegisterUserOutput{ID: user.ID, Email: user.Email}, nil
}
```

### ③ Infrastructure layer — `internal/modules/<ctx>/infrastructure/`

**Implements the interfaces declared in the domain. This is where pgx, HTTP clients, GCS SDK, etc. live.**

Example — `internal/modules/user/infrastructure/persistence/postgres/user_repository.go`:

```go
type userRepository struct {
    pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
    return &userRepository{pool: pool}
}

func (r *userRepository) Save(ctx context.Context, u *domain.User) error {
    const q = `INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)`
    if _, err := r.pool.Exec(ctx, q, u.ID, u.Email, u.PasswordHash, u.CreatedAt); err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return apperr.Conflict("email already registered")
        }
        return apperr.Internal("save user", err)
    }
    return nil
}
```

> The constructor returns the *interface type* (`domain.UserRepository`), not the concrete struct. This is enforced dependency inversion.

### ④ Interface layer — `internal/modules/<ctx>/interface/`

**HTTP handlers, gRPC servers, CLI commands. Translates outside-world calls into use-case invocations.**

Example — `internal/modules/user/interface/rest/handler.go`:

```go
type UserHandler struct {
    register *application.RegisterUser
}

func NewUserHandler(register *application.RegisterUser) *UserHandler {
    return &UserHandler{register: register}
}

func (h *UserHandler) Register(c fiber.Ctx) error {
    var in application.RegisterUserInput
    if err := c.Bind().Body(&in); err != nil {
        return apperr.Validation("invalid request body", err.Error())
    }
    out, err := h.register.Execute(c.Context(), in)
    if err != nil {
        return err // central error handler maps *apperr.Error → status code + JSON
    }
    return c.Status(fiber.StatusCreated).JSON(out)
}
```

Routes — `internal/modules/user/interface/rest/routes.go`:

```go
func RegisterRoutes(r fiber.Router, h *UserHandler) {
    g := r.Group("/users")
    g.Post("/register", h.Register)
}
```

### Module wiring — `internal/modules/<ctx>/module.go`

A tiny helper that builds all four layers and exposes only what `cmd/main.go` needs:

```go
type Module struct {
    Handler *rest.UserHandler
}

func New(pool *pgxpool.Pool) *Module {
    repo := pgrepo.NewUserRepository(pool)
    registerUC := application.NewRegisterUser(repo)
    handler := rest.NewUserHandler(registerUC)
    return &Module{Handler: handler}
}
```

### Composition root — `cmd/main.go`

The *only* place where all layers meet. Reads config, builds resources (DB pool), constructs modules, mounts routes, starts the server.

```go
func main() {
    ctx := context.Background()
    config.LoadFromSecretManager(ctx)

    pool, _ := postgres.New(ctx, os.Getenv("DATABASE_URL"))
    defer pool.Close()

    userMod := user.New(pool)

    app := httpserver.New()
    api := app.Group("/api/v1")
    userrest.RegisterRoutes(api, userMod.Handler)

    app.Listen(":3000")
}
```

---

## 4. Dependency rule (the one rule)

**Inner layers must not import outer layers.**

| Layer | Can import from |
|---|---|
| `domain/` | Standard library + `internal/platform/apperr` only |
| `application/` | `domain/` + `internal/platform/{apperr,validator}` |
| `infrastructure/` | `domain/` + drivers (pgx, GCS SDK, etc.) + `internal/platform/apperr` |
| `interface/` | `application/` + `internal/platform/*` + Fiber |
| `cmd/` | Everything (it's the composition root) |

You can verify this with the lint script:

```bash
./scripts/check-deps.sh
```

It greps every file under `internal/modules/*/domain/` and `internal/modules/*/application/` for forbidden imports (pgx, fiber, validator, sibling layers) and prints any violations with `file:line`. Exits non-zero on violations, so it slots into CI or a pre-commit hook directly.

Wire it into CI (GitHub Actions sketch):

```yaml
- name: Check Clean Architecture dependency rule
  run: ./scripts/check-deps.sh
```

Or as a pre-commit hook (`.git/hooks/pre-commit`):

```bash
#!/usr/bin/env bash
./scripts/check-deps.sh || exit 1
```

---

## 5. Request lifecycle

Walk through what happens for `POST /api/v1/users/register`:

```
1.  HTTP request arrives at Fiber.
       ↓
2.  Fiber routes it to UserHandler.Register
    (registered via userrest.RegisterRoutes in main.go).
       ↓
3.  Handler parses JSON body into application.RegisterUserInput.
    On parse error → returns apperr.Validation(...).
       ↓
4.  Handler calls registerUC.Execute(ctx, input).
       ↓
5.  Use case validates input via validator.Struct(input).
    On invalid → returns apperr.Validation(...).
       ↓
6.  Use case calls users.FindByEmail(ctx, email).
       ↓
7.  Postgres repository runs SELECT. If found → use case returns
    apperr.Conflict. If not found → continues.
       ↓
8.  Use case calls domain.NewUser(email, password).
    Domain enforces invariants, hashes password with bcrypt.
       ↓
9.  Use case calls users.Save(ctx, user).
       ↓
10. Repository runs INSERT. On unique violation → apperr.Conflict.
       ↓
11. Use case returns RegisterUserOutput{ID, Email}.
       ↓
12. Handler returns 201 + JSON.
       ↓
13. If any step returned *apperr.Error, Fiber's central ErrorHandler
    (apperr.FiberErrorHandler) maps it to status + JSON body:
       - validation  → 400
       - conflict    → 409
       - not_found   → 404
       - unauthorized → 401
       - internal    → 500
```

---

## 6. How to add a new feature (step by step)

Suppose you want to add **"login"** to the existing `user` module.

### Step 1 — Plan the use case

Write down, in plain English, the rule:

> Given an email + password, find the user, verify the password, and return a session token. On bad credentials, return 401.

### Step 2 — Decide what changes per layer

| Layer | Change |
|---|---|
| `domain/` | Maybe add `User.CheckPassword(plaintext) bool` if not present. Maybe add new repository method (none needed here — `FindByEmail` already exists). |
| `application/` | Add `LoginUser` use case + `LoginUserInput` / `LoginUserOutput` DTOs. |
| `infrastructure/` | Add a `TokenIssuer` adapter (JWT) if not present, behind a domain interface. |
| `interface/rest/` | Add `Login` handler + route `POST /users/login`. |
| `module.go` | Wire the new use case. |

### Step 3 — Write the code (in order: inside-out)

1. **Domain first.** Add `CheckPassword` on `User`:
   ```go
   func (u *User) CheckPassword(plaintext string) bool {
       return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintext)) == nil
   }
   ```
2. **Application.** Create `internal/modules/user/application/login_user.go`:
   ```go
   type LoginUser struct {
       users  domain.UserRepository
       tokens domain.TokenIssuer // interface declared in domain
   }
   func NewLoginUser(u domain.UserRepository, t domain.TokenIssuer) *LoginUser { ... }
   func (uc *LoginUser) Execute(ctx, in LoginUserInput) (LoginUserOutput, error) {
       // validate input
       // load user by email — if not found → apperr.Unauthorized (don't reveal which)
       // user.CheckPassword(in.Password) — if false → apperr.Unauthorized
       // tokens.Issue(user.ID) → return token in output
   }
   ```
3. **Infrastructure.** Implement `TokenIssuer` in `internal/modules/user/infrastructure/jwt/` (or wherever).
4. **Interface.** Add `Login` to `handler.go` and a route in `routes.go`.
5. **Module wiring.** Update `module.go`:
   ```go
   type Module struct {
       Handler *rest.UserHandler
   }
   func New(pool *pgxpool.Pool, jwtSecret string) *Module {
       repo := pgrepo.NewUserRepository(pool)
       tokens := jwtadapter.NewIssuer(jwtSecret)
       registerUC := application.NewRegisterUser(repo)
       loginUC := application.NewLoginUser(repo, tokens)
       handler := rest.NewUserHandler(registerUC, loginUC)
       return &Module{Handler: handler}
   }
   ```
6. **Composition root.** Update `cmd/main.go` to pass any new config (e.g. `os.Getenv("JWT_SECRET")`).

### Step 4 — Test it

1. Unit test the use case with a fake `UserRepository`:
   ```go
   // internal/modules/user/application/login_user_test.go
   type fakeRepo struct{ user *domain.User }
   func (f *fakeRepo) FindByEmail(...) (*domain.User, error) { return f.user, nil }
   func (f *fakeRepo) Save(...) error { return nil }

   func TestLoginUser_Success(t *testing.T) { /* ... */ }
   func TestLoginUser_WrongPassword(t *testing.T) { /* ... */ }
   ```
2. Integration test (later) with a real Postgres in `test/`.

### Step 5 — Run & verify

```bash
go build ./...
go vet ./...
go run ./cmd
curl -X POST localhost:3000/api/v1/users/login -d '{"email":"a@b.com","password":"hunter2x"}'
```

---

## 7. How to add a new bounded context (module)

When you have a *new business area* (artwork, order, payment, …), create a new module.

### Step-by-step

1. **Pick a name.** Should be a noun from the business language (`artwork`, `order`, not `data` or `service`).
2. **Scaffold folders:**
   ```bash
   CTX=artwork
   mkdir -p internal/modules/$CTX/{domain,application,infrastructure/persistence/postgres,interface/rest}
   ```
3. **Define the aggregate root** in `internal/modules/<ctx>/domain/<ctx>.go` with a `NewXxx(...)` constructor.
4. **Define the repository interface** in `internal/modules/<ctx>/domain/<ctx>_repository.go`.
5. **Implement the repository** in `internal/modules/<ctx>/infrastructure/persistence/postgres/<ctx>_repository.go`.
6. **Write your first use case** in `internal/modules/<ctx>/application/<verb>_<ctx>.go` (e.g. `create_artwork.go`).
7. **Define DTOs** in `internal/modules/<ctx>/application/dto.go`.
8. **Add the handler + route** in `internal/modules/<ctx>/interface/rest/{handler.go,routes.go}`.
9. **Write the module wiring** in `internal/modules/<ctx>/module.go`.
10. **Add a migration** in `migrations/000N_create_<ctx>s.up.sql` (+ `.down.sql`).
11. **Mount it in `cmd/main.go`:**
    ```go
    artworkMod := artwork.New(pool)
    artworkrest.RegisterRoutes(api, artworkMod.Handler)
    ```
12. **Build & verify:**
    ```bash
    go build ./...
    go vet ./...
    psql "$DATABASE_URL" -f migrations/000N_create_artworks.up.sql
    go run ./cmd
    ```

> Tip: copy `internal/modules/user/` as a template and rename everything.

---

## 8. Conventions & rules

### Naming

- **Packages**: lowercase, single word. `domain`, `application`, `rest`. The folder `interface/rest/` has package name `rest`.
- **Files**: snake_case. `register_user.go`, `user_repository.go`.
- **Types**: PascalCase. `RegisterUser`, `UserRepository`.
- **Use cases**: one type per file, named as a verb phrase: `RegisterUser`, `LoginUser`, `ArchiveArtwork`. The single method is always `Execute`.
- **Repository methods**: `Save`, `FindByX`, `DeleteByID`. Return `(*Entity, error)` or `error`.

### Errors

- **Always return `*apperr.Error`** for known business cases. Never `errors.New(...)` in user-facing paths.
- **Wrap unknown errors** with `apperr.Internal("context", err)` at the boundary where they're caught.
- **Don't leak driver errors** (`pgx.ErrNoRows`, `pgconn.PgError`) past the infrastructure layer.

### Validation

- Input DTOs use `validate:` tags from `go-playground/validator`.
- Validation is performed at the *start* of every use case via `validator.Struct(in)`.
- Domain constructors enforce invariants again — defense in depth.

### Dependency injection

- **No DI framework.** Manual constructor injection in `module.go` and `cmd/main.go`.
- Always inject **interfaces**, never concrete types, into use cases.
- Constructors are named `New<Type>(deps...)`.

### SQL & migrations

- Migrations live in `migrations/` with `golang-migrate` naming: `NNNN_description.{up,down}.sql`.
- Run them with `go run ./cmd/migrate up` (see [§9.1 Migrations](#91-migrations)).
- Use positional placeholders `$1, $2, ...` (pgx).
- Repository unique-violation handling: check for SQLSTATE `23505` and return `apperr.Conflict`.

### Testing

- Use cases: unit-tested with fake repository implementations (no DB).
- Repositories: integration-tested against a real Postgres (in `test/` or `_test.go` alongside the impl).
- HTTP handlers: rarely need their own tests if use cases are well-covered.

### What NOT to do

- ❌ Don't import `pgx` or `fiber` inside `domain/` or `application/`.
- ❌ Don't put SQL in the use case. SQL lives in the repository implementation.
- ❌ Don't parse JSON in the use case. JSON lives in the HTTP handler.
- ❌ Don't add a "manager" or "service" struct that does HTTP + SQL + business logic. That's the anti-pattern this whole structure prevents.
- ❌ Don't share entities across modules. If `order` needs user data, expose a minimal read model or call the `user` module's use case — don't import `user.domain.User` directly into `order.application`.

---

## 9. Local development setup

### Prerequisites

- Go 1.25+
- Postgres (local Docker is fine)
- `gcloud` CLI logged in: `gcloud auth application-default login`
- A GCP project with Secret Manager enabled
- A secret named e.g. `artsgoz-env` containing a `.env`-formatted blob

### One-time setup

```bash
# 1. Clone & install deps
git clone <repo>
cd artsgoz-backend
go mod download

# 2. Create .env.local (only loads GCP bootstrap vars)
cat > .env.local <<'EOF'
GCP_PROJECT_ID=your-gcp-project
GCP_SECRET_ID=artsgoz-env
EOF

# 3. Put real config in GCP Secret Manager
cat > /tmp/artsgoz.env <<'EOF'
DATABASE_URL=postgres://artsgoz:hunter2@localhost:5432/artsgoz?sslmode=disable
JWT_SECRET=changeme
EOF
gcloud secrets create artsgoz-env --data-file=/tmp/artsgoz.env
rm /tmp/artsgoz.env

# 4. Start a local Postgres
docker run -d --name artsgoz-pg -p 5432:5432 \
  -e POSTGRES_USER=artsgoz -e POSTGRES_PASSWORD=hunter2 -e POSTGRES_DB=artsgoz \
  postgres:16

# 5. Run migrations (see §9.1 below for details)
go run ./cmd/migrate up

# 6. Run the server
go run ./cmd
```

### 9.1 Migrations

A **migration** is a SQL file that takes the database schema from one version to the next. They live in `migrations/` with the naming `NNNN_description.{up,down}.sql`:

- `0001_create_users.up.sql` — applies the change (creates the `users` table)
- `0001_create_users.down.sql` — reverses it (drops the table)

A small migration *runner* tracks which files have already been applied (using a `schema_migrations` table inside your DB) and runs only the new ones. We ship one as `cmd/migrate/`. It reads `DATABASE_URL` from the same Secret Manager blob as the main app, so it runs with the exact same auth.

**Common commands:**

```bash
# Apply every pending migration (use this 99% of the time)
go run ./cmd/migrate up

# Roll back the most recent migration (development only — never in prod
# unless you really mean it)
go run ./cmd/migrate down

# Show the current schema version
go run ./cmd/migrate version

# Jump to a specific version (up OR down)
go run ./cmd/migrate goto 5
```

**Creating a new migration:**

1. Pick the next number. If the latest file is `0003_*`, your new one is `0004_*`.
2. Create *both* an up and a down file, even if down is rarely used:
   ```bash
   touch migrations/0004_add_user_display_name.up.sql
   touch migrations/0004_add_user_display_name.down.sql
   ```
3. Write SQL:
   ```sql
   -- 0004_add_user_display_name.up.sql
   ALTER TABLE users ADD COLUMN display_name TEXT;

   -- 0004_add_user_display_name.down.sql
   ALTER TABLE users DROP COLUMN display_name;
   ```
4. Apply: `go run ./cmd/migrate up`
5. Commit both files alongside your code change.

**Migration rules of thumb:**

- ✅ **One change per migration.** A migration that creates a table *and* adds a column to another table is harder to roll back. Split it.
- ✅ **Always write the `down` file.** Even if you'll never run it, writing it forces you to think about reversibility.
- ✅ **Migrations are append-only once merged.** Never edit an already-merged migration — write a new one that fixes it.
- ✅ **Make migrations idempotent where possible.** `CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`. Saves your bacon during partial failures.
- ⚠️ **Be careful with destructive changes.** `DROP COLUMN`, `DROP TABLE`, `ALTER COLUMN` on a live DB can be slow or lock tables. For prod databases, do these in two steps: ship code that no longer reads the column, *then* drop it in a later migration.
- ⚠️ **Don't include test data.** Migrations are for schema. Seed data belongs in a separate `cmd/seed/` (future).

**The "dirty" state:**

If a migration fails halfway, golang-migrate marks the version as `dirty` and refuses to run more migrations until you fix it. Workflow:

```bash
go run ./cmd/migrate version     # outputs: version: 4 (DIRTY ...)
# Manually fix the DB to the state the up.sql should have produced
go run ./cmd/migrate force 4     # tells migrate "we're fine at v4 now"
go run ./cmd/migrate up          # continue
```

This happens rarely in dev (just re-create the DB) and should *never* happen in prod (CI runs migrations in a transaction-friendly way — see below).

**Where migrations run in different environments:**

| Environment | How |
|---|---|
| **Local dev** | You manually: `go run ./cmd/migrate up` after pulling new code. |
| **CI** | Run against a throwaway Postgres in the test job, before integration tests. |
| **Staging/Prod** | A dedicated CI step (or a Kubernetes Job / Cloud Run Job) runs `go run ./cmd/migrate up` *before* the new app version is rolled out. **Never** auto-migrate from the running app — multiple replicas would race. |

### Smoke test

```bash
# Register a user
curl -X POST localhost:3000/api/v1/users/register \
  -H 'content-type: application/json' \
  -d '{"email":"a@b.com","password":"hunter2x"}'
# → 201 {"id":"...","email":"a@b.com"}

# Duplicate → 409
curl -X POST localhost:3000/api/v1/users/register \
  -H 'content-type: application/json' \
  -d '{"email":"a@b.com","password":"hunter2x"}'

# Validation error → 400
curl -X POST localhost:3000/api/v1/users/register \
  -H 'content-type: application/json' \
  -d '{"email":"nope","password":"x"}'
```

---

## 10. FAQ

**Q: Why so many folders for one feature?**
A: Each folder has *one* job, which makes the code easy to scan and impossible to mis-place. The boilerplate cost pays itself back the second time you change databases or write a unit test. If your feature truly only needs a handler (e.g. `/health`), it's fine to skip the application layer for that one route.

**Q: Where do utility functions go?**
A: If it's domain-related (e.g. `IsValidThaiPhone`), inside the relevant `domain/` package. If it's project-wide infrastructure (e.g. JWT encoder), under `internal/platform/`. If it's reusable across projects (and very rare), `pkg/`. **Never** create a `utils/` grab-bag.

**Q: When do I add a new module vs. just a new use case?**
A: New module = new bounded context = new business noun. New use case = new verb on an existing noun. "User can change email" → new use case under `user/`. "We now sell merchandise" → new module `merchandise/`.

**Q: Can two use cases share code?**
A: Yes, via the domain. Pull shared logic into the entity (a method on `User`) or a domain service. Don't extract a "shared use case helper" — that usually means the logic belongs in the entity.

**Q: My use case needs data from another module — what do I do?**
A: Expose it through the other module's use cases (call them) or via a read-only query interface. Never reach into another module's `domain/` directly.

**Q: Where do I put cross-cutting middleware (auth, logging, CORS)?**
A: `internal/platform/httpserver/` for Fiber-level middleware. Domain-aware guards (e.g. "must be admin") can live in a shared `internal/modules/auth/interface/rest/` module that other modules import.

**Q: This feels like a lot of ceremony for small features.**
A: It is. The payoff appears around the 5th feature and grows from there. If you find yourself fighting the structure for a one-off endpoint, it's okay to keep the handler thin and skip the application layer — but document the exception.

**Q: How do I add a domain event later (e.g. `UserRegistered`)?**
A: Define the event struct in `domain/events/`, give the entity a way to record it (`user.PullEvents()`), and add a publisher interface in `domain/`. Implement the publisher in `infrastructure/` (PubSub, Kafka, outbox). The use case calls `publisher.Publish(user.PullEvents())` after `repo.Save`.

---

## References

- Eric Evans, *Domain-Driven Design* (the original book)
- Vaughn Vernon, *Implementing Domain-Driven Design* (more practical)
- Robert C. Martin, *Clean Architecture*
- [The Clean Architecture blog post](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- Internal sibling: `wise-monorepo/apps/wise-api` (NestJS implementation of the same patterns)
