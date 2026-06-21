# ADR-002: Provider Metrics & Token Usage Instrumentation

**Status:** Accepted  
**Date:** 2026-06-21  
**Deciders:** Rahul  

## Context

ADR-001 established per-request token tracking via OpenAI-compatible `Usage` objects (`prompt_tokens`, `completion_tokens`, `total_tokens`). However, Ollama's native API (`/api/chat`, `/api/generate`) returns richer per-request metrics not available through the OpenAI-compatible path:

- `total_duration` — wall-clock time for the entire request
- `load_duration` — time spent loading the model into memory
- `prompt_eval_duration` — time spent evaluating the prompt (prefill)
- `eval_duration` — time spent generating output tokens

These enable actionable insights: throughput (tokens/sec), prefill vs decode breakdown, and session-level cost estimation. We want to capture these now in an Ollama-specific implementation, with a design that generalizes to other providers later.

## Decision

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Scope** | Token counts + timing fields | Tokens alone skip performance insights |
| **Data source** | Ollama native `/api/chat` raw HTTP | OpenAI-compatible API doesn't expose timing via `openai-go` SDK |
| **Display** | TUI footer + session summary on exit | Footer shows live; exit prints final summary |
| **Abstraction** | Unified display, zero for absent fields | Avoid premature abstraction; one display template |
| **Data flow** | Extend `provider.Usage` with timing fields | One struct, no extra channels |
| **HTTP /metrics** | Out of scope for MVP | Can add in future ADR |

## Detailed Design

### 1. Extended `provider.Usage` struct

```go
type Usage struct {
    PromptTokens     int   `json:"promptTokens"`
    CompletionTokens int   `json:"completionTokens"`
    TotalTokens      int   `json:"totalTokens"`

    // Timing (Ollama-specific, zero for other providers)
    TotalDurationNs      int64 `json:"totalDurationNs,omitempty"`
    LoadDurationNs       int64 `json:"loadDurationNs,omitempty"`
    PromptEvalDurationNs int64 `json:"promptEvalDurationNs,omitempty"`
    EvalDurationNs       int64 `json:"evalDurationNs,omitempty"`
}

// TokensPerSec returns completion tokens per second, or 0 if no eval timing.
func (u Usage) TokensPerSec() float64 {
    if u.EvalDurationNs == 0 || u.CompletionTokens == 0 {
        return 0
    }
    return float64(u.CompletionTokens) / float64(u.EvalDurationNs) * 1e9
}
```

### 2. Ollama provider sources timing from native API

Replace `openai-go` streaming with raw HTTP streaming to **Ollama's native `/api/chat`** endpoint. The native API returns both content deltas and the rich metrics in the final `done: true` chunk.

The provider interface stays the same (`ChatStream` → `StreamEvent`), so other providers can implement with their own API.

**Native `/api/chat` request:**
```json
{
  "model": "qwen3.6:27b",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}
```

**Native `/api/chat` streaming response (final chunk):**
```json
{
  "model": "qwen3.6:27b",
  "message": {"role": "assistant", "content": ""},
  "done": true,
  "total_duration": 2145000000,
  "load_duration": 50000000,
  "prompt_eval_count": 150,
  "prompt_eval_duration": 245000000,
  "eval_count": 2450,
  "eval_duration": 1995000000
}
```

**SSE parsing rules:**
- Each line is raw JSON (NDJSON format) — no `data:` prefix
- Empty lines are ignored
- `done: true` signals the final chunk with full metrics

### 3. Unified display in TUI footer

```
Prompt: 150 | Completion: 2450 | Context: 2600/32768 | Speed: 12.3 tok/s | Latency: 2.1s
```

When `TokensPerSec() > 0`, show Speed and Latency. When zero (other providers), show only token counts as today.

### 4. Session summary on exit

After TUI quits, print:
```
Session saved: abc123-...
Prompt tokens: 150 | Completion: 2450 | Total: 2600
Latency: 2.1s | Speed: 12.3 tok/s
To resume: sprout --session abc123-...
```

### 5. Session persistence

Extend `session.TokenUsage` with timing fields and accumulate across the session:

```json
{
  "tokenUsage": {
    "promptTokens": 150,
    "completionTokens": 2450,
    "totalTokens": 2600,
    "totalDurationNs": 2100000000,
    "evalDurationNs": 1995000000
  }
}
```

## Files Changed

| File | Change |
|------|--------|
| `provider/provider.go` | Extend `Usage` struct with timing fields; add `TokensPerSec()` |
| `provider/ollama.go` | Switch from `openai-go` streaming → raw HTTP Ollama native NDJSON stream; populate timing |
| `session/session.go` | Extend `TokenUsage` with timing fields; update `UpdateTokenUsage` to accumulate |
| `tui/footer.go` | Add duration/tok/s display |
| `main.go` | Print summary line after TUI exits |
| `go.mod` | Remove `openai-go` dependency (no longer used) |

## Consequences

- **Positive:** Richer per-request and per-session performance metrics
- **Positive:** Provider interface unchanged — future providers populate different `Usage` fields
- **Positive:** No new dependencies (raw HTTP + stdlib JSON)
- **Negative:** Lose `openai-go` convenience (types, retries) — must handle raw NDJSON
- **Mitigation:** Ollama NDJSON is simple: raw JSON per line, ~60 lines of parsing code

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Keep `openai-go` + parse raw JSON extensions | `openai-go` doesn't expose raw response body for extension fields |
| Dual API calls (OpenAI stream + native fetch) | Wasteful — two requests per message |
| Postpone to multi-provider ADR | Now is the right time — provider/Usage already exists and needs extension |
