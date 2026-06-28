# Stories: ADR-004 Tool Calling

## Problem

Sprout is a chat-only interface. The model can discuss code but can't execute commands, read files, or perform actions. This limits its usefulness as a developer tool.

## Goals

1. **Agentic** — Model can take actions, not just chat
2. **Safe** — User approves every tool call
3. **Multi-step** — Model can chain tool calls for complex workflows
4. **Transparent** — Tool calls and results visible in chat

---

## Epic 1: Tool Interface

### Story 1.1: Tool interface and types
**As a** developer, **I want** a Tool interface and ToolCall types
**So that** tools have a standard contract

**Acceptance Criteria:**
- [ ] `Tool` interface: `Name()`, `Description()`, `Parameters()`, `Execute()`
- [ ] `ToolCall` struct: `Function ToolCallFunction`
- [ ] `ToolCallFunction` struct: `Name string`, `Arguments map[string]interface{}`
- [ ] `ToolCalls []ToolCall` on `StreamEvent`
- [ ] `ToolCallRequest ToolCall` on `StreamEvent`
- [ ] `ToolCallResult string` on `StreamEvent`
- [ ] `ToolCallError string` on `StreamEvent`

**Files:** `tools/tool.go`, `provider/provider.go`

---

### Story 1.2: Message store tool fields
**As a** developer, **I want** messages to store tool calls and results
**So that** tool history persists across sessions

**Acceptance Criteria:**
- [ ] `RoleTool Role = "tool"` constant
- [ ] `ToolCalls []ToolCall` field on `Message`
- [ ] `ToolName string` field on `Message`
- [ ] JSON tags for serialization

**Files:** `message/store.go`

---

## Epic 2: Tool Registry

### Story 2.1: Registry
**As a** developer, **I want** a tool registry
**So that** tools are discovered and converted to Ollama format

**Acceptance Criteria:**
- [ ] `Registry` struct with `Register()`, `Get()`, `List()`
- [ ] `ToOllamaTools()` converts to Ollama API format
- [ ] `NewRegistry()` registers all 9 default tools

**Files:** `tools/registry.go`

---

## Epic 3: Individual Tools

### Story 3.1: Shell tool
**As a** user, **I want** to execute shell commands
**So that** I can run git, npm, build commands

**Acceptance Criteria:**
- [ ] `execute_shell` tool
- [ ] Parameters: `command` (required), `workdir` (optional)
- [ ] Captures stdout + stderr
- [ ] 30-second timeout
- [ ] Returns exit code in output

**Files:** `tools/shell.go`

---

### Story 3.2: File tools
**As a** user, **I want** to read, write, and edit files
**So that** I can work with code files

**Acceptance Criteria:**
- [ ] `read_file`: reads file, returns content
- [ ] `write_file`: writes content, creates parent dirs
- [ ] `edit_file`: find/replace in file
- [ ] Error handling: file not found, permission denied

**Files:** `tools/file.go`

---

### Story 3.3: Filesystem tools
**As a** user, **I want** to list, find, and search files
**So that** I can explore the codebase

**Acceptance Criteria:**
- [ ] `list_directory`: lists contents
- [ ] `find_files`: glob pattern search
- [ ] `search_in_files`: regex grep
- [ ] Respects `.gitignore`

**Files:** `tools/fs.go`

---

### Story 3.4: Web tools
**As a** user, **I want** to fetch URLs and search the web
**So that** I can look up documentation

**Acceptance Criteria:**
- [ ] `fetch_url`: fetches URL, returns content
- [ ] `web_search`: DuckDuckGo instant answer API
- [ ] 10-second timeout
- [ ] Max response size: 1MB

**Files:** `tools/web.go`

---

## Epic 4: Provider Tools

### Story 4.1: Provider tool support
**As a** developer, **I want** the provider to send tools to Ollama
**So that** the model can request tool execution

**Acceptance Criteria:**
- [ ] `Tools []map[string]interface{}` on `chatRequest`
- [ ] `tool_calls` parsed from `chatStreamChunk.message.tool_calls`
- [ ] `ToolCalls` emitted in `StreamEvent`
- [ ] `ChatStreamWithTools()` method on provider

**Files:** `provider/ollama.go`

---

## Epic 5: Agentic Loop

### Story 5.1: Multi-turn loop
**As a** developer, **I want** a multi-turn agentic loop
**So that** the model can chain tool calls

**Acceptance Criteria:**
- [ ] `AgenticLoop()` on `agent.Loop`
- [ ] Loop: model calls tool → result fed back → model continues
- [ ] Max turns configurable (default 5)
- [ ] Loop ends when model returns content (no tool_calls)

**Files:** `agent/loop.go`

---

### Story 5.2: User approval
**As a** user, **I want** to approve tool calls before execution
**So that** I control what runs on my system

**Acceptance Criteria:**
- [ ] Approval callback: `func(ToolCall) bool`
- [ ] User prompted for each tool call
- [ ] 'y' approves, 'n' rejects
- [ ] Rejected calls added as user message

**Files:** `agent/loop.go`, `tui/model.go`

---

## Epic 6: TUI Integration

### Story 6.1: Tool call display
**As a** user, **I want** to see tool calls and results in chat
**So that** I have full context

**Acceptance Criteria:**
- [ ] `AddToolCall()`: renders `🔧 tool_name({args})`
- [ ] `AddToolResult()`: renders tool output
- [ ] Tool calls shown during streaming
- [ ] Tool results shown after execution

**Files:** `tui/chat.go`

---

### Story 6.2: Approval prompt
**As a** user, **I want** an approval prompt for tool calls
**So that** I can approve/reject/edit

**Acceptance Criteria:**
- [ ] `stateToolApproval` state
- [ ] Input shows: `Execute tool_name? [y/n/e(dit)]`
- [ ] `y` → execute, `n` → reject, `e` → edit args
- [ ] Edit mode: input shows JSON args

**Files:** `tui/model.go`, `tui/input.go`

---

## Epic 7: Config

### Story 7.1: Tool config
**As a** user, **I want** to control tool settings
**So that** I can configure agentic behavior

**Acceptance Criteria:**
- [ ] `ToolsEnabled bool` in config (default `true`)
- [ ] `MaxTurns int` in config (default `5`)
- [ ] `--tools` / `--tools=false` CLI flag
- [ ] `--max-turns=N` CLI flag

**Files:** `config/config.go`, `main.go`

---

## Epic 8: Integration

### Story 8.1: Wire TUI to agentic loop
**As a** developer, **I want** the TUI to use the agentic loop
**So that** tools work end-to-end

**Acceptance Criteria:**
- [ ] TUI calls `AgenticLoop()` instead of `SendMessage()`
- [ ] Tool call events routed to display and approval
- [ ] Tool results fed back to loop
- [ ] Session saved after each tool exchange
- [ ] Esc cancels entire agentic loop

**Files:** `tui/model.go`

---

## Definition of Done

- [ ] 9 tools implemented and registered
- [ ] Model can call tools via Ollama native API
- [ ] Multi-turn loop works (up to 5 turns)
- [ ] User approves each tool call
- [ ] Tool calls and results shown in chat
- [ ] Session persists tool history
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
