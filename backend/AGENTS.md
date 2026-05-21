# Backend

Go API gateway server. Echo v4, JWT auth, SQLite persistence.

## STRUCTURE

```
backend/
├── cmd/server/main.go    # Entry point + all route definitions
├── internal/
│   ├── handler/           # All HTTP handlers (handler.go ~613 lines)
│   ├── middleware/         # JWT auth, share-key auth, rate limiter (token limits)
│   ├── model/model.go     # All domain types
│   ├── proxy/             # ProviderAdapter strategy, adapters, stream transforms, Responses API
│   └── store/             # All DB access (store.go ~488 lines), schema via migrate()
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add route | `cmd/server/main.go` | Routes defined inline, not in handler |
| Add handler | `internal/handler/handler.go` | Single file, all handlers |
| Add DB table/query | `internal/store/store.go` | Schema in `migrate()`, all CRUD in one file |
| Add domain type | `internal/model/model.go` | Structs only |
| Add provider adapter | `internal/proxy/adapter.go` | Factory `NewAdapter()`, then new adapter file |
| Change rate limit logic | `internal/middleware/ratelimit.go` | Triple limits: rolling window + daily requests + daily token in/out |
| Change stream format | `internal/proxy/transform.go` | Bidirectional OpenAI↔Anthropic conversion (585 lines) |
| Change Responses API | `internal/proxy/responses.go` | OpenAI Responses API ↔ Chat Completions conversion (747 lines) |
| Change auth logic | `internal/middleware/` | KeyAuth reads `Authorization: Bearer` or `x-api-key` |

## CONVENTIONS

- **Build**: `CGO_ENABLED=0 GOOS=linux go build -o /throtl-server ./cmd/server`
- **SQLite**: `modernc.org/sqlite` (pure Go), WAL mode enabled
- **JWT**: 72-hour expiry
- **ProviderAdapter**: interface in `proxy/adapter.go`, factory `NewAdapter()`, per-provider files
- **Anthropic adapter** (`anthropic.go`, 698 lines): converts Anthropic stream events to OpenAI SSE format
- **Rate limiter**: triple limits, rolling window (N per X hours) + daily requests + daily token in/out (resets midnight UTC)
- **Auth**: `KeyAuth` middleware accepts both `Authorization: Bearer` and `x-api-key` header

## ANTI-PATTERNS

- **God files** — responses.go (747), anthropic.go (698), handler.go (613), transform.go (585), store.go (488)
- **No transaction usage** — store does individual queries, no `BEGIN`/`COMMIT` blocks
- **Route defs in main** — adding a handler requires editing two files (main.go + handler.go)
- **Error swallowing** — 72+ `_ =` blank identifiers silently discard errors across codebase
- **Silent provider fallback** — `NewAdapter()` defaults to OpenAI for unknown provider types

## TESTS

- `internal/middleware/ratelimit_test.go` — 270 lines, comprehensive rate limiter tests
- `internal/handler/handler_test.go` — Handler tests
- `internal/store/store_test.go` — Store tests
- `internal/store/store_token_test.go` — Token store tests
- `internal/store/store_test_helpers.go` — In-memory SQLite fixtures
