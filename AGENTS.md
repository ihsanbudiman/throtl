# PROJECT KNOWLEDGE BASE

**Generated:** 2026-05-21
**Commit:** 63ef82e
**Branch:** main

## OVERVIEW
API gateway for sharing AI API access without sharing keys. Go backend (Echo) + React frontend (Vite + Tailwind v4). SQLite persistence.

## STRUCTURE
```
throtl/
├── backend/           # Go API gateway server (Echo v4, JWT auth, SQLite)
├── frontend/          # React admin dashboard (Vite 8, Tailwind v4, shadcn/ui)
└── docker-compose.yml # Two-service orchestration
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add API route | `backend/cmd/server/main.go` + `backend/internal/handler/handler.go` | Route defs in main, handlers in handler |
| Add DB table/query | `backend/internal/store/store.go` + `backend/internal/model/model.go` | Schema in store.migrate(), types in model |
| Add provider type | `backend/internal/proxy/adapter.go` (factory) + new adapter file | Implement ProviderAdapter interface |
| Add frontend page | `frontend/src/pages/` + register in `App.tsx` + add to `Sidebar.tsx` | Lazy-loaded via React.lazy() |
| Add API call | `frontend/src/lib/api.ts` | Single fetch wrapper, all endpoints |
| Change design tokens | `frontend/src/index.css` @theme block | Tailwind v4 CSS-based config, no tailwind.config.js |
| Change rate limit logic | `backend/internal/middleware/ratelimit.go` | Triple limits: rolling window + daily requests + daily token in/out |
| Change proxy/streaming | `backend/internal/proxy/` | Anthropic↔OpenAI stream transforms, responses.go (747 lines) for OpenAI Responses API |
| Add chart/widget | `frontend/src/pages/UsagePage.tsx` | Recharts, stacked bar charts, custom dropdowns |
| Change token limits | `backend/internal/middleware/ratelimit.go` | Token in/out daily limits added to rate limiter context |

## CONVENTIONS
- **Go**: Standard `cmd/` + `internal/` layout. Single binary. `CGO_ENABLED=0` static build.
- **Frontend**: `@/` path alias. `import type` mandatory (verbatimModuleSyntax). Dark-first design.
- **Docker**: Multi-stage builds. `docker compose up -d` is the only run command.
- **Env vars**: `THROTL_PORT`, `THROTL_DB_URL`, `THROTL_JWT_SECRET` (auto-generated if unset)
- **Charts**: Recharts v3, CSS variables for theme support, custom tooltips/legends

## ANTI-PATTERNS (THIS PROJECT)
- **No CI/CD** — no GitHub Actions, no lint gate, no automated builds
- **Default JWT secret** in docker-compose — MUST override in production
- **Backend .dockerignore has `shareai-server`** — naming remnant from rename, harmless but inconsistent
- **Error swallowing** — 72+ `_ =` blank identifiers in Go code silently discard errors
- **God files** — responses.go (747), anthropic.go (698), handler.go (613), transform.go (585), store.go (488)
- **No DB transactions** — store does individual queries, no BEGIN/COMMIT blocks

## UNIQUE STYLES
- Design tokens use evocative names: `canvas`, `ink`, `body`, `hairline`, `accent-sunset/dusk/twilight/breeze/midnight`
- All radii: `8px` (single `--radius` variable)
- Custom animations: `.pulse-dot`, `.fade-in-up`, `.shimmer`, `.skeleton`, `.fade-in-stagger` (cascading children)
- SQLite via `modernc.org/sqlite` (pure Go, no CGO) — not `mattn/go-sqlite3`
- Provider model format: `provider-id/model-name` (e.g. `wafer/GLM-5.1`)
- Share key prefix: `sk-share-...`

## COMMANDS
```bash
# Full stack
docker compose up -d

# Backend only
cd backend && go run ./cmd/server           # :8080

# Frontend only
cd frontend && npm run dev                  # :5173 (proxies :8080)

# Frontend build
cd frontend && npm run build                # tsc -b && vite build

# Frontend lint
cd frontend && npm run lint                  # eslint

# Frontend type check
cd frontend && npm run type-check           # tsc --noEmit

# Frontend tests
cd frontend && npm run test                 # vitest run

# Backend tests
cd backend && go test ./...
```

## TESTS
- **Backend**: 4 `_test.go` files — `ratelimit_test.go` (270 lines), `handler_test.go`, `store_test.go`, `store_token_test.go` (9 tests for token limits)
- **Frontend**: 3 `.test.tsx` files — `KeysPage.test.tsx`, `GenerateKeyDialog.test.tsx`, `setup.test.ts`
- **Config**: Vitest (jsdom, globals), `frontend/src/test/setup.ts` imports `@testing-library/jest-dom`

## NOTES
- Frontend dev proxy (`vite.config.ts`) points to `localhost:8080`, but Docker maps backend to `8081` — use `go run` for local dev, not Docker backend
- nginx `/v1/` proxy has 300s timeout + `proxy_buffering off` for AI streaming
- No Prettier, no EditorConfig, no Makefile, no precommit hooks
- No Go linter config (`.golangci.yml`) — relies on `go vet` only
