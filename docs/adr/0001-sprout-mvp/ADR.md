# ADR-001: Sprout MVP

**Status:** Accepted
**Date:** 2026-06-13
**Deciders:** Rahul

## Context

Build a minimal working CLI chat application for local Ollama models, inspired by OpenCode but stripped to MVP essentials. This serves as a learning project and foundation for future extensions.

## Decision

### Core Architecture

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Existing codebase is Go, consistent skill set |
| LLM Provider | Ollama (OpenAI-compatible) | Local, no API keys, OpenAI SDK reuses |
| TUI Framework | Bubble Tea | Proven in OpenCode, handles resize/scroll/input cleanly |
| CLI Framework | Cobra | Industry standard Go CLI |
| Config | Viper + JSON | Flexible (env + file) |
| Session Storage | JSON files | Simple persistence, no DB overhead |

### Dependencies

```go
require (
    github.com/alecthomas/chroma/v2 v2.15.0      // Syntax highlighting
    github.com/atotto/clipboard v0.1.4           // Clipboard copy
    github.com/charmbracelet/bubbletea v1.3.5    // TUI framework
    github.com/charmbracelet/bubbles v0.21.0     // TUI components
    github.com/charmbracelet/lipgloss v1.1.0     // TUI styling
    github.com/google/uuid v1.6.0                // Unique IDs
    github.com/openai/openai-go v0.1.0-beta.2   // Ollama client
    github.com/spf13/cobra v1.9.1                // CLI framework
    github.com/spf13/viper v1.20.0               // Config
)
```

### UI Layout

```
+--------------------------------------------------+
| /Users/rahul/project | Session: abc123 | qwen3:27b |
+--------------------------------------------------+
|                                                   |
| User: Hello, explain Go context                   |
|                                                   |
| Assistant: Go context is a struct that carries... |
| ```go                                             |
| ctx := context.Background()                       |
| ```                                               |
|                                                   |
+--------------------------------------------------+
| Tokens: 150/2,450 | Enter send | Esc cancel       |
+--------------------------------------------------+
| [input textarea]                                  |
+--------------------------------------------------+
```

### Features

| Feature | Implementation |
|---------|----------------|
| Streaming | Token-by-token via Ollama OpenAI-compatible API |
| Syntax Highlighting | chroma for code blocks in chat messages |
| History Navigation | Up/Down arrows in input textarea |
| Clipboard Copy | Ctrl+Shift+C copies focused message |
| Session Persistence | JSON file per session in `.sessions/` (repo-local) |

### Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | Newline in input |
| `Up` / `Down` | History navigation |
| `Ctrl+C` | Quit (save session) |
| `Esc` | Cancel streaming |
| `Ctrl+Shift+C` | Copy last assistant message |

### Config Structure

```json
{
  "provider": {
    "baseURL": "http://localhost:11434/v1",
    "model": "qwen3:27b",
    "apiKey": "ollama"
  },
  "dataDir": "sessions"
}
```

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Full OpenCode fork | Too complex for MVP learning goal |
| Python/Rust | Go chosen for consistency and team familiarity |
| SQLite | Overkill for session storage at MVP stage |
| Multiple providers | Ollama-only keeps scope minimal |
| Interactive model picker | Config-hardcoded is simpler for MVP |

## Consequences

- **Positive:** Working chat app in ~800-1000 lines of Go
- **Positive:** Local-first, no API key management
- **Positive:** Foundation for adding tools, providers, LSP later
- **Negative:** No multi-provider support initially
- **Negative:** No session sharing/collaboration
- **Mitigation:** Design interfaces to allow future extension
