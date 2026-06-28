# Checklist: ADR-004 Tool Calling

## Tool Interface
- [ ] Define `Tool` interface (`tools/tool.go`)
- [ ] Define `ToolCall`, `ToolCallFunction` structs (`tools/tool.go`)
- [ ] Add `RoleTool` constant (`message/store.go`)
- [ ] Add `ToolCalls`, `ToolName` to `Message` (`message/store.go`)
- [ ] Add `ToolCalls`, `ToolCallRequest`, `ToolCallResult`, `ToolCallError` to `StreamEvent` (`provider/provider.go`)

## Tool Registry
- [ ] Create `Registry` struct (`tools/registry.go`)
- [ ] Implement `Register()`, `Get()`, `List()`
- [ ] Implement `ToOllamaTools()`

## Individual Tools
- [ ] `execute_shell` (`tools/shell.go`)
- [ ] `read_file` (`tools/file.go`)
- [ ] `write_file` (`tools/file.go`)
- [ ] `edit_file` (`tools/file.go`)
- [ ] `list_directory` (`tools/fs.go`)
- [ ] `find_files` (`tools/fs.go`)
- [ ] `search_in_files` (`tools/fs.go`)
- [ ] `fetch_url` (`tools/web.go`)
- [ ] `web_search` (`tools/web.go`)

## Provider
- [ ] Add `Tools` to `chatRequest` (`provider/ollama.go`)
- [ ] Parse `tool_calls` from chunks (`provider/ollama.go`)
- [ ] Add `ChatStreamWithTools()` method (`provider/ollama.go`)

## Agentic Loop
- [ ] Implement `AgenticLoop()` (`agent/loop.go`)
- [ ] Multi-turn with max turns
- [ ] Approval callback
- [ ] Esc cancellation

## TUI
- [ ] Add `stateToolApproval` state (`tui/model.go`)
- [ ] Add `AddToolCall()` to chat (`tui/chat.go`)
- [ ] Add `AddToolResult()` to chat (`tui/chat.go`)
- [ ] Approval prompt in input (`tui/input.go`)
- [ ] Wire TUI to agentic loop (`tui/model.go`)

## Config
- [ ] Add `ToolsEnabled bool` (`config/config.go`)
- [ ] Add `MaxTurns int` (`config/config.go`)
- [ ] Add `--tools` CLI flag (`main.go`)
- [ ] Add `--max-turns` CLI flag (`main.go`)

## Testing
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] Tool calls execute correctly
- [ ] Multi-turn loop works
- [ ] Approval prompt works
- [ ] Esc cancels loop
