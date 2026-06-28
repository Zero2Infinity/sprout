# ADR-003: Model Thinking Display

**Status:** Proposed
**Date:** 2026-06-28
**Deciders:** Rahul

## Context

The project needs to show model reasoning ("thinking") for qwen3 models. Some models (qwen3, deepseek-r1) produce a `<think>` tag in the content string, but Ollama's native API provides a dedicated `thinking` field in the response — no custom parsing needed.

This feature transforms Sprout from a "black box" chat into a transparent one where users can see the model's reasoning process, making it easier to debug prompts and understand complex responses.

## Decision

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Data source** | Ollama native `think` parameter | Dedicated field; no `<think>` tag parsing |
| **Display** | Inline in chat, dimmed/italic | Stays in conversation flow; visually distinct |
| **Persistence** | Always saved in session JSON | Useful for debugging and review |
| **Toggle** | `t` key during chat | User control without losing data |
| **Default** | Thinking enabled | Full transparency out of the box |

## Detailed Design

### 1. Ollama API

Ollama's `/api/chat` endpoint accepts a `think` parameter. When enabled, the model's reasoning is returned in a dedicated `message.thinking` field — separate from `message.content`.

**Request:**
```json
{
  "model": "qwen3.6:27b",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Explain Go concurrency"}
  ],
  "stream": true,
  "think": true
}
```

**Streaming response chunk (during thinking phase):**
```json
{
  "model": "qwen3.6:27b",
  "message": {
    "role": "assistant",
    "thinking": "The user wants to understand Go concurrency. I should explain goroutines, channels, and the select statement...",
    "content": ""
  },
  "done": false
}
```

**Streaming response chunk (during response phase):**
```json
{
  "model": "qwen3.6:27b",
  "message": {
    "role": "assistant",
    "thinking": "",
    "content": "Go concurrency is built on two core primitives: goroutines and channels..."
  },
  "done": false
}
```

The `thinking` field arrives incrementally during streaming, just like `content`. During the thinking phase, `content` is empty. During the response phase, `thinking` is empty.

### 2. Display

Thinking content renders **inline in the chat**, visually distinct from assistant content:

```
You: Explain Go concurrency

┃ The user wants to understand Go concurrency. I should explain
┃ goroutines, channels, and the select statement...

Assistant: Go concurrency is built on two core primitives: goroutines
and channels. A goroutine is a lightweight thread managed by the Go
runtime...
```

**Styling:**
- Italic font style
- Dimmed color (terminal color `242`)
- Box-drawing prefix (`┃`) for visual separation
- Default: **always visible**

**Toggle:** Press `t` during chat to toggle thinking visibility on/off. The thinking content is always captured and persisted — the toggle only affects display.

### 3. Persistence

Thinking is stored per assistant message in the session JSON:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Explain Go concurrency",
      "thinking": "",
      "timestamp": "2026-06-28T10:00:00Z"
    },
    {
      "role": "assistant",
      "thinking": "The user wants to understand Go concurrency...",
      "content": "Go concurrency is built on two core primitives...",
      "timestamp": "2026-06-28T10:00:01Z"
    }
  ]
}
```

Thinking is always persisted regardless of display toggle state.

### 4. Data Flow

```
User sends message
  → POST /api/chat { "think": true, ... }
  → NDJSON stream: chunks with message.thinking + message.content
  → Provider emits ThinkingDelta + ContentDelta StreamEvents
  → TUI accumulates both, renders thinking dimmed above content
  → On completion: both fields saved to message store → session JSON
```

### 5. Struct Changes

**`provider/provider.go` — StreamEvent:**
```go
type StreamEvent struct {
    ContentDelta  string
    ThinkingDelta string    // NEW: thinking content delta
    Complete      bool
    Usage         *Usage
    Err           error
}
```

**`message/store.go` — Message:**
```go
type Message struct {
    Role      Role
    Content   string
    Thinking  string    // NEW: model's reasoning
    Timestamp time.Time
}
```

**`provider/ollama.go` — chatRequest:**
```go
type chatRequest struct {
    Model    string        `json:"model"`
    Messages []chatMessage `json:"messages"`
    Stream   bool          `json:"stream"`
    Think    bool          `json:"think,omitempty"`  // NEW
}
```

**`provider/ollama.go` — chatMessage (extended):**
```go
type chatMessage struct {
    Role     string `json:"role"`
    Content  string `json:"content"`
    Thinking string `json:"thinking,omitempty"`  // NEW
}
```

### 6. Config Changes

```go
type Config struct {
    // ... existing fields ...
    Think bool `json:"think"`  // NEW: default true
}
```

CLI flag: `--think` (default: true), `--think=false` to disable.

### 7. TUI Changes

| File | Change |
|------|--------|
| `tui/model.go` | Accumulate `thinkingContent` alongside `streamingContent`. Handle `ThinkingDelta` from stream events. Add `t` key toggle. |
| `tui/chat.go` | Add `SetThinkingContent(string)` method. Render thinking block with dimmed style before assistant content. |
| `tui/styles.go` | Add `thinkingStyle` — italic, foreground color `242`, left-padded with `┃` prefix. |

### 8. Files Changed

| File | Change |
|------|--------|
| `provider/provider.go` | Add `ThinkingDelta` to `StreamEvent` |
| `provider/ollama.go` | Add `Think` to request, parse `thinking` from chunks, emit `ThinkingDelta` |
| `message/store.go` | Add `Thinking` field to `Message` |
| `tui/model.go` | Accumulate thinking, handle toggle, pass to chat |
| `tui/chat.go` | Render thinking block inline, dimmed |
| `tui/styles.go` | Add `thinkingStyle` |
| `config/config.go` | Add `Think bool` to Config |
| `main.go` | Add `--think` flag |

## Consequences

- **Positive:** Leverages Ollama-native support — no custom parsing, no prompt engineering
- **Positive:** Thinking always persisted — useful for debugging and review
- **Positive:** Toggle key gives user control over display without losing data
- **Negative:** Adds `Thinking` field to every Message — increases session file size
- **Mitigation:** Thinking content is typically small relative to assistant content

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Parse `` tags from content | Fragile; depends on model output format; Ollama provides dedicated field |
| Separate thinking panel | Takes vertical space; inline is more natural for conversation flow |
| Thinking only on CLI flag | Less discoverable; toggle key is more ergonomic |
| Thinking only persisted when visible | Adds complexity; always-persist is simpler and safer |
