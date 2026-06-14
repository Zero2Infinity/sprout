# MVP Checklist: Sprout

Quick reference for tracking implementation progress.

---

## Phase 1: Foundation
- [ ] 1.1 Load JSON config (`config/config.go`)
- [ ] 1.2 Create data directory (`config/config.go`, `message/store.go`)
- [ ] 2.1 Connect to Ollama (`provider/ollama.go`)
- [ ] Initialize `go.mod` with dependencies

## Phase 2: Core Loop
- [ ] 2.2 Stream response token-by-token (`provider/ollama.go`, `agent/loop.go`)
- [ ] 3.1 Create new session (`session/session.go`, `message/store.go`)
- [ ] 8.2 Handle Ollama unavailable (`provider/ollama.go`, `tui/model.go`)

## Phase 3: TUI Layout
- [ ] 4.1 Display header (`tui/header.go`)
- [ ] 4.2 Display chat messages (`tui/chat.go`)
- [ ] 4.3 Display footer (`tui/footer.go`)
- [ ] 4.4 Text input area (`tui/input.go`)

## Phase 4: Session Persistence
- [ ] 3.2 Load existing session (`session/session.go`, `message/store.go`)
- [ ] 3.3 Save session on exit (`session/session.go`, `message/store.go`)

## Phase 5: Enhancements
- [ ] 5.1 Syntax highlighting (`tui/chat.go`, `tui/styles.go`)
- [ ] 6.1 Navigate input history (`tui/input.go`, `session/session.go`)
- [ ] 7.1 Copy message to clipboard (`tui/chat.go`, `tui/toast.go`)
- [ ] 8.1 Cancel stream with Esc (`agent/loop.go`, `tui/model.go`)

---

## Testing Checklist
- [ ] Ollama not running → clear error message
- [ ] New session → created on first run
- [ ] Existing session → loaded on restart
- [ ] Streaming → tokens appear in real-time
- [ ] Ctrl+C → session saved
- [ ] Esc during stream → partial response kept
- [ ] Up/Down → history navigation works
- [ ] Ctrl+Shift+C → message copied to clipboard
- [ ] Resize → layout adjusts correctly
