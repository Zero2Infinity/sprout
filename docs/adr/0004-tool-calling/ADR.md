# ADR-004: Agentic Tool Calling

**Status:** Proposed
**Date:** 2026-06-28
**Deciders:** Rahul

## Context

The project needs tool calling to enable agentic workflows. The model should be able to execute shell commands, read/write files, search the web, and perform filesystem operations. This transforms Sprout from a chat app into a developer productivity tool.

Ollama's native `/api/chat` endpoint supports a `tools` parameter that accepts JSON tool definitions. The model returns structured `tool_calls` in its response — no prompt-based parsing required.

## Decision

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Tool API** | Ollama native `tools` parameter | Structured tool calls; no prompt parsing |
| **Tools** | 9 built-in tools | Covers core developer workflows |
| **Loop** | Multi-turn, max 5 turns | Enables complex multi-step workflows |
| **Approval** | Always ask user | Safety first — user controls execution |
| **Display** | Inline in chat | Full context visible |
| **Output** | Full output shown | No truncation — complete information |

## Detailed Design

### 1. Ollama API

**Request with tools:**
```json
{
  "model": "qwen3.6:27b",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "read main.go and explain it"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "description": "Read the contents of a file at the given path",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "The file path to read"
            }
          },
          "required": ["path"]
        }
      }
    }
  ],
  "stream": true
}
```

**Response with tool call:**
```json
{
  "model": "qwen3.6:27b",
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [
      {
        "function": {
          "name": "read_file",
          "arguments": {"path": "main.go"}
        }
      }
    ]
  },
  "done": false
}
```

**Tool result sent back:**
```json
{
  "model": "qwen3.6:27b",
  "messages": [
    {"role": "user", "content": "read main.go and explain it"},
    {
      "role": "assistant",
      "content": "",
      "tool_calls": [{"function": {"name": "read_file", "arguments": {"path": "main.go"}}}]
    },
    {
      "role": "tool",
      "content": "<file contents here>",
      "tool_name": "read_file"
    }
  ],
  "tools": [...],
  "stream": true
}
```

**Final response after tool result:**
```json
{
  "model": "qwen3.6:27b",
  "message": {
    "role": "assistant",
    "content": "Here's what main.go does: It sets up a Cobra root command..."
  },
  "done": true
}
```

### 2. Tool Registry

| Tool | Description | Parameters |
|------|-------------|------------|
| `execute_shell` | Run shell command with output capture | `command: string`, `workdir?: string` |
| `read_file` | Read file contents | `path: string` |
| `write_file` | Write/create/overwrite file | `path: string`, `content: string` |
| `edit_file` | Edit file with find/replace | `path: string`, `old_string: string`, `new_string: string` |
| `list_directory` | List directory contents | `path?: string` (defaults to cwd) |
| `find_files` | Find files by glob pattern | `pattern: string`, `path?: string` |
| `search_in_files` | Search file contents (grep) | `pattern: string`, `path?: string`, `include?: string` |
| `fetch_url` | Fetch URL content | `url: string`, `format?: "text"\|"markdown"\|"html"` |
| `web_search` | Search the web | `query: string`, `num_results?: int` |

### 3. New Package: `tools/`

```
tools/
├── tool.go        # Tool interface + types
├── registry.go    # Registry + Ollama format conversion
├── shell.go       # execute_shell
├── file.go        # read_file, write_file, edit_file
├── fs.go          # list_directory, find_files, search_in_files
└── web.go         # fetch_url, web_search
```

**Tool interface:**
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}  // JSON Schema
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

**Registry:**
```go
type Registry struct {
    tools map[string]Tool
}

func NewRegistry() *Registry {
    r := &Registry{tools: make(map[string]Tool)}
    r.Register(&ShellTool{})
    r.Register(&ReadFileTool{})
    r.Register(&WriteFileTool{})
    r.Register(&EditFileTool{})
    r.Register(&ListDirTool{})
    r.Register(&FindFilesTool{})
    r.Register(&SearchInFilesTool{})
    r.Register(&FetchURLTool{})
    r.Register(&WebSearchTool{})
    return r
}

func (r *Registry) Get(name string) (Tool, bool) {
    t, ok := r.tools[name]
    return t, ok
}

func (r *Registry) List() []Tool { ... }

func (r *Registry) ToOllamaTools() []map[string]interface{} {
    // Convert each Tool to Ollama's tool definition format
    var tools []map[string]interface{}
    for _, t := range r.tools {
        tools = append(tools, map[string]interface{}{
            "type": "function",
            "function": map[string]interface{}{
                "name":        t.Name(),
                "description": t.Description(),
                "parameters":  t.Parameters(),
            },
        })
    }
    return tools
}
```

### 4. Agentic Loop

The agentic loop runs in `agent/loop.go` and orchestrates multi-turn tool calls:

```
1. Send message + tools to Ollama
2. Stream response
3. If tool_calls in response:
   a. Display tool call inline: "🔧 read_file({path: "main.go"})"
   b. Prompt user: "Execute? [y/n/e(dit args)]"
   c. If 'y': execute tool → add role:"tool" message → goto 1
   d. If 'n': add rejection as user message → goto 1
   e. If 'e': let user edit args → re-execute
4. If content (no tool_calls): final response → done
5. Max 5 turns — after that, force final response
```

**Pseudocode:**
```go
func (l *Loop) AgenticLoop(ctx context.Context, content string, registry *tools.Registry, maxTurns int, approveFn func(ToolCall) bool) (<-chan StreamEvent, error) {
    // Add user message to store
    l.store.Add(message.Message{Role: message.RoleUser, Content: content})

    events := make(chan StreamEvent)

    go func() {
        defer close(events)

        for turn := 0; turn < maxTurns; turn++ {
            // Send to Ollama with tools
            streamEvents, err := l.provider.ChatStreamWithTools(ctx, l.store, registry.ToOllamaTools())
            if err != nil {
                events <- StreamEvent{Err: err}
                return
            }

            // Collect response
            var content string
            var thinking string
            var toolCalls []ToolCall

            for ev := range streamEvents {
                if ev.ThinkingDelta != "" {
                    thinking += ev.ThinkingDelta
                    events <- ev  // forward thinking
                }
                if ev.ContentDelta != "" {
                    content += ev.ContentDelta
                    events <- ev  // forward content
                }
                if len(ev.ToolCalls) > 0 {
                    toolCalls = ev.ToolCalls
                }
                if ev.Err != nil {
                    events <- ev
                    return
                }
            }

            // No tool calls → final response, done
            if len(toolCalls) == 0 {
                l.store.Add(message.Message{
                    Role:      message.RoleAssistant,
                    Content:   content,
                    Thinking:  thinking,
                    Timestamp: time.Now(),
                })
                events <- StreamEvent{Complete: true}
                return
            }

            // Tool calls → ask approval
            for _, tc := range toolCalls {
                events <- StreamEvent{ToolCallRequest: tc}

                if !approveFn(tc) {
                    // User rejected — add rejection and continue
                    l.store.Add(message.Message{
                        Role:    message.RoleAssistant,
                        Content: content,
                        Thinking: thinking,
                        ToolCalls: toolCalls,
                        Timestamp: time.Now(),
                    })
                    l.store.Add(message.Message{
                        Role:    message.RoleUser,
                        Content: fmt.Sprintf("Tool call %s was rejected by user.", tc.Function.Name),
                        Timestamp: time.Now(),
                    })
                    continue
                }

                // Execute tool
                tool, ok := registry.Get(tc.Function.Name)
                if !ok {
                    events <- StreamEvent{ToolCallError: fmt.Sprintf("Unknown tool: %s", tc.Function.Name)}
                    continue
                }

                result, err := tool.Execute(ctx, tc.Function.Arguments)
                if err != nil {
                    events <- StreamEvent{ToolCallError: err.Error()}
                    result = fmt.Sprintf("Error: %v", err)
                }

                events <- StreamEvent{ToolCallResult: result}

                // Add messages to store
                l.store.Add(message.Message{
                    Role:      message.RoleAssistant,
                    Content:   content,
                    Thinking:  thinking,
                    ToolCalls: toolCalls,
                    Timestamp: time.Now(),
                })
                l.store.Add(message.Message{
                    Role:     message.RoleTool,
                    Content:  result,
                    ToolName: tc.Function.Name,
                    Timestamp: time.Now(),
                })
            }

            // Continue loop — next turn
        }

        // Max turns reached
        events <- StreamEvent{ContentDelta: "\n\n[Max tool turns reached]", Complete: true}
    }()

    return events, nil
}
```

### 5. TUI States

```
stateIdle → stateStreaming → stateToolApproval → stateIdle (loop)
                                    ↓
                              stateCancelled (Esc)
```

New state: `stateToolApproval` — model requested a tool, waiting for user approval.

**TUI flow:**
1. `stateIdle`: User types message, presses Enter → `stateStreaming`
2. `stateStreaming`: Tokens stream in. If tool_calls arrive → `stateToolApproval`
3. `stateToolApproval`: Show tool call inline. Input area shows: `Execute read_file? [y/n/e]`
4. User types `y` → execute → `stateStreaming` (next turn)
5. User types `n` → reject → `stateStreaming` (next turn)
6. User types `e` → edit args → input shows editable JSON → Enter confirms
7. At any point: Esc → `stateCancelled` → cancel entire loop

### 6. Tool Approval UX

When tool call arrives:
1. Tool call shown inline in chat (read-only):
   ```
   🔧 read_file({path: "main.go"})
   ```
2. Input area replaced with approval prompt:
   ```
   Execute read_file? [y/n/e(dit)]
   ```
3. User types:
   - `y` or `Y`: Execute the tool
   - `n` or `N`: Reject the tool call
   - `e` or `E`: Edit the arguments (input shows JSON, Enter confirms)
4. On `e`: input area shows editable JSON args:
   ```
   Edit args: {"path":"main.go"}
   ```
   User modifies, presses Enter to confirm.

### 7. Struct Changes

**`provider/provider.go`:**
```go
type ToolCallFunction struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}

type ToolCall struct {
    Function ToolCallFunction `json:"function"`
}

type StreamEvent struct {
    ContentDelta    string
    ThinkingDelta   string
    ToolCalls       []ToolCall   // NEW: tool calls from model
    ToolCallRequest ToolCall     // NEW: tool call needing approval
    ToolCallResult  string       // NEW: tool execution result
    ToolCallError   string       // NEW: tool execution error
    Complete        bool
    Usage           *Usage
    Err             error
}
```

**`message/store.go`:**
```go
type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
    RoleTool      Role = "tool"      // NEW
)

type Message struct {
    Role      Role
    Content   string
    Thinking  string
    ToolCalls []ToolCall  // NEW: tool calls in this message
    ToolName  string      // NEW: for role=tool messages
    Timestamp time.Time
}
```

**`provider/ollama.go`:**
```go
type chatRequest struct {
    Model    string                 `json:"model"`
    Messages []chatMessage          `json:"messages"`
    Stream   bool                   `json:"stream"`
    Think    bool                   `json:"think,omitempty"`
    Tools    []map[string]interface{} `json:"tools,omitempty"`  // NEW
}

type chatStreamChunk struct {
    // ... existing fields ...
    Message *chatMessage `json:"message"`  // now includes tool_calls
}

type chatMessage struct {
    Role     string       `json:"role"`
    Content  string       `json:"content"`
    Thinking string       `json:"thinking,omitempty"`
    ToolCalls []ToolCall  `json:"tool_calls,omitempty"`  // NEW
}
```

### 8. Config Changes

```go
type Config struct {
    // ... existing fields ...
    Think        bool `json:"think"`         // default true
    ToolsEnabled bool `json:"tools"`         // default true
    MaxTurns     int  `json:"maxTurns"`      // default 5
}
```

CLI flags:
- `--tools` / `--tools=false` — enable/disable tool calling
- `--max-turns=N` — set max agentic turns (default 5)

### 9. Files Changed

| File | Change |
|------|--------|
| `tools/tool.go` | NEW: Tool interface + types |
| `tools/registry.go` | NEW: Registry + Ollama format conversion |
| `tools/shell.go` | NEW: Shell execution tool |
| `tools/file.go` | NEW: File read/write/edit tools |
| `tools/fs.go` | NEW: Filesystem list/find/search tools |
| `tools/web.go` | NEW: URL fetch + web search tools |
| `provider/ollama.go` | Add `Tools` to request, parse `tool_calls` from chunks |
| `provider/provider.go` | Add `ToolCall` struct, `ToolCalls` to `StreamEvent` |
| `agent/loop.go` | Add `AgenticLoop()` — multi-turn loop with approval callback |
| `message/store.go` | Add `RoleTool`, `ToolCalls`, `ToolName` to Message |
| `tui/model.go` | Handle tool calls: display, approval, results, state machine |
| `tui/chat.go` | Add `AddToolCall()`, `AddToolResult()` for rendering |
| `tui/input.go` | Approval prompt mode, edit args mode |
| `config/config.go` | Add `ToolsEnabled`, `MaxTurns` |
| `main.go` | Add `--tools`, `--max-turns` flags, build registry |

### 10. Session JSON

```json
{
  "messages": [
    {
      "role": "user",
      "content": "read main.go and explain it",
      "thinking": "",
      "timestamp": "2026-06-28T10:00:00Z"
    },
    {
      "role": "assistant",
      "thinking": "I need to read the file first...",
      "content": "",
      "tool_calls": [{"name": "read_file", "arguments": {"path": "main.go"}}],
      "timestamp": "2026-06-28T10:00:01Z"
    },
    {
      "role": "tool",
      "content": "<file contents here>",
      "tool_name": "read_file",
      "thinking": "",
      "tool_calls": [],
      "timestamp": "2026-06-28T10:00:02Z"
    },
    {
      "role": "assistant",
      "thinking": "Now I can explain the code...",
      "content": "Here's what main.go does...",
      "tool_calls": [],
      "timestamp": "2026-06-28T10:00:03Z"
    }
  ]
}
```

## Consequences

- **Positive:** Agentic workflows — model can take actions, not just chat
- **Positive:** Ollama-native tool calling — no prompt parsing, structured JSON
- **Positive:** Safety via always-ask approval — user controls execution
- **Positive:** Multi-turn loop enables complex multi-step workflows
- **Negative:** Always-ask adds friction for simple operations
- **Mitigation:** Could add auto-approve mode later (config flag)
- **Negative:** 9 tools × JSON Schema adds complexity
- **Mitigation:** Each tool is simple, isolated in its own file
- **Negative:** Max turns limit may prevent very complex workflows
- **Mitigation:** Configurable limit; user can increase if needed

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Prompt-based tool calling | Fragile parsing; Ollama supports native tools |
| Auto-approve all tools | Safety risk — shell execution needs user consent |
| Single tool per turn | Limits complex workflows |
| Separate tool panel | Inline is simpler; no new TUI components needed |
| Separate tool history | Mixing with messages keeps context unified |
