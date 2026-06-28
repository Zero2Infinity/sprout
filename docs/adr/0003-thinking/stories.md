# Stories: ADR-003 Thinking Display

## Problem

Users can't see the model's reasoning process. When qwen3 generates a response, the thinking happens "inside the box" — making it hard to debug prompts or understand complex responses.

## Goals

1. **Transparency** — Show the model's reasoning before its response
2. **Persistence** — Always save thinking for review
3. **Control** — Toggle visibility without losing data

---

## Epic 1: Provider Thinking Support

### Story 1.1: Stream thinking tokens
**As a** developer, **I want** the provider to emit thinking deltas
**So that** the TUI can display reasoning in real-time

**Acceptance Criteria:**
- [ ] `ThinkingDelta string` field on `StreamEvent`
- [ ] `Think bool` field on `chatRequest`
- [ ] `thinking` parsed from `chatStreamChunk.message.thinking`
- [ ] Empty thinking fields don't emit events
- [ ] Thinking streams during thinking phase, content streams during response phase

**Files:** `provider/provider.go`, `provider/ollama.go`

---

## Epic 2: Message Store

### Story 1.2: Message thinking field
**As a** developer, **I want** messages to store thinking content
**So that** thinking persists across sessions

**Acceptance Criteria:**
- [ ] `Thinking string` field on `Message` struct
- [ ] JSON tag: `json:"thinking"`
- [ ] `All()` and `LastAssistant()` return thinking
- [ ] `Add()` preserves thinking field

**Files:** `message/store.go`

---

## Epic 3: Config

### Story 1.3: Thinking config
**As a** user, **I want** to control thinking via config and CLI
**So that** I can disable thinking if needed

**Acceptance Criteria:**
- [ ] `Think bool` field on `Config` struct
- [ ] Default value: `true`
- [ ] Config file: `"think": true`
- [ ] CLI flag: `--think` / `--think=false`
- [ ] Env var: `SPROUT_THINK=false`

**Files:** `config/config.go`, `main.go`

---

## Epic 4: TUI Display

### Story 1.4: Thinking style
**As a** user, **I want** thinking to look different from assistant content
**So that** I can distinguish reasoning from response

**Acceptance Criteria:**
- [ ] `thinkingStyle` in `tui/styles.go`
- [ ] Italic font
- [ ] Dimmed color (terminal color `242`)
- [ ] Box-drawing prefix: `┃`
- [ ] Indented from left edge

**Files:** `tui/styles.go`

---

### Story 1.5: Thinking rendering
**As a** user, **I want** thinking to appear inline above assistant content
**So that** I see reasoning before the response

**Acceptance Criteria:**
- [ ] `SetThinkingContent(string)` on `ChatModel`
- [ ] Thinking block rendered before assistant content
- [ ] Thinking shown during streaming (real-time)
- [ ] Thinking shown in message history (persisted)

**Files:** `tui/chat.go`

---

### Story 1.6: Thinking toggle
**As a** user, **I want** to toggle thinking visibility
**So that** I can hide reasoning when I want a cleaner view

**Acceptance Criteria:**
- [ ] `t` key toggles `thinkingVisible bool` on TUI model
- [ ] When hidden, thinking block not rendered in chat
- [ ] Thinking always persisted regardless of toggle
- [ ] Footer shows thinking state: "T: on" / "T: off"

**Files:** `tui/model.go`, `tui/chat.go`, `tui/footer.go`

---

## Definition of Done

- [ ] Thinking streams in real-time during model response
- [ ] Thinking displayed dimmed above assistant content
- [ ] Thinking persisted in session JSON
- [ ] `t` key toggles visibility
- [ ] `--think=false` disables thinking
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
