# PROXY

AI provider request proxying and bidirectional format transformation between OpenAI and Anthropic APIs.

## WHERE TO LOOK

| Task | File | Notes |
|------|------|-------|
| Add new provider | `adapter.go` + new adapter file | Implement `ProviderAdapter` (3 methods: `ProxyChat`, `ListModels`, `FetchModels`), add case to `NewAdapter()` switch |
| Change dispatch/routing logic | `proxy.go` | `Gateway` struct + `ProxyHandler` dispatches to adapter by provider type |
| Change model list handling | `proxy.go` | `ListModels` parses `provider-id/model-name` format |
| Change OpenAI streaming | `openai.go` | Passthrough, streams SSE as-is, no format conversion |
| Change Anthropic conversion | `anthropic.go` + `transform.go` | `anthropic.go`: adapter logic. `transform.go`: format conversion functions |
| Change Responses API handler | `responses.go` | OpenAI Responses API ↔ Chat Completions conversion (747 lines). Handles input_text, web_search, file_search, computer_20260124, code_interpreter python_output |
| Change request transformation | `transform.go` | `transformToAnthropic()` converts OpenAI request → Anthropic Messages API |
| Change response transformation | `transform.go` | `transformOpenAIResponse()` converts Anthropic response → OpenAI format |
| Change stream event mapping | `anthropic.go` | `processAnthropicStreamEvent()` maps Anthropic SSE → OpenAI chunks |

## CONVENTIONS

- `NewAdapter()` factory: `"anthropic"` → `AnthropicAdapter`, default → `OpenAIAdapter`
- Model format everywhere: `provider-id/model-name` (e.g. `wafer/GLM-5.1`)
- OpenAI adapter is pure passthrough — zero format conversion, SSE streamed as-is
- Anthropic adapter handles both streaming and non-streaming, plus tool calls and content blocks
- Stream event mapping (Anthropic → OpenAI): `message_start` → role chunk, `content_block_delta` (`text_delta`) → content delta, `content_block_delta` (`input_json_delta`) → tool call arguments delta, `message_delta` → `finish_reason`, `message_stop` → `[DONE]`
- `transform.go` is bidirectional: request goes OpenAI→Anthropic, response comes Anthropic→OpenAI. Handles `tool_calls`, content blocks, streaming deltas

## ANTI-PATTERNS

- `anthropic.go` (698 lines), `transform.go` (585 lines), `responses.go` (747 lines) are large — consider splitting if adding more providers with similar conversion needs
- Adding a provider without updating `NewAdapter()` silently falls through to OpenAI adapter (wrong for non-OpenAI-compatible APIs)
- `processAnthropicStreamEvent()` event type mapping is fragile — new Anthropic event types will silently drop
- Error swallowing: 30+ `_ =` blank identifiers in transform.go discard JSON extraction errors
