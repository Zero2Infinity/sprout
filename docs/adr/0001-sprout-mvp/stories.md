# User Stories: Sprout MVP

## Problem

Existing CLI chat tools for local LLMs are either complex to configure or require cloud API keys. Developers working with Ollama need a simple, local-first chat interface that just works — no accounts, no config files to hunt for, no internet required.

## Goals

1. **Zero-setup local chat** — Run the binary, talk to your Ollama model
2. **Developer-friendly** — All data stays in the repo for easy inspection
3. **Fast feedback** — Token-by-token streaming so you see responses as they generate
4. **Session continuity** — Pick up where you left off without copy-pasting
5. **Learning foundation** — Clean Go codebase that's easy to extend with tools, providers, or new features

**Epic:** Build a minimal CLI chat app for local Ollama models
**Target:** Working MVP with streaming, syntax highlighting, history, clipboard

---

## Epic 1: Configuration & Setup

### Story 1.0: Initialize Go Module
**As a** developer, **I want** a `go.mod` with all dependencies declared
**So that** the project builds and `go mod tidy` is clean

**Acceptance Criteria:**
- [ ] `go.mod` exists with module name `github.com/user/sprout` (or similar)
- [ ] All dependencies from ADR declared: bubbletea, bubbles, lipgloss, cobra, viper, openai-go, chroma, clipboard, uuid
- [ ] `go mod tidy` passes with no errors

**Files:** `go.mod`, `go.sum`

---

### Story 1.1: Load JSON Config
**As a** user, **I want** the app to load config from `config/config.json` in the repo
**So that** I don't have to specify model/endpoint every time

**Acceptance Criteria:**
- [ ] Loads config from `config/config.json` (repo-local)
- [ ] Falls back to defaults if file missing
- [ ] Returns clear error if file exists but is malformed JSON
- [ ] Supports env var overrides (`OLLAMA_BASE_URL`, `OLLAMA_MODEL`)
- [ ] Default model: `qwen3:27b`
- [ ] Default endpoint: `http://localhost:11434/v1`

**Files:** `config/config.go`

---

### Story 1.2: Create Data Directory
**As a** user, **I want** sessions saved automatically
**So that** I can resume conversations later

**Acceptance Criteria:**
- [ ] Creates `sessions/` directory in repo on first run
- [ ] Handles permission errors gracefully
- [ ] Creates `config/` directory in repo if missing

**Files:** `config/config.go`, `message/store.go`

---

### Story 1.3: CLI Entry Point
**As a** developer, **I want** a `main.go` with Cobra root command
**So that** the binary can be built and run with flags

**Acceptance Criteria:**
- [ ] `main.go` initializes Cobra root command
- [ ] Root command has `--session <id>` flag to resume a specific session
- [ ] Root command has `--model <name>` flag to override config model
- [ ] Root command has `--endpoint <url>` flag to override config endpoint
- [ ] Runs config loading and data dir creation before TUI launch
- [ ] `go build ./...` succeeds

**Files:** `main.go`

---

## Epic 2: Ollama Provider

### Story 2.1: Connect to Ollama
**As a** user, **I want** the app to connect to my local Ollama instance
**So that** I can chat with local models

**Acceptance Criteria:**
- [ ] Uses OpenAI-compatible client with `baseURL` from config
- [ ] Creates chat completion request with model from config
- [ ] Returns a streaming iterator of response chunks
- [ ] Error handling delegated to Story 8.2

**Files:** `provider/ollama.go`

---

### Story 2.2: Stream Response Token-by-Token
**As a** user, **I want** to see responses as they're generated
**So that** I get immediate feedback

**Acceptance Criteria:**
- [ ] Streams tokens via OpenAI streaming API
- [ ] Emits `ContentDelta` events per token
- [ ] Emits `Complete` event with token usage
- [ ] Handles cancellation via context
- [ ] Shows spinner during connection

**Files:** `provider/ollama.go`, `agent/loop.go`

---

### Story 2.3: System Prompt
**As a** user, **I want** a default system prompt so the model knows it's in a CLI chat
**So that** responses are appropriately formatted

**Acceptance Criteria:**
- [ ] Default system prompt: "You are a helpful assistant in a terminal chat interface. Respond concisely. Format code with markdown fences."
- [ ] Prompt is prepended to every conversation
- [ ] Configurable via `"systemPrompt"` field in config JSON
- [ ] Not exposed as a CLI flag for MVP

**Files:** `config/config.go`, `provider/ollama.go`

---

### Story 2.4: Message Store
**As a** developer, **I want** a message store API
**So that** sessions and the TUI have a clean interface for message operations

**Acceptance Criteria:**
- [ ] `Message` struct: `Role` (user/assistant/system), `Content` string, `Tokens` int, `Timestamp` time.Time
- [ ] `Store` methods: `Add(msg)`, `All() []Message`, `LastAssistant() *Message`, `Clear()`
- [ ] Store is in-memory; persistence handled by session layer
- [ ] Thread-safe (mutex) since TUI and agent loop may access concurrently

**Files:** `message/store.go`

---

## Epic 3: Session Management

### Story 3.1: Create New Session
**As a** user, **I want** a new session created on start
**So that** my conversation is tracked

**Acceptance Criteria:**
- [ ] Generates UUID for session ID
- [ ] Records model, timestamp
- [ ] Saves to JSON file in `sessions/` (repo-local)
- [ ] Shows session ID in header
- [ ] This is the default behavior (no `--session` flag)

**Files:** `session/session.go`, `message/store.go`

---

### Story 3.2: Load Existing Session
**As a** user, **I want** to resume a previous session with `--session <id>`
**So that** I don't lose context

**Acceptance Criteria:**
- [ ] `--session <id>` flag loads session from `sessions/<id>.json`
- [ ] Restores message history and token counts into message store
- [ ] Restores prompt history for Up/Down navigation
- [ ] Shows error if session ID not found
- [ ] Without `--session`, always creates a new session (Story 3.1)

**Files:** `session/session.go`, `message/store.go`

---

### Story 3.3: Save Session on Exit
**As a** user, **I want** my session saved when I quit
**So that** I can resume later

**Acceptance Criteria:**
- [ ] Saves on `Ctrl+C`
- [ ] Saves message history
- [ ] Saves token usage totals
- [ ] Saves prompt history

**Files:** `session/session.go`, `message/store.go`

---

## Epic 4: TUI Layout

### Story 4.1: Display Header
**As a** user, **I want** to see current context at a glance
**So that** I know which model/session/directory I'm in

**Acceptance Criteria:**
- [ ] Shows current working directory
- [ ] Shows session ID (truncated)
- [ ] Shows model name
- [ ] Fixed at top, doesn't scroll

**Files:** `tui/header.go`

---

### Story 4.2: Display Chat Messages
**As a** user, **I want** to see my conversation
**So that** I can read the full context

**Acceptance Criteria:**
- [ ] Shows user and assistant messages
- [ ] Scrolls automatically to bottom on new message
- [ ] Supports mouse/keyboard scrolling
- [ ] Word-wraps long lines
- [ ] Shows placeholder text when no messages ("Start a conversation...")

**Files:** `tui/chat.go`

---

### Story 4.3: Display Footer
**As a** user, **I want** to see token usage and key hints
**So that** I know my budget and available actions

**Acceptance Criteria:**
- [ ] Shows current message tokens / total session tokens
- [ ] Updates in real-time during streaming
- [ ] Shows key hints (Enter, Esc, Ctrl+C)
- [ ] Fixed at bottom, doesn't scroll

**Files:** `tui/footer.go`

---

### Story 4.4: Text Input Area
**As a** user, **I want** a multi-line input area
**So that** I can compose complex prompts

**Acceptance Criteria:**
- [ ] Textarea at bottom of screen
- [ ] `Enter` sends message
- [ ] `Shift+Enter` adds newline
- [ ] Placeholder text when empty
- [ ] Focus indicator

**Files:** `tui/input.go`

---

### Story 4.5: Bubble Tea Model
**As a** developer, **I want** a root Bubble Tea model that composes all TUI components
**So that** the app has a single Update/View loop

**Acceptance Criteria:**
- [ ] Root model holds: header, chat, footer, input, toast sub-models
- [ ] Routes keyboard events to the correct sub-model
- [ ] Handles window resize messages
- [ ] Manages app state: idle, streaming, cancelled
- [ ] Sends user messages to agent loop on Enter
- [ ] Receives streaming tokens and updates chat + footer

**Files:** `tui/model.go`

---

## Epic 5: Syntax Highlighting

### Story 5.1: Highlight Code Blocks
**As a** user, **I want** code in responses to be syntax-highlighted
**So that** it's easier to read

**Acceptance Criteria:**
- [ ] Detects ```language fences in assistant messages
- [ ] Highlights using chroma with language detection
- [ ] Falls back to plain text if language unknown
- [ ] Uses a dark theme (Tokyo Night / Catppuccin-like)

**Files:** `tui/chat.go`, `tui/styles.go`

---

## Epic 6: History Navigation

### Story 6.1: Navigate Input History
**As a** user, **I want** to recall previous prompts with Up/Down arrows
**So that** I can reuse or edit them

**Acceptance Criteria:**
- [ ] `Up` arrow loads previous prompt into textarea
- [ ] `Down` arrow loads next prompt
- [ ] History persisted in session JSON
- [ ] Doesn't modify history until sent

**Files:** `tui/input.go`, `session/session.go`

---

## Epic 7: Clipboard Copy

### Story 7.1: Copy Message to Clipboard
**As a** user, **I want** to copy the last assistant response
**So that** I can paste it elsewhere

**Acceptance Criteria:**
- [ ] `Ctrl+Shift+C` copies last assistant message
- [ ] Uses `atotto/clipboard` (cross-platform)
- [ ] Shows "Copied!" toast in footer for 1 second
- [ ] Copies plain text (no markdown formatting)

**Files:** `tui/chat.go`, `tui/toast.go`

---

## Epic 8: Streaming & Cancellation

### Story 8.1: Cancel Stream with Esc
**As a** user, **I want** to cancel a streaming response
**So that** I can stop if it's going off track

**Acceptance Criteria:**
- [ ] `Esc` cancels current stream
- [ ] Keeps partial response
- [ ] Returns to input mode
- [ ] Shows "Cancelled" indicator

**Files:** `agent/loop.go`, `tui/model.go`

---

### Story 8.2: Handle Ollama Unavailable
**As a** user, **I want** clear error if Ollama isn't running
**So that** I know how to fix it

**Acceptance Criteria:**
- [ ] Shows error: "Cannot connect to Ollama at http://localhost:11434"
- [ ] Suggests: "Run `ollama serve` to start"
- [ ] Doesn't crash, returns to input

**Files:** `provider/ollama.go`, `tui/model.go`

---

## Implementation Order

### Parallel Tracks

| Track | Agent | Stories | Depends on | Files |
|-------|-------|---------|------------|-------|
| **A: Foundation** | Agent A | 1.0, 1.1, 1.2, 2.4, 1.3 | — | `go.mod`, `main.go`, `config/config.go`, `message/store.go` |
| **B: Provider** | Agent B | 2.1, 2.3, 2.2, 8.2 | A (1.1) | `provider/ollama.go`, `agent/loop.go` |
| **C: Session** | Agent C | 3.1, 3.2, 3.3 | A (1.2, 2.4) | `session/session.go` |
| **D: TUI** | Agent D | 4.1, 4.2, 4.3, 4.4, 5.1, 6.1, 7.1, 8.1 | A (2.4) | `tui/header.go`, `chat.go`, `footer.go`, `input.go`, `toast.go`, `styles.go` |
| **E: Integration** | Agent E | 4.5 | B, C, D | `tui/model.go` |

### Execution Flow

```
Agent A (Foundation)
  ├── Agent B (Provider)    ─┐
  ├── Agent C (Session)     ─┤── Agent E (Integration)
  └── Agent D (TUI)         ─┘
```

Tracks B, C, D run in parallel. E merges everything.

### Effort

| Track | Effort |
|-------|--------|
| A: Foundation | 1-2 hours |
| B: Provider | 1-2 hours |
| C: Session | 1 hour |
| D: TUI | 2-3 hours |
| E: Integration | 30 min |
| **Total (parallel)** | **~4-5 hours** |
| **Total (sequential)** | **~8-13 hours** |

---

## Definition of Done

- [ ] All stories pass acceptance criteria
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] No panics on common error paths
- [ ] Session survives app restart
- [ ] Streaming works with Ollama
- [ ] Syntax highlighting renders code blocks
- [ ] History navigation works
- [ ] Clipboard copy works
