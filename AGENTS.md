# AGENTS.md

## Status

Greenfield Go project — no code yet, only planning docs in `docs/adr/0001-sprout-mvp/`. All implementation starts from scratch.

## Project

CLI chat app for local Ollama models. MVP scope — single provider, no multi-user, no DB.

## Architecture (from docs/adr/0001-sprout-mvp/ADR.md)

| Component | Choice |
|-----------|--------|
| Language | Go |
| LLM Provider | Ollama (OpenAI-compatible API) |
| TUI Framework | Bubble Tea |
| CLI Framework | Cobra |
| Config | Viper + JSON |
| Session Storage | JSON files in `.sessions/` (repo-local) |

## Planned package layout (from checklist)

- `config/` — Viper config loading, data dir setup
- `provider/` — Ollama OpenAI-compatible client, streaming
- `agent/` — Main chat loop, message dispatch
- `session/` — Session create/load/save, UUID IDs
- `message/` — Message store, JSON persistence
- `tui/` — Bubble Tea components: header, chat, footer, input, toast, styles

## Implementation order (from docs/adr/0001-sprout-mvp/stories.md)

1. Foundation: config, data dirs, Ollama connection
2. Core Loop: streaming, session creation, error handling
3. TUI Layout: header, chat, footer, input
4. Session Persistence: load existing, save on exit
5. Enhancements: syntax highlighting, history nav, clipboard, Esc cancel

## Key details

- Config path: `.config/config.json` (repo-local, hidden)
- Default model: `qwen3.6:27b`
- Default endpoint: `http://localhost:11434/v1`
- Env var overrides: `OLLAMA_BASE_URL`, `OLLAMA_MODEL`
- Sessions stored: `.sessions/` (JSON per session, repo-local, hidden)
- Syntax highlighting: chroma library, dark theme (Tokyo Night / Catppuccin)
- Clipboard: `atotto/clipboard` (cross-platform)
- Estimated effort: 8-13 hours total

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
