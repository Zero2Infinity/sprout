# ADR-005: JSONL Session Storage

**Status:** Proposed
**Date:** 2026-06-28
**Deciders:** Rahul

## Context

Session files currently store the entire conversation as a single JSON array. This has drawbacks:

1. **Rewrite on every save** — the entire file is rewritten after each exchange, even though only new messages are appended
2. **Large sessions** — as conversations grow, the JSON file becomes large and slow to parse
3. **Version control** — JSON arrays create noisy diffs since the entire structure changes
4. **Atomicity** — partial writes can corrupt the file if the process crashes mid-write

JSONL (JSON Lines) solves these problems by storing each record as a separate line. The first line is session metadata, subsequent lines are messages.

## Decision

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Format** | JSONL (JSON Lines) | Append-only writes, better diffs, streamable |
| **Structure** | Header line + message lines | Standard JSONL pattern; metadata on first line |
| **Migration** | Clean break — no migration | Old `.json` sessions won't load; users create new sessions |
| **Save behavior** | Rewrite full file on save | Simple; JSONL enables future append optimization |

## Detailed Design

### 1. JSONL Format

Each session file is a `.jsonl` file with one JSON object per line:

**Line 1 — Session metadata:**
```json
{"id":"abc-123-def","model":"qwen3.6:27b","createdAt":"2026-06-28T10:00:00Z","updatedAt":"2026-06-28T10:05:00Z","tokenUsage":{"promptTokens":150,"completionTokens":2450,"totalTokens":2600,"totalDurationNs":2100000000,"evalDurationNs":1995000000}}
```

**Lines 2+ — Messages:**
```json
{"role":"user","content":"Explain Go concurrency","thinking":"","timestamp":"2026-06-28T10:00:00Z"}
{"role":"assistant","thinking":"The user wants to understand Go concurrency...","content":"Go concurrency is built on two core primitives...","tool_calls":[],"timestamp":"2026-06-28T10:00:01Z"}
{"role":"tool","content":"<file contents>","tool_name":"read_file","thinking":"","tool_calls":[],"timestamp":"2026-06-28T10:00:02Z"}
{"role":"assistant","thinking":"","content":"Here's what main.go does...","tool_calls":[],"timestamp":"2026-06-28T10:00:03Z"}
```

**File extension:** `.jsonl` (not `.json`)

**File path:** `.sessions/<uuid>.jsonl`

### 2. Load

Load only `.jsonl` files. No fallback to `.json`:

```go
func Load(dataDir, id string) (*Session, error) {
    path := filepath.Join(dataDir, id+".jsonl")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("session not found: %s (old JSON sessions are no longer supported)", id)
    }
    return parseJSONL(data)
}

func parseJSONL(data []byte) (*Session, error) {
    lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
    if len(lines) == 0 {
        return nil, fmt.Errorf("empty session file")
    }

    // Line 1: metadata
    var sess Session
    if err := json.Unmarshal(lines[0], &sess); err != nil {
        return nil, fmt.Errorf("parsing session metadata: %w", err)
    }

    // Lines 2+: messages
    for i := 1; i < len(lines); i++ {
        var msg Message
        if err := json.Unmarshal(lines[i], &msg); err != nil {
            continue // skip malformed messages
        }
        sess.Messages = append(sess.Messages, msg)
    }

    return &sess, nil
}
```

### 3. Save

Rewrite the full file on save. The JSONL format enables future append-only optimization:

```go
func Save(dataDir string, sess *Session) error {
    path := filepath.Join(dataDir, sess.ID+".jsonl")

    var buf bytes.Buffer

    // Line 1: metadata
    meta := map[string]interface{}{
        "id": sess.ID, "model": sess.Model,
        "createdAt": sess.CreatedAt, "updatedAt": time.Now(),
        "tokenUsage": sess.TokenUsage,
    }
    json.NewEncoder(&buf).Encode(meta)

    // Lines 2+: all messages
    for _, msg := range sess.Messages {
        json.NewEncoder(&buf).Encode(msg)
    }

    return os.WriteFile(path, buf.Bytes(), 0644)
}
```

### 4. SyncFromStore

```go
func SyncFromStore(sess *Session, store *message.Store) {
    sess.Messages = store.All()
    sess.UpdatedAt = time.Now()
}
```

After sync, the next `Save()` call writes the updated session.

### 5. Session Struct

```go
type Session struct {
    ID         string     `json:"id"`
    Model      string     `json:"model"`
    CreatedAt  time.Time  `json:"createdAt"`
    UpdatedAt  time.Time  `json:"updatedAt"`
    Messages   []Message  `json:"messages"`
    TokenUsage TokenUsage `json:"tokenUsage"`
}
```

No structural changes — the same `Session` struct is used. The change is only in how it's serialized/deserialized.

### 6. Files Changed

| File | Change |
|------|--------|
| `session/session.go` | Rewrite `Load()` — read `.jsonl` only, no JSON fallback |
| `session/session.go` | Rewrite `Save()` — write JSONL format |
| `session/session.go` | Add `parseJSONL()` helper |
| `session/session.go` | Update file extension from `.json` to `.jsonl` |
| `session/session.go` | Update `List()` to find `.jsonl` files |

### 7. Data Flow

```
Session created
  → Save() writes .jsonl file
  → Line 1: metadata
  → Lines 2+: empty (no messages yet)

User sends message
  → Message added to store
  → SyncFromStore() copies to session
  → Save() writes .jsonl file
  → Line 1: metadata (updated)
  → Lines 2+: all messages

Session loaded
  → Load() reads .jsonl file
  → parseJSONL() reads line by line
  → Line 1: metadata → Session struct
  → Lines 2+: messages → Session.Messages

List sessions
  → List() reads .sessions/*.jsonl files
  → Returns sessions sorted by updatedAt
```

## Consequences

- **Positive:** Append-only writes possible (future optimization)
- **Positive:** Better diffs — each message is a separate line
- **Positive:** Streamable — can process large sessions line by line
- **Positive:** Simple code — no migration logic, no format detection
- **Negative:** Old `.json` sessions won't load
- **Mitigation:** Users create new sessions; conversation history is in the message store
- **Negative:** File extension changes from `.json` to `.jsonl`
- **Mitigation:** Clear error message if old session is requested

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Auto-migrate JSON to JSONL | Adds complexity; old sessions are low-value (MVP stage) |
| Keep JSON format | Full rewrite on every save; poor diffs for large sessions |
| SQLite | Overkill for single-user local storage |
| Binary format | Not human-readable; breaks session inspection |
