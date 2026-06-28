# Implementation Plan: Thinking, Tools, JSONL

Three features implemented in **3 separate phases** (ADRs, branches, PRs).

## Overview

| Phase | Feature | ADR | Branch | Estimated Effort |
|-------|---------|-----|--------|-----------------|
| 1 | Thinking capability | ADR-003 | `ai/thinking` | 2-3 hours |
| 2 | Tool calling (agentic loop) | ADR-004 | `ai/tools` | 4-5 hours |
| 3 | JSONL session storage | ADR-005 | `ai/jsonl-sessions` | 1-2 hours |

**Total estimated:** 7-10 hours (parallel), ~13-17 hours (sequential)

---

## Phase 1: Thinking Capability

### Stories

#### Story 1.1: Provider thinking support
**As a** developer, **I want** the provider to parse thinking from Ollama chunks
**So that** thinking content is available to the TUI

**Acceptance Criteria:**
- [ ] `ThinkingDelta string` added to `StreamEvent`
- [ ] `Think bool` added to `chatRequest`
- [ ] `thinking` field parsed from `chatStreamChunk` and emitted as `ThinkingDelta`
- [ ] Empty thinking field does not emit event

**Files:** `provider/provider.go`, `provider/ollama.go`

---

#### Story 1.2: Message store thinking field
**As a** developer, **I want** the Message struct to support thinking
**So that** thinking is persisted in sessions

**Acceptance Criteria:**
- [ ] `Thinking string` added to `Message` struct
- [ ] `All()` and `LastAssistant()` return thinking field
- [ ] JSON serialization includes thinking

**Files:** `message/store.go`

---

#### Story 1.3: Config and CLI flag
**As a** user, **I want** to control thinking via config and CLI
**So that** I can disable thinking if needed

**Acceptance Criteria:**
- [ ] `Think bool` added to `Config` (default `true`)
- [ ] `--think` / `--think=false` CLI flag
- [ ] Config file supports `"think": false`
- [ ] Env var override: `SPROUT_THINK=false`

**Files:** `config/config.go`, `main.go`

---

#### Story 1.4: TUI thinking display
**As a** user, **I want** to see model thinking inline in chat
**So that** I can understand the reasoning process

**Acceptance Criteria:**
- [ ] `thinkingStyle` added to `tui/styles.go` (italic, color 242)
- [ ] `SetThinkingContent()` method on `ChatModel`
- [ ] Thinking block rendered with `┃` prefix, dimmed
- [ ] Thinking appears above assistant content

**Files:** `tui/styles.go`, `tui/chat.go`

---

#### Story 1.5: TUI thinking accumulation
**As a** developer, **I want** the TUI model to accumulate thinking tokens
**So that** thinking is displayed in real-time during streaming

**Acceptance Criteria:**
- [ ] `thinkingContent string` field on TUI `Model`
- [ ] `ThinkingDelta` events accumulate in `thinkingContent`
- [ ] `thinkingContent` passed to `ChatModel.SetThinkingContent()`
- [ ] `thinkingContent` cleared on stream complete

**Files:** `tui/model.go`

---

#### Story 1.6: Thinking toggle
**As a** user, **I want** to toggle thinking visibility with a key
**So that** I can hide thinking when I want a cleaner view

**Acceptance Criteria:**
- [ ] `t` key toggles `thinkingVisible bool`
- [ ] When hidden, thinking block not rendered
- [ ] Thinking always persisted regardless of toggle state
- [ ] Visual indicator in footer shows thinking state

**Files:** `tui/model.go`, `tui/chat.go`, `tui/footer.go`

---

### Task Dependency Graph (Phase 1)

```
1.1 Provider thinking ──┐
                       ├── 1.5 TUI accumulation ── 1.6 Toggle
1.2 Message store ──────┘         │
                                  │
1.3 Config + CLI ──────────────── │
                                  │
1.4 TUI display ─────────────────┘
```

**Parallel tracks:**
- Track A: 1.1 → 1.5 → 1.6
- Track B: 1.2 (parallel with A)
- Track C: 1.3 (parallel with A, B)
- Track D: 1.4 (parallel with A, B, C)

---

## Phase 2: Tool Calling

### Stories

#### Story 2.1: Tool interface and types
**As a** developer, **I want** a Tool interface and ToolCall types
**So that** tools have a standard contract

**Acceptance Criteria:**
- [ ] `Tool` interface: `Name()`, `Description()`, `Parameters()`, `Execute()`
- [ ] `ToolCall`, `ToolCallFunction` structs
- [ ] `ToolCalls`, `ToolCallRequest`, `ToolCallResult`, `ToolCallError` on `StreamEvent`
- [ ] `RoleTool` constant, `ToolCalls`, `ToolName` on `Message`

**Files:** `tools/tool.go`, `provider/provider.go`, `message/store.go`

---

#### Story 2.2: Tool registry
**As a** developer, **I want** a tool registry
**So that** tools are discovered and converted to Ollama format

**Acceptance Criteria:**
- [ ] `Registry` struct with `Register()`, `Get()`, `List()`
- [ ] `ToOllamaTools()` converts to Ollama API format
- [ ] Default registry includes all 9 tools

**Files:** `tools/registry.go`

---

#### Story 2.3: Shell tool
**As a** user, **I want** to execute shell commands via the model
**So that** I can run git, npm, build commands, etc.

**Acceptance Criteria:**
- [ ] `execute_shell` tool with `command` and optional `workdir` parameters
- [ ] Captures stdout and stderr
- [ ] Timeout after 30 seconds
- [ ] Returns exit code

**Files:** `tools/shell.go`

---

#### Story 2.4: File tools
**As a** user, **I want** to read, write, and edit files via the model
**So that** I can work with code files

**Acceptance Criteria:**
- [ ] `read_file`: reads file contents, returns as string
- [ ] `write_file`: writes content to file, creates parent dirs
- [ ] `edit_file`: finds and replaces string in file
- [ ] Error handling for missing files, permission errors

**Files:** `tools/file.go`

---

#### Story 2.5: Filesystem tools
**As a** user, **I want** to list, find, and search files
**So that** I can explore the codebase

**Acceptance Criteria:**
- [ ] `list_directory`: lists directory contents with optional path
- [ ] `find_files`: finds files by glob pattern
- [ ] `search_in_files`: greps for regex pattern in files
- [ ] Respects `.gitignore` for find/search

**Files:** `tools/fs.go`

---

#### Story 2.6: Web tools
**As a** user, **I want** to fetch URLs and search the web
**So that** I can look up documentation

**Acceptance Criteria:**
- [ ] `fetch_url`: fetches URL, returns text/markdown/html
- [ ] `web_search`: searches via DuckDuckGo instant answer API
- [ ] Timeout after 10 seconds
- [ ] Max response size limit

**Files:** `tools/web.go`

---

#### Story 2.7: Provider tool support
**As a** developer, **I want** the provider to send tools and parse tool_calls
**So that** Ollama can request tool execution

**Acceptance Criteria:**
- [ ] `Tools` field added to `chatRequest`
- [ ] `tool_calls` parsed from `chatStreamChunk`
- [ ] `ToolCalls` emitted in `StreamEvent`
- [ ] Empty tool_calls does not emit event

**Files:** `provider/ollama.go`

---

#### Story 2.8: Agentic loop
**As a** developer, **I want** a multi-turn agentic loop
**So that** the model can call multiple tools in sequence

**Acceptance Criteria:**
- [ ] `AgenticLoop()` method on `Loop`
- [ ] Multi-turn: model calls tool → result fed back → model continues
- [ ] Max turns configurable (default 5)
- [ ] Approval callback for user consent
- [ ] Esc cancels entire loop

**Files:** `agent/loop.go`

---

#### Story 2.9: TUI tool display
**As a** user, **I want** to see tool calls and results inline
**So that** I have full context of what happened

**Acceptance Criteria:**
- [ ] `AddToolCall()` renders: `🔧 tool_name({args})`
- [ ] `AddToolResult()` renders: tool output
- [ ] Tool calls shown during streaming
- [ ] Tool results shown after execution

**Files:** `tui/chat.go`

---

#### Story 2.10: TUI approval prompt
**As a** user, **I want** to approve or reject tool calls
**So that** I control what executes on my system

**Acceptance Criteria:**
- [ ] `stateToolApproval` state added to TUI
- [ ] Input area shows: `Execute tool_name? [y/n/e(dit)]`
- [ ] `y` executes, `n` rejects, `e` shows editable args
- [ ] `e` mode: input shows JSON args, Enter confirms

**Files:** `tui/model.go`, `tui/input.go`

---

#### Story 2.11: Config and CLI flags
**As a** user, **I want** to control tool settings
**So that** I can configure the agentic behavior

**Acceptance Criteria:**
- [ ] `ToolsEnabled bool` in config (default `true`)
- [ ] `MaxTurns int` in config (default `5`)
- [ ] `--tools` / `--tools=false` CLI flag
- [ ] `--max-turns=N` CLI flag

**Files:** `config/config.go`, `main.go`

---

#### Story 2.12: Wire TUI to agentic loop
**As a** developer, **I want** the TUI to use the agentic loop
**So that** tool calls work end-to-end

**Acceptance Criteria:**
- [ ] TUI calls `AgenticLoop()` instead of `SendMessage()`
- [ ] Tool call events routed to display and approval
- [ ] Tool results fed back to loop
- [ ] Session saved after each tool exchange

**Files:** `tui/model.go`

---

### Task Dependency Graph (Phase 2)

```
2.1 Tool interface + types ──┐
                             ├── 2.2 Registry
2.3-2.6 Individual tools ────┘         │
                                       │
2.7 Provider tool support ── 2.8 Agentic loop ── 2.12 Wire TUI
                                       │
2.9 TUI tool display ────── 2.10 Approval prompt
                                       │
2.11 Config + CLI ─────────────────────┘
```

**Parallel tracks:**
- Track A: 2.1 → 2.2 → 2.3-2.6 (parallel)
- Track B: 2.7 → 2.8 → 2.12
- Track C: 2.9 → 2.10
- Track D: 2.11 (parallel with all)

---

## Phase 3: JSONL Session Storage

### Stories

#### Story 3.1: JSONL save
**As a** developer, **I want** sessions saved in JSONL format
**So that** saves are efficient and diffs are clean

**Acceptance Criteria:**
- [ ] `Save()` writes `.jsonl` file
- [ ] Line 1: metadata (id, model, timestamps, tokenUsage)
- [ ] Lines 2+: messages (one JSON object per line)
- [ ] File extension: `.jsonl`

**Files:** `session/session.go`

---

#### Story 3.2: JSONL load
**As a** developer, **I want** sessions loaded from JSONL format
**So that** sessions are restored correctly

**Acceptance Criteria:**
- [ ] `Load()` reads `.jsonl` file only (no `.json` fallback)
- [ ] Parses line 1 as metadata
- [ ] Parses lines 2+ as messages
- [ ] Returns `Session` struct
- [ ] Clear error if `.jsonl` not found

**Files:** `session/session.go`

---

#### Story 3.3: SyncFromStore update
**As a** developer, **I want** SyncFromStore to work with JSONL
**So that** incremental saves work correctly

**Acceptance Criteria:**
- [ ] `SyncFromStore()` copies store messages to session
- [ ] `Save()` writes updated session as JSONL
- [ ] Token usage accumulated correctly

**Files:** `session/session.go`

---

#### Story 3.4: List sessions
**As a** user, **I want** `sprout ls` to find JSONL sessions
**So that** I can list and resume sessions

**Acceptance Criteria:**
- [ ] `List()` reads `.sessions/*.jsonl` files
- [ ] Returns sessions sorted by updatedAt desc
- [ ] Shows ID, age, message count, model

**Files:** `session/session.go`

---

#### Story 3.5: Session tests
**As a** developer, **I want** tests for JSONL session storage
**So that** I can verify correctness

**Acceptance Criteria:**
- [ ] Test JSONL save/load roundtrip
- [ ] Test message integrity
- [ ] Test token usage accumulation
- [ ] Test list sessions

**Files:** `session/session_test.go`

---

### Task Dependency Graph (Phase 3)

```
3.1 JSONL save ── 3.2 JSONL load ── 3.3 SyncFromStore ── 3.4 List ── 3.5 Tests
```

**Sequential within phase** — each step builds on the previous.

---

## Cross-Phase Dependencies

```
Phase 1 (Thinking)
    │
    ├── Phase 2 (Tools) — starts after Phase 1
    │       │
    │       └── Phase 3 (JSONL) — can overlap with Phase 2
    │
    └── Phase 3 (JSONL) — can start after Phase 1, parallel with Phase 2
```

**Key insight:** Phase 3 (JSONL) only depends on Phase 1 (for the `Thinking` field in `Message`). It can run in parallel with Phase 2.

---

## Branch Strategy

| Branch | Base | Contains |
|--------|------|----------|
| `ai/thinking` | `main` | ADR-003 + Stories 1.1-1.6 |
| `ai/tools` | `ai/thinking` | ADR-004 + Stories 2.1-2.12 |
| `ai/jsonl-sessions` | `ai/thinking` | ADR-005 + Stories 3.1-3.5 |

**Note:** `ai/tools` and `ai/jsonl-sessions` both branch from `ai/thinking`. They can be developed in parallel. When both are ready, merge `ai/jsonl-sessions` first (smaller), then `ai/tools` (larger).

---

## Effort Summary

| Phase | Parallel | Sequential |
|-------|----------|------------|
| Phase 1: Thinking | 2-3 hours | 4-5 hours |
| Phase 2: Tools | 4-5 hours | 8-10 hours |
| Phase 3: JSONL | 1-2 hours | 1-2 hours |
| **Total** | **7-10 hours** | **13-17 hours** |
