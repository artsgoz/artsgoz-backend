# Artsgoz Backend Repo

gcloud auth

`gcloud auth login`

`gcloud config set project chula-artsgoz-website`

check if env loads correctly
`go run cmd/main.go`

currently it will return `postgres: postgres: empty DSN` if it's correctly configured.

---

# artsgoz-monorepo

Nx monorepo for the artsgoz project, containing the public-facing frontend and the backoffice admin application.

## Tech Stack

- **Framework:** Next.js 16 (App Router)
- **Language:** TypeScript 5.9
- **Styling:** Tailwind CSS 3.4
- **Monorepo:** Nx 22.5
- **Package Manager:** pnpm
- **Testing:** Jest 30 + Playwright
- **Linting:** ESLint 9 + Prettier
- **Commit Convention:** Conventional Commits (enforced via commitlint + husky)

## Project Structure

```
artsgoz-monorepo/
├── apps/
│   ├── artsgoz-frontend/        # Public-facing Next.js app
│   ├── artsgoz-frontend-e2e/    # E2E tests for frontend (Playwright)
│   ├── artsgoz-backoffice/      # Admin backoffice Next.js app
│   └── artsgoz-backoffice-e2e/  # E2E tests for backoffice (Playwright)
├── packages/                    # Shared libraries
├── nx.json                      # Nx workspace configuration
├── pnpm-workspace.yaml          # pnpm workspace definition
└── tsconfig.base.json           # Shared TypeScript config
```

## Prerequisites

- **Node.js** >= 22
- **pnpm** >= 9

## Getting Started

### 1. Install dependencies

```bash
pnpm install
```

### 2. Start development servers

```bash
# All apps
pnpm dev

# Individual apps
pnpm dev:frontend
pnpm dev:backoffice
```

### 3. Open in browser

- **Frontend:** http://localhost:3000
- **Backoffice:** http://localhost:3001

## Available Scripts

| Command | Description |
| --- | --- |
| `pnpm dev` | Start all dev servers |
| `pnpm dev:frontend` | Start frontend dev server |
| `pnpm dev:backoffice` | Start backoffice dev server |
| `pnpm build` | Build all projects |
| `pnpm build:frontend` | Build frontend only |
| `pnpm build:backoffice` | Build backoffice only |
| `pnpm test` | Run unit tests for all projects |
| `pnpm lint` | Lint all projects |
| `pnpm lint:fix` | Lint and auto-fix all projects |
| `pnpm typecheck` | Type-check all projects |
| `pnpm e2e` | Run all E2E tests |
| `pnpm affected:build` | Build only affected projects |
| `pnpm affected:test` | Test only affected projects |
| `pnpm affected:lint` | Lint only affected projects |
| `pnpm graph` | Open the Nx dependency graph |

## Adding Dependencies

Install at the workspace root:

```bash
pnpm add -Dw <package>
```

Install to a specific app:

```bash
pnpm add <package> --filter @org/artsgoz-frontend
pnpm add <package> --filter @org/artsgoz-backoffice
```

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) enforced by commitlint and husky.

**Format:** `<type>(<scope>): <description>`

**Allowed types:**

| Type | Description |
| --- | --- |
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only |
| `style` | Code style (formatting, semicolons, etc.) |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf` | Performance improvement |
| `test` | Adding or updating tests |
| `build` | Build system or external dependencies |
| `ci` | CI/CD configuration |
| `chore` | Other changes that don't modify src or test files |
| `revert` | Reverts a previous commit |

**Examples:**

```
feat(frontend): add user login page
fix(backoffice): resolve sidebar navigation bug
docs: update README with setup instructions
chore: upgrade Next.js to 16.0.1
```

## Git Hooks

Husky runs the following hooks automatically:

- **pre-commit** — runs `pnpm test` before each commit
- **commit-msg** — validates commit message format via commitlint

## Contributing

1. Create a new branch from `main`:
   ```bash
   git checkout -b feat/your-feature
   ```

2. Make your changes and ensure all checks pass:
   ```bash
   pnpm lint
   pnpm typecheck
   pnpm test
   ```

3. Commit using the conventional commit format:
   ```bash
   git commit -m "feat(frontend): add new feature"
   ```

4. Push and open a pull request against `main`.

## Nx Commands

Run any Nx target on a specific project:

```bash
pnpm nx <target> <project>

# Examples
pnpm nx build @org/artsgoz-frontend
pnpm nx test @org/artsgoz-backoffice
pnpm nx lint @org/artsgoz-frontend
```

Generate a new app or library:

```bash
pnpm nx g @nx/next:app my-app
pnpm nx g @nx/js:lib my-lib
```

## Recommended VS Code Extensions

- [Nx Console](https://marketplace.visualstudio.com/items?itemName=nrwl.angular-console)
- [Prettier](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode)
- [ESLint](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint)
- [Playwright Test](https://marketplace.visualstudio.com/items?itemName=ms-playwright.playwright)
- [Jest Runner](https://marketplace.visualstudio.com/items?itemName=firsttris.vscode-jest-runner)
