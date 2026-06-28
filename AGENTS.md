# AGENTS.md

## Status

MVP implemented — Go CLI chat app for local Ollama models. Active development; most MVP features complete. See `CHECKLIST.md` for remaining polish items.

## Project

CLI chat app for local Ollama models. MVP scope — single provider, no multi-user, no DB.

## Architecture (from docs/adr/0001-sprout-mvp/ADR.md and docs/adr/0002-provider-metrics/ADR.md)

| Component | Choice |
|-----------|--------|
| Language | Go |
| LLM Provider | Ollama native `/api/chat` HTTP NDJSON streaming |
| TUI Framework | Bubble Tea |
| CLI Framework | Cobra |
| Config | Raw `encoding/json` + env var overrides |
| Session Storage | JSON files in `.sessions/` (repo-local) |
| Metrics/Timing | Token timing (duration, speed) via ADR-002 |

## Upcoming features (proposed ADRs)

| Feature | ADR | Branch | Status |
|---------|-----|--------|--------|
| Model thinking display | ADR-003 | `ai/thinking` | Proposed |
| Agentic tool calling | ADR-004 | `ai/tools` | Proposed |
| JSONL session storage | ADR-005 | `ai/jsonl-sessions` | Proposed |

**Implementation plan:** See `docs/adr/IMPLEMENTATION_PLAN.md` for full task breakdown with dependencies.

## Package layout

- `config/` — Raw JSON loading, env overrides (`OLLAMA_BASE_URL`, `OLLAMA_MODEL`), data dir setup
- `provider/` — Native Ollama `/api/chat` HTTP NDJSON client, streaming, token timing metrics
- `agent/` — Main chat loop, message dispatch
- `session/` — Session CRUD (create/load/save/list), token timing accumulation, `SyncFromStore()`
- `message/` — Thread-safe (`sync.RWMutex`) in-memory message store
- `tui/` — Bubble Tea components: model (root), header, chat (viewport), footer, input (textarea), toast, styles (chroma), cancel
- `tools/` — **(planned)** Tool interface, registry, and 9 built-in tools (shell, file, fs, web)

## Completed milestones (from docs/adr/0001-sprout-mvp/stories.md)

1. Foundation: config, data dirs, Ollama connection
2. Core Loop: streaming, session creation, error handling
3. TUI Layout: header, chat (scrollable viewport), footer, input (textarea)
4. Session Persistence: load existing, save on exit, sync after every exchange
5. Enhancements: Esc cancel (keeps partial), history nav (Up/Down, in-memory), model context window display
6. Metrics & Provider Switch: token timing (duration/speed), native `/api/chat` endpoint, ADR-002

## Key details

- Config path: `.config/config.json` (repo-local, hidden)
- Default model: `qwen3.6:27b`
- Default endpoint: `http://localhost:11434/v1`
- Env var overrides: `OLLAMA_BASE_URL`, `OLLAMA_MODEL`
- Sessions stored: `.sessions/` (JSON per session, repo-local, hidden)
- Syntax highlighting: chroma library, monokai theme — `HighlightCode()` exists but **not yet wired** into chat rendering
- Clipboard: **not implemented** (`atotto/clipboard` is indirect dep only)
- `sprout ls` — lists all sessions with ID, age, message count, model
- Exit summary — prints session ID, token counts, speed/latency, resume command
- Prompt history (Up/Down) — in-memory only, not persisted to session JSON
- Estimated effort: ~8-13 hours invested, MVP complete

## Commit conventions

Agent and human commits must be distinguishable in git history.

**Branch-based separation:**
- Agent works on `ai/*` branches (e.g., `ai/config-loading`, `ai/tui-header`)
- Human reviews and squash-merges via PR to `main`
- PR author = human, commit content = agent

**Commit message prefix (during development):**
- All agent commits use `[AI]` prefix: `[AI] Add config loading`
- Human commits use no prefix or `[human]`

**When merging to main:**
- Squash-merge the `ai/*` branch
- Clean up the commit message if needed
- The PR itself documents human review
