# Checklist: ADR-003 Thinking Display

## Provider
- [ ] Add `ThinkingDelta string` to `StreamEvent` (`provider/provider.go`)
- [ ] Add `Think bool` to `chatRequest` (`provider/ollama.go`)
- [ ] Parse `thinking` from `chatStreamChunk` (`provider/ollama.go`)
- [ ] Emit `ThinkingDelta` events (`provider/ollama.go`)

## Message Store
- [ ] Add `Thinking string` to `Message` struct (`message/store.go`)

## Config
- [ ] Add `Think bool` to `Config` (`config/config.go`)
- [ ] Default value: `true`
- [ ] Add `--think` CLI flag (`main.go`)

## TUI
- [ ] Add `thinkingStyle` to `tui/styles.go`
- [ ] Add `SetThinkingContent()` to `ChatModel` (`tui/chat.go`)
- [ ] Accumulate thinking in `tui/model.go`
- [ ] Render thinking block in `tui/chat.go`
- [ ] Add `t` key toggle (`tui/model.go`)
- [ ] Show thinking state in footer (`tui/footer.go`)

## Testing
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] Thinking streams in real-time
- [ ] Thinking persisted in session JSON
- [ ] Toggle works
