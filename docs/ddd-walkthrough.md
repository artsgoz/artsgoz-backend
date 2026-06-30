# DDD & Clean Architecture — Step-by-Step Walkthrough

> Companion to [`architecture.md`](./architecture.md). That doc is the *reference* — this one is the *tutorial*. Read this once when you join, then keep `architecture.md` open while you code.

This walkthrough builds a complete bounded context — **`artwork`** — from scratch, the way a contributor would when adding a new business area to artsgoz-backend. Every step explains *what* to do, *why* it's that way, and *what would go wrong* if you skipped it.

By the end you will have:

- A full mental model of how DDD identifies what to build.
- A concrete `artwork` module with all four layers, plus tests and a migration.
- Decision rules you can apply to *future* modules without copy-pasting.

---

## Table of contents

1. [Mental model: the 30-second version](#1-mental-model-the-30-second-version)
2. [Phase 0 — Before you write code](#2-phase-0--before-you-write-code)
3. [Phase 1 — Discover the bounded context](#3-phase-1--discover-the-bounded-context)
4. [Phase 2 — Find entities, value objects, and invariants](#4-phase-2--find-entities-value-objects-and-invariants)
5. [Phase 3 — Choose aggregate roots](#5-phase-3--choose-aggregate-roots)
6. [Phase 4 — List the use cases](#6-phase-4--list-the-use-cases)
7. [Phase 5 — Design the repository port](#7-phase-5--design-the-repository-port)
8. [Phase 6 — Plan the HTTP surface](#8-phase-6--plan-the-http-surface)
9. [Phase 7 — Implement, layer by layer](#9-phase-7--implement-layer-by-layer)
10. [Phase 8 — Tests](#10-phase-8--tests)
11. [Phase 9 — Wire and migrate](#11-phase-9--wire-and-migrate)
12. [Decision flowcharts](#12-decision-flowcharts)
13. [Common mistakes & how to recover](#13-common-mistakes--how-to-recover)
14. [Cheat sheet](#14-cheat-sheet)

---

## 1. Mental model: the 30-second version

> **DDD** tells you *what* to build (what objects, what rules, what boundaries).
> **Clean Architecture** tells you *where* to put it (which layer owns which code).

The two reinforce each other:

- DDD says "the rule that an artwork's price must be > 0 is a domain rule."
- Clean Architecture says "therefore that rule lives in `domain/`, never in the HTTP handler."

If you only remember three things:

1. **Push business rules inward** (toward `domain/`). HTTP/SQL stay at the edge.
2. **Depend on interfaces, not concrete types.** Domain declares; infrastructure implements.
3. **Each bounded context is its own folder.** Don't share entities across modules.

---

## 2. Phase 0 — Before you write code

Resist the urge to `mkdir` immediately. Spend 15 minutes answering these questions on paper / a doc / a whiteboard:

1. **What is this feature for?** One sentence.
2. **Who triggers it?** A user? A scheduled job? Another service?
3. **What are the nouns?** (These will become entities or value objects.)
4. **What are the verbs?** (These will become use cases.)
5. **What rules / invariants must always hold?** (These shape the entity constructors.)
6. **What questions will we ask the database?** (These become repository methods.)

> **Tip**: write this as a 1-page brief and paste it at the top of your PR. Reviewers don't need to re-discover the requirements from code.

### Example brief — Artworks

> The platform sells digital artworks. An **artist** can **publish** an artwork by submitting a title, an asset URL, a price in THB, and a category. Once published, an artwork can be **viewed** (anyone) or **archived** (only the original artist). Prices must be > 0. Titles must be 1–120 characters. Archived artworks don't show up in listings. Artists are existing `user` records.

That single paragraph determines the entire module's shape. Let's extract it.

---

## 3. Phase 1 — Discover the bounded context

A **bounded context** is a chunk of the business with its own consistent language and rules. It maps to one folder under `internal/modules/`.

**Three tests** to confirm artwork is its own bounded context (and not, say, part of `user`):

| Test | Result |
|---|---|
| Does it have its own primary entity? | Yes — `Artwork` |
| Could a different team own it without touching `user`? | Yes |
| Does it use vocabulary unique to itself? | Yes — *publish, archive, category, asset URL* |

✅ → it's a separate bounded context. New module: `internal/modules/artwork/`.

❌ If all answers had been "no" (e.g. "user's avatar URL"), it would belong inside `user/`.

> **Why this matters**: cross-cutting an artificial boundary causes 80% of clean-architecture pain. Two separate-but-shared modules are easier than one tangled module.

---

## 4. Phase 2 — Find entities, value objects, and invariants

Re-read the brief and **circle every noun**:

> The platform sells digital **artworks**. An **artist** can **publish** an **artwork** by submitting a **title**, an **asset URL**, a **price** in THB, and a **category**. Once published, an **artwork** can be **viewed** (anyone) or **archived** (only the original artist). **Prices** must be > 0. **Titles** must be 1–120 characters. Archived artworks don't show up in listings. Artists are existing user records.

### Classify each noun

For each one, ask: **does its identity matter, or only its value?**

| Noun | Identity matters? | Classification |
|---|---|---|
| Artwork | Yes — has an ID, changes over time (publish → archive) | **Entity** (and an aggregate root) |
| Artist | Yes — but lives in the `user` module already | External reference (`UserID`) |
| Title | No — "Sunset" is "Sunset" anywhere | **Value object** (or just a `string` with validation) |
| Asset URL | No | **Value object** (or just a `string`) |
| Price | No, but has structure (amount + currency) | **Value object** |
| Category | No — just a label | **Enum / value object** |

> **Rule of thumb**: when in doubt, start with primitive Go types (`string`, `int`) and validate in the entity constructor. Promote to a value-object type *only* if the validation logic is reused across entities or if grouping fields together (Money = amount + currency) buys you safety.

### List the invariants

These are rules the entity must enforce *itself*. Once they're listed, they become checks inside the constructor or a method.

For `Artwork`:

1. `Title` is 1–120 characters, trimmed.
2. `Price` is > 0.
3. `Category` is one of a known set.
4. `ArtistID` is non-empty.
5. Only the original artist may archive the artwork.
6. An archived artwork cannot be archived again.

Invariants 1–4 belong in the constructor. Invariants 5–6 belong in an `Archive(actingUserID)` method.

> **Anti-pattern alert**: don't put these checks in the use case. The use case can call methods on the entity, but the entity must never trust its caller to have validated invariants. This is how DDD avoids the dreaded *anemic model* (where entities are just bags of fields and all logic is elsewhere).

---

## 5. Phase 3 — Choose aggregate roots

An **aggregate** is a cluster of related objects treated as a single unit for data changes. The **aggregate root** is the only object outside code can hold a reference to.

For `Artwork`, the cluster is trivial — there's just one entity. So `Artwork` *is* the aggregate root.

When does it stop being trivial?

> Example: if we added `Comments` on artworks, you'd ask: "are comments meaningful without the artwork?" If no, they're inside the `Artwork` aggregate (load via the root, no separate `CommentRepository`). If yes, they're their own aggregate with their own root and repository, linked by `ArtworkID`.

**Aggregate sizing rules:**

- ✅ Keep aggregates **small**. One root + a few value objects beats one root + 50 children.
- ✅ Modify **one aggregate per transaction**. If a use case needs to mutate two aggregates atomically, that's a signal you've drawn the boundary wrong (or you need a saga).
- ❌ Don't create giant aggregates "just in case." You can always split later; merging is harder.

---

## 6. Phase 4 — List the use cases

Use cases are **verb phrases** that the system *does*. Pull them straight from the brief:

| Verb phrase | Use case name |
|---|---|
| "An artist publishes an artwork" | `PublishArtwork` |
| "Anyone views an artwork" | `GetArtwork` |
| "The artist archives an artwork" | `ArchiveArtwork` |
| "Anyone lists active artworks" | `ListArtworks` |

Each becomes one file in `application/`. One file = one struct = one `Execute` method.

> **Granularity rule**: a use case is one **complete business operation**, not one **CRUD function**. "Publish an artwork" is a use case. "Update artwork.title" usually isn't — it's part of "Edit artwork details" if such a thing exists in the brief.

---

## 7. Phase 5 — Design the repository port

The repository interface (the *port*) lives in `domain/`. It encodes "what does this module need to ask the database?" without saying *how*.

Start by listing the queries each use case needs:

| Use case | Query needed |
|---|---|
| PublishArtwork | Insert a new artwork |
| GetArtwork | Look up by ID |
| ArchiveArtwork | Look up by ID, then update |
| ListArtworks | Find all where archived = false, ordered, paginated |

That gives us a draft interface:

```go
type ArtworkRepository interface {
    Save(ctx context.Context, a *Artwork) error
    FindByID(ctx context.Context, id string) (*Artwork, error)
    ListActive(ctx context.Context, limit, offset int) ([]*Artwork, error)
}
```

> **Notice**: there's no `Update` method. We use `Save` for both insert and update (the repository decides based on whether the row exists). This keeps the domain unaware of SQL operations.

> **Notice**: there's no `Delete`. We don't physically delete artworks; we archive them — a domain operation, not a persistence operation. (If you did want hard delete, that'd be `Delete(ctx, id)`.)

---

## 8. Phase 6 — Plan the HTTP surface

| Method | Path | Use case |
|---|---|---|
| POST | `/api/v1/artworks` | PublishArtwork |
| GET | `/api/v1/artworks/:id` | GetArtwork |
| POST | `/api/v1/artworks/:id/archive` | ArchiveArtwork |
| GET | `/api/v1/artworks` | ListArtworks |

> **Note on REST style**: `POST /:id/archive` is more honest than `PATCH /:id { archived: true }` because "archive" is a domain verb with rules attached. REST purists will argue; pick what reads better.

---

## 9. Phase 7 — Implement, layer by layer

Now the typing starts. Order matters: **work inside-out** (domain → application → infrastructure → interface). This way each layer is testable as soon as you finish it.

### 9.1 Domain layer

**File: `internal/modules/artwork/domain/category.go`** — small value-object example.

```go
package domain

import "github.com/artsgoz/artsgoz-backend/internal/platform/apperr"

type Category string

const (
    CategoryPainting     Category = "painting"
    CategoryPhotography  Category = "photography"
    CategoryIllustration Category = "illustration"
    CategorySculpture    Category = "sculpture"
)

func ParseCategory(s string) (Category, error) {
    switch Category(s) {
    case CategoryPainting, CategoryPhotography, CategoryIllustration, CategorySculpture:
        return Category(s), nil
    default:
        return "", apperr.Validation("unknown category: "+s, nil)
    }
}
```

> Why a typed `Category` instead of a `string`? It makes invalid categories *unrepresentable* in code that already has a `Category` value. The validation happens once, at the edge.

**File: `internal/modules/artwork/domain/artwork.go`** — the aggregate root.

```go
package domain

import (
    "strings"
    "time"

    "github.com/google/uuid"

    "github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type Artwork struct {
    ID         string
    ArtistID   string
    Title      string
    AssetURL   string
    PriceTHB   int64       // store money as cents/satang to avoid floats
    Category   Category
    Archived   bool
    CreatedAt  time.Time
    ArchivedAt *time.Time
}

// NewArtwork enforces all invariants at construction time. Once it returns
// successfully, the *Artwork is guaranteed valid.
func NewArtwork(artistID, title, assetURL string, priceSatang int64, category Category) (*Artwork, error) {
    title = strings.TrimSpace(title)
    if artistID == "" {
        return nil, apperr.Validation("artistID is required", nil)
    }
    if title == "" || len(title) > 120 {
        return nil, apperr.Validation("title must be 1..120 characters", nil)
    }
    if assetURL == "" {
        return nil, apperr.Validation("assetURL is required", nil)
    }
    if priceSatang <= 0 {
        return nil, apperr.Validation("price must be > 0", nil)
    }
    return &Artwork{
        ID:        uuid.NewString(),
        ArtistID:  artistID,
        Title:     title,
        AssetURL:  assetURL,
        PriceTHB:  priceSatang,
        Category:  category,
        Archived:  false,
        CreatedAt: time.Now().UTC(),
    }, nil
}

// Archive transitions the aggregate. It enforces that only the original
// artist may archive, and that an already-archived artwork can't be archived
// again. The use case never has to remember these rules — the entity does.
func (a *Artwork) Archive(actingUserID string) error {
    if a.ArtistID != actingUserID {
        return apperr.Unauthorized("only the artist may archive this artwork")
    }
    if a.Archived {
        return apperr.Conflict("artwork already archived")
    }
    now := time.Now().UTC()
    a.Archived = true
    a.ArchivedAt = &now
    return nil
}
```

> **The Archive method is where the *real* DDD value lives.** If you put `if artwork.ArtistID != actingUserID` in the use case (or worse, the HTTP handler), then **every** use case that archives has to re-check it. Eventually someone forgets. Putting the rule on the entity guarantees it can't be bypassed.

**File: `internal/modules/artwork/domain/artwork_repository.go`**

```go
package domain

import "context"

type ArtworkRepository interface {
    Save(ctx context.Context, a *Artwork) error
    FindByID(ctx context.Context, id string) (*Artwork, error)
    ListActive(ctx context.Context, limit, offset int) ([]*Artwork, error)
}
```

### 9.2 Application layer

**File: `internal/modules/artwork/application/dto.go`**

```go
package application

import "time"

type PublishArtworkInput struct {
    ArtistID    string `json:"-"                       validate:"required"` // injected by auth, not from body
    Title       string `json:"title"                    validate:"required,min=1,max=120"`
    AssetURL    string `json:"assetUrl"                 validate:"required,url"`
    PriceSatang int64  `json:"priceSatang"              validate:"required,gt=0"`
    Category    string `json:"category"                 validate:"required"`
}

type ArtworkOutput struct {
    ID         string     `json:"id"`
    ArtistID   string     `json:"artistId"`
    Title      string     `json:"title"`
    AssetURL   string     `json:"assetUrl"`
    PriceSatang int64     `json:"priceSatang"`
    Category   string     `json:"category"`
    Archived   bool       `json:"archived"`
    CreatedAt  time.Time  `json:"createdAt"`
    ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}

type ArchiveArtworkInput struct {
    ArtworkID    string `json:"-" validate:"required"`
    ActingUserID string `json:"-" validate:"required"`
}

type ListArtworksInput struct {
    Limit  int `json:"-" validate:"min=1,max=100"`
    Offset int `json:"-" validate:"min=0"`
}
```

> Notice `json:"-"` on fields that come from the URL path or auth context, not from the request body. We bind those in the handler, not from JSON.

**File: `internal/modules/artwork/application/publish_artwork.go`**

```go
package application

import (
    "context"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/domain"
    "github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type PublishArtwork struct {
    repo domain.ArtworkRepository
}

func NewPublishArtwork(r domain.ArtworkRepository) *PublishArtwork {
    return &PublishArtwork{repo: r}
}

func (uc *PublishArtwork) Execute(ctx context.Context, in PublishArtworkInput) (ArtworkOutput, error) {
    if err := validator.Struct(in); err != nil {
        return ArtworkOutput{}, err
    }
    cat, err := domain.ParseCategory(in.Category)
    if err != nil {
        return ArtworkOutput{}, err
    }
    art, err := domain.NewArtwork(in.ArtistID, in.Title, in.AssetURL, in.PriceSatang, cat)
    if err != nil {
        return ArtworkOutput{}, err
    }
    if err := uc.repo.Save(ctx, art); err != nil {
        return ArtworkOutput{}, err
    }
    return toOutput(art), nil
}

func toOutput(a *domain.Artwork) ArtworkOutput {
    return ArtworkOutput{
        ID:          a.ID,
        ArtistID:    a.ArtistID,
        Title:       a.Title,
        AssetURL:    a.AssetURL,
        PriceSatang: a.PriceTHB,
        Category:    string(a.Category),
        Archived:    a.Archived,
        CreatedAt:   a.CreatedAt,
        ArchivedAt:  a.ArchivedAt,
    }
}
```

**File: `internal/modules/artwork/application/archive_artwork.go`**

```go
package application

import (
    "context"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/domain"
    "github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type ArchiveArtwork struct {
    repo domain.ArtworkRepository
}

func NewArchiveArtwork(r domain.ArtworkRepository) *ArchiveArtwork {
    return &ArchiveArtwork{repo: r}
}

func (uc *ArchiveArtwork) Execute(ctx context.Context, in ArchiveArtworkInput) (ArtworkOutput, error) {
    if err := validator.Struct(in); err != nil {
        return ArtworkOutput{}, err
    }
    art, err := uc.repo.FindByID(ctx, in.ArtworkID)
    if err != nil {
        return ArtworkOutput{}, err
    }
    if err := art.Archive(in.ActingUserID); err != nil {
        return ArtworkOutput{}, err
    }
    if err := uc.repo.Save(ctx, art); err != nil {
        return ArtworkOutput{}, err
    }
    return toOutput(art), nil
}
```

> See how trivial the use case is? Load → call domain method → save. The intelligence lives in `art.Archive(...)`. That's the goal.

### 9.3 Infrastructure layer

**File: `internal/modules/artwork/infrastructure/persistence/postgres/artwork_repository.go`**

```go
package postgres

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/domain"
    "github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type artworkRepository struct {
    pool *pgxpool.Pool
}

func NewArtworkRepository(pool *pgxpool.Pool) domain.ArtworkRepository {
    return &artworkRepository{pool: pool}
}

func (r *artworkRepository) Save(ctx context.Context, a *domain.Artwork) error {
    const q = `
        INSERT INTO artworks (id, artist_id, title, asset_url, price_satang,
                              category, archived, created_at, archived_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO UPDATE SET
            title        = EXCLUDED.title,
            asset_url    = EXCLUDED.asset_url,
            price_satang = EXCLUDED.price_satang,
            category     = EXCLUDED.category,
            archived     = EXCLUDED.archived,
            archived_at  = EXCLUDED.archived_at
    `
    _, err := r.pool.Exec(ctx, q,
        a.ID, a.ArtistID, a.Title, a.AssetURL, a.PriceTHB,
        string(a.Category), a.Archived, a.CreatedAt, a.ArchivedAt,
    )
    if err != nil {
        return apperr.Internal("save artwork", err)
    }
    return nil
}

func (r *artworkRepository) FindByID(ctx context.Context, id string) (*domain.Artwork, error) {
    const q = `
        SELECT id, artist_id, title, asset_url, price_satang, category,
               archived, created_at, archived_at
        FROM artworks
        WHERE id = $1
    `
    var a domain.Artwork
    var cat string
    err := r.pool.QueryRow(ctx, q, id).Scan(
        &a.ID, &a.ArtistID, &a.Title, &a.AssetURL, &a.PriceTHB, &cat,
        &a.Archived, &a.CreatedAt, &a.ArchivedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperr.NotFound("artwork not found")
        }
        return nil, apperr.Internal("find artwork", err)
    }
    a.Category = domain.Category(cat)
    return &a, nil
}

func (r *artworkRepository) ListActive(ctx context.Context, limit, offset int) ([]*domain.Artwork, error) {
    const q = `
        SELECT id, artist_id, title, asset_url, price_satang, category,
               archived, created_at, archived_at
        FROM artworks
        WHERE archived = FALSE
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
    rows, err := r.pool.Query(ctx, q, limit, offset)
    if err != nil {
        return nil, apperr.Internal("list artworks", err)
    }
    defer rows.Close()

    var out []*domain.Artwork
    for rows.Next() {
        var a domain.Artwork
        var cat string
        if err := rows.Scan(&a.ID, &a.ArtistID, &a.Title, &a.AssetURL, &a.PriceTHB,
            &cat, &a.Archived, &a.CreatedAt, &a.ArchivedAt); err != nil {
            return nil, apperr.Internal("scan artwork", err)
        }
        a.Category = domain.Category(cat)
        out = append(out, &a)
    }
    return out, rows.Err()
}
```

### 9.4 Interface layer

**File: `internal/modules/artwork/interface/rest/handler.go`**

```go
package rest

import (
    "strconv"

    "github.com/gofiber/fiber/v3"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/application"
    "github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type ArtworkHandler struct {
    publish *application.PublishArtwork
    archive *application.ArchiveArtwork
    // get, list use cases would go here
}

func NewArtworkHandler(p *application.PublishArtwork, a *application.ArchiveArtwork) *ArtworkHandler {
    return &ArtworkHandler{publish: p, archive: a}
}

func (h *ArtworkHandler) Publish(c fiber.Ctx) error {
    var in application.PublishArtworkInput
    if err := c.Bind().Body(&in); err != nil {
        return apperr.Validation("invalid request body", err.Error())
    }
    // In a real app, ArtistID comes from authenticated session, not the body.
    in.ArtistID = c.Get("X-User-ID") // placeholder until auth middleware exists
    out, err := h.publish.Execute(c.Context(), in)
    if err != nil {
        return err
    }
    return c.Status(fiber.StatusCreated).JSON(out)
}

func (h *ArtworkHandler) Archive(c fiber.Ctx) error {
    in := application.ArchiveArtworkInput{
        ArtworkID:    c.Params("id"),
        ActingUserID: c.Get("X-User-ID"),
    }
    out, err := h.archive.Execute(c.Context(), in)
    if err != nil {
        return err
    }
    return c.JSON(out)
}

// Helper for parsing query params with defaults.
func parseIntQuery(c fiber.Ctx, key string, def int) int {
    if v, err := strconv.Atoi(c.Query(key)); err == nil {
        return v
    }
    return def
}
```

**File: `internal/modules/artwork/interface/rest/routes.go`**

```go
package rest

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router, h *ArtworkHandler) {
    g := r.Group("/artworks")
    g.Post("/", h.Publish)
    g.Post("/:id/archive", h.Archive)
}
```

### 9.5 Module wiring

**File: `internal/modules/artwork/module.go`**

```go
package artwork

import (
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/application"
    pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/infrastructure/persistence/postgres"
    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/interface/rest"
)

type Module struct {
    Handler *rest.ArtworkHandler
}

func New(pool *pgxpool.Pool) *Module {
    repo := pgrepo.NewArtworkRepository(pool)
    publishUC := application.NewPublishArtwork(repo)
    archiveUC := application.NewArchiveArtwork(repo)
    handler := rest.NewArtworkHandler(publishUC, archiveUC)
    return &Module{Handler: handler}
}
```

---

## 10. Phase 8 — Tests

The dependency inversion means you can test the use case **without a database**.

**File: `internal/modules/artwork/application/publish_artwork_test.go`**

```go
package application_test

import (
    "context"
    "errors"
    "testing"

    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/application"
    "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/domain"
    "github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

type fakeRepo struct {
    saved []*domain.Artwork
}

func (f *fakeRepo) Save(_ context.Context, a *domain.Artwork) error {
    f.saved = append(f.saved, a)
    return nil
}
func (f *fakeRepo) FindByID(_ context.Context, _ string) (*domain.Artwork, error) {
    return nil, apperr.NotFound("not used here")
}
func (f *fakeRepo) ListActive(_ context.Context, _, _ int) ([]*domain.Artwork, error) {
    return nil, nil
}

func TestPublishArtwork_Success(t *testing.T) {
    repo := &fakeRepo{}
    uc := application.NewPublishArtwork(repo)

    out, err := uc.Execute(context.Background(), application.PublishArtworkInput{
        ArtistID:    "user-1",
        Title:       "Sunset",
        AssetURL:    "https://cdn.example/x.png",
        PriceSatang: 12000, // 120.00 THB
        Category:    "painting",
    })
    if err != nil {
        t.Fatalf("unexpected: %v", err)
    }
    if len(repo.saved) != 1 {
        t.Fatalf("Save calls = %d, want 1", len(repo.saved))
    }
    if out.Title != "Sunset" || out.Category != "painting" {
        t.Errorf("unexpected output: %+v", out)
    }
}

func TestPublishArtwork_InvalidPrice(t *testing.T) {
    uc := application.NewPublishArtwork(&fakeRepo{})
    _, err := uc.Execute(context.Background(), application.PublishArtworkInput{
        ArtistID: "user-1", Title: "x", AssetURL: "https://x", PriceSatang: -1, Category: "painting",
    })
    var e *apperr.Error
    if !errors.As(err, &e) || e.Kind != apperr.KindValidation {
        t.Fatalf("want validation error, got %v", err)
    }
}
```

For `ArchiveArtwork`, the test asserts the *domain rule* fires correctly:

```go
func TestArchiveArtwork_OnlyArtistMayArchive(t *testing.T) {
    art, _ := domain.NewArtwork("artist-1", "x", "https://x", 100, domain.CategoryPainting)
    repo := &fakeRepo{}
    repo.findResult = art // assume fakeRepo extended with findResult

    uc := application.NewArchiveArtwork(repo)
    _, err := uc.Execute(context.Background(), application.ArchiveArtworkInput{
        ArtworkID: art.ID, ActingUserID: "someone-else",
    })
    var e *apperr.Error
    if !errors.As(err, &e) || e.Kind != apperr.KindUnauthorized {
        t.Fatalf("want unauthorized, got %v", err)
    }
}
```

> **The whole point**: this test runs in milliseconds, runs anywhere (no Docker, no Postgres), and breaks loudly if a future contributor moves the "only artist may archive" check elsewhere and forgets to enforce it.

---

## 11. Phase 9 — Wire and migrate

**Migration: `migrations/0002_create_artworks.up.sql`**

```sql
CREATE TABLE IF NOT EXISTS artworks (
    id            UUID PRIMARY KEY,
    artist_id     UUID NOT NULL REFERENCES users(id),
    title         TEXT NOT NULL,
    asset_url     TEXT NOT NULL,
    price_satang  BIGINT NOT NULL CHECK (price_satang > 0),
    category      TEXT NOT NULL,
    archived      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS artworks_active_created_idx
    ON artworks (archived, created_at DESC);
```

**Migration: `migrations/0002_create_artworks.down.sql`**

```sql
DROP TABLE IF EXISTS artworks;
```

**Mount in `cmd/main.go`** (add two lines):

```go
import (
    artwork "github.com/artsgoz/artsgoz-backend/internal/modules/artwork"
    artworkrest "github.com/artsgoz/artsgoz-backend/internal/modules/artwork/interface/rest"
)

// ... inside main() after `userMod := user.New(pool)`:
artworkMod := artwork.New(pool)
artworkrest.RegisterRoutes(api, artworkMod.Handler)
```

**Verify:**

```bash
go run ./cmd/migrate up        # apply 0002_create_artworks
go build ./...                 # compile
go test ./...                  # unit tests pass
./scripts/check-deps.sh        # no dependency-rule violations
go run ./cmd                   # serve
```

Smoke test:

```bash
curl -X POST localhost:3000/api/v1/artworks \
  -H 'content-type: application/json' \
  -H 'X-User-ID: <some-user-uuid>' \
  -d '{"title":"Sunset","assetUrl":"https://cdn/x.png","priceSatang":12000,"category":"painting"}'
```

---

## 12. Decision flowcharts

### Should this be an entity or a value object?

```
Does it have a stable identity (an ID) that survives changes to its attributes?
│
├── YES ──► Entity. Probably needs a repository.
│
└── NO ───► Value Object. Validate in constructor, treat as immutable.
            Examples: Money (amount+currency), EmailAddress, Coordinates
```

### Where should this logic live?

```
Is the logic a rule about a single aggregate's internal state?
│
├── YES ──► Method on the aggregate root.
│           e.g. artwork.Archive(userID), user.ChangeEmail(newEmail)
│
└── NO ───► Does it orchestrate multiple aggregates / external services?
            │
            ├── YES ──► Use case (application layer).
            │           e.g. ProcessOrder calls UserRepo + PaymentService + OrderRepo
            │
            └── NO ───► Domain service (rare — only when logic doesn't fit on
                        a single entity AND isn't orchestration).
                        e.g. CurrencyConverter, PricingPolicy
```

### Should this be a new bounded context or a new use case in an existing one?

```
Does the new feature primarily manipulate an existing aggregate?
│
├── YES ──► New use case in the existing module.
│           e.g. "change user email" → new use case in user/
│
└── NO ───► Is there a new primary noun with its own lifecycle and rules?
            │
            ├── YES ──► New bounded context (new module).
            │           e.g. orders, payments, comments
            │
            └── NO ───► Probably belongs in `internal/platform/` as a
                        cross-cutting concern. e.g. notifications, audit log.
```

### Should I add a repository method, or filter in code?

```
Will this query be used in production hot paths or with large datasets?
│
├── YES ──► Add a repository method. Push the filter into SQL.
│           e.g. ListActiveByCategory(ctx, cat, limit, offset)
│
└── NO ───► Use an existing method and filter in code.
            Avoids method explosion. Refactor later if perf demands it.
```

---

## 13. Common mistakes & how to recover

### Mistake 1 — Anemic domain model

**Symptom**: your entities are pure data (`struct { ID, Email string }`) with no methods. All logic lives in use cases. Same rules get re-checked in multiple use cases.

**Fix**: pull the rules into entity methods. `user.ChangeEmail(newEmail)` instead of doing the email validation in 4 different use cases. Start with one rule, refactor incrementally.

### Mistake 2 — Repository methods that leak persistence

**Symptom**: `repo.FindByEmailWhereStatusActiveAndCreatedAfter(...)` — repository methods named for SQL clauses.

**Fix**: name methods in domain terms (`FindRecentActiveByEmail`), or expose a small query DSL (`repo.Find(query)` where query is a domain struct).

### Mistake 3 — Cross-aggregate transaction

**Symptom**: a use case modifies two aggregates and you need both writes to succeed-or-fail together. You wrap them in `BEGIN ... COMMIT`.

**Two valid fixes**:
- **Reconsider boundaries**: maybe they should be one aggregate after all.
- **Eventual consistency**: write one aggregate now, emit a domain event, project the other aggregate in a separate transaction. (Add an `outbox` table for reliability.)

### Mistake 4 — God-entity

**Symptom**: `Order` has 60 fields and 40 methods. Every change touches it.

**Fix**: extract value objects (`Address`, `LineItem`, `PaymentInfo`). Split into multiple aggregates if independent lifecycles exist (`Order` + `Shipment` + `Invoice`).

### Mistake 5 — Shared entity across modules

**Symptom**: `artwork/application` imports `user/domain/User` to read the artist's name.

**Fix**: don't. Either (a) expose a small read model in `user` that returns just `{ID, DisplayName}`, or (b) duplicate the field in `artworks` if the dependency is too tight. Cross-module reaches kill the value of bounded contexts.

### Mistake 6 — "Service" struct that does everything

**Symptom**: `UserService` has 30 methods doing HTTP parsing, SQL, business logic, and email sending.

**Fix**: split. Each method becomes a use case (`RegisterUser`, `LoginUser`, `SendPasswordReset`). Repository, mailer, etc. are separate interfaces injected into the use cases.

### Mistake 7 — Using the database as a domain event bus

**Symptom**: triggers/cron polling `WHERE processed = false` to react to changes.

**Fix**: emit domain events from the aggregate (`user.PullEvents()`), persist them in an outbox table in the same transaction as the entity change, then publish to a real bus (PubSub, Kafka) asynchronously.

---

## 14. Cheat sheet

| Layer | Owns | Imports allowed |
|---|---|---|
| `domain/` | Entities, value objects, aggregate methods, repository **interfaces**, domain services, domain events | stdlib only (+ `internal/platform/apperr`) |
| `application/` | Use cases (one struct, one `Execute` method), input/output DTOs | `domain/`, `internal/platform/{apperr,validator}` |
| `infrastructure/` | Repository **implementations**, external service clients (mail, storage, payments) | `domain/`, drivers (pgx, GCS SDK, …), `internal/platform/apperr` |
| `interface/` | HTTP handlers, route registration, request parsing, auth middleware | `application/`, Fiber, `internal/platform/*` |
| `cmd/` | Composition root (`main.go`) | everything |

| When you... | Create / edit |
|---|---|
| Need a new business operation on existing nouns | A use case in `application/` |
| Discover a new business noun with its own lifecycle | A new module under `internal/modules/<noun>/` |
| Need a new way to query the database | A method on the repository interface + implementation |
| Need a new HTTP endpoint for an existing use case | A handler method + route in `interface/rest/` |
| Need to enforce a rule about an entity's state | A method on the entity (NOT in the use case) |
| Need to talk to a third-party API | An interface in `domain/`, an implementation in `infrastructure/` |
| Need cross-cutting infra (logger, metrics, auth) | `internal/platform/<thing>/` |

| Commands |
|---|
| `go run ./cmd/migrate up` — apply migrations |
| `go run ./cmd` — start the server |
| `go test ./...` — run all unit tests |
| `./scripts/check-deps.sh` — enforce dependency rule |
| `go build ./...` — compile everything |
| `go vet ./...` — static checks |

---

## Further reading

- **For the philosophy**: *Domain-Driven Design Distilled* by Vaughn Vernon (~150 pages, the gentlest intro).
- **For the patterns**: *Implementing Domain-Driven Design* by Vaughn Vernon (the canonical reference).
- **For the architecture rules**: *Clean Architecture* by Robert C. Martin, chapters 16–22.
- **For Go specifics**: [Domain-Driven Hexagon (Go example)](https://github.com/Sairyss/domain-driven-hexagon) — sibling architecture in Go.

---

**You should now be able to:** open this repo cold, pick a new feature off the brief, and know exactly which file to create and what goes in it. If a decision isn't covered here, that's a signal to update this doc.
