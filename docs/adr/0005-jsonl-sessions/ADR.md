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
| **Migration** | Auto-detect old JSON format | Backward compatible; old sessions load fine |
| **Save behavior** | Append new lines, rewrite header | Efficient for incremental saves |

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

### 2. Migration from Legacy JSON

Auto-detect old format on load:

```go
func Load(dataDir, id string) (*Session, error) {
    path := filepath.Join(dataDir, id+".jsonl")
    data, err := os.ReadFile(path)
    if err != nil {
        // Try legacy .json extension
        path = filepath.Join(dataDir, id+".json")
        data, err = os.ReadFile(path)
        if err != nil {
            return nil, fmt.Errorf("session not found: %s", id)
        }
    }

    // Detect format
    if isLegacyJSON(data) {
        return migrateToJSONL(data, path)
    }
    return parseJSONL(data)
}

func isLegacyJSON(data []byte) bool {
    data = bytes.TrimSpace(data)
    if len(data) == 0 {
        return false
    }
    // Legacy format starts with '{' and contains "messages" key
    // JSONL starts with '{' but first line is metadata (has "id" key, no "messages" key)
    if data[0] != '{' {
        return false
    }
    // Try to parse as legacy
    var legacy struct {
        Messages json.RawMessage `json:"messages"`
    }
    if json.Unmarshal(data, &legacy) == nil && legacy.Messages != nil {
        return true
    }
    return false
}

func migrateToJSONL(data []byte, jsonlPath string) (*Session, error) {
    // Parse legacy JSON
    var legacy struct {
        ID         string       `json:"id"`
        Model      string       `json:"model"`
        CreatedAt  time.Time    `json:"createdAt"`
        UpdatedAt  time.Time    `json:"updatedAt"`
        Messages   []Message    `json:"messages"`
        TokenUsage TokenUsage   `json:"tokenUsage"`
    }
    if err := json.Unmarshal(data, &legacy); err != nil {
        return nil, fmt.Errorf("parsing legacy session: %w", err)
    }

    // Build JSONL content
    var buf bytes.Buffer

    // Line 1: metadata
    meta := map[string]interface{}{
        "id": legacy.ID, "model": legacy.Model,
        "createdAt": legacy.CreatedAt, "updatedAt": legacy.UpdatedAt,
        "tokenUsage": legacy.TokenUsage,
    }
    json.NewEncoder(&buf).Encode(meta)

    // Lines 2+: messages
    for _, msg := range legacy.Messages {
        json.NewEncoder(&buf).Encode(msg)
    }

    // Write to .jsonl file
    if err := os.WriteFile(jsonlPath, buf.Bytes(), 0644); err != nil {
        return nil, fmt.Errorf("writing migrated JSONL: %w", err)
    }

    // Optionally remove old .json file
    // os.Remove(oldPath)

    return &Session{
        ID: legacy.ID, Model: legacy.Model,
        CreatedAt: legacy.CreatedAt, UpdatedAt: legacy.UpdatedAt,
        Messages: legacy.Messages, TokenUsage: legacy.TokenUsage,
    }, nil
}
```

### 3. Save Behavior

**New session (first save):**
Write all lines: metadata line + all message lines.

**Existing session (incremental save):**
- Update the metadata line (first line) with new `updatedAt` and `tokenUsage`
- Append new message lines at the end

```go
func Save(dataDir string, sess *Session) error {
    path := filepath.Join(dataDir, sess.ID+".jsonl")

    var buf bytes.Buffer

    // Line 1: metadata (always rewritten)
    meta := map[string]interface{}{
        "id": sess.ID, "model": sess.Model,
        "createdAt": sess.CreatedAt, "updatedAt": time.Now(),
        "tokenUsage": sess.TokenUsage,
    }
    json.NewEncoder(&buf).Encode(meta)

    // Lines 2+: all messages (full rewrite for simplicity)
    // Future optimization: track new messages and append only
    for _, msg := range sess.Messages {
        json.NewEncoder(&buf).Encode(msg)
    }

    return os.WriteFile(path, buf.Bytes(), 0644)
}
```

**Note:** For MVP, we rewrite the full file on save. The JSONL format enables future optimization to append-only writes by tracking which messages have been persisted.

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
| `session/session.go` | Rewrite `Load()` — detect format, parse JSONL or migrate from JSON |
| `session/session.go` | Rewrite `Save()` — write JSONL format |
| `session/session.go` | Update `SyncFromStore()` — no change needed |
| `session/session.go` | Add `isLegacyJSON()`, `migrateToJSONL()`, `parseJSONL()` helpers |
| `session/session.go` | Update file extension from `.json` to `.jsonl` |

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
  → Detects format (JSONL or legacy JSON)
  → If legacy: migrate to .jsonl, return session
  → If JSONL: parse line by line
  → Line 1: metadata → Session struct
  → Lines 2+: messages → Session.Messages
```

## Consequences

- **Positive:** Append-only writes — no full file rewrite needed (future optimization)
- **Positive:** Better diffs — each message is a separate line in version control
- **Positive:** Streamable — can process large sessions line by line
- **Positive:** Backward compatible — auto-migrates old JSON sessions
- **Negative:** Two formats during migration period
- **Mitigation:** Auto-detection handles both; migration is one-way
- **Negative:** File extension changes from `.json` to `.jsonl`
- **Mitigation:** Auto-detect on load; old `.json` files still work

## Alternatives Considered

| Alternative | Rejected Because |
|-------------|------------------|
| Keep JSON format | Full rewrite on every save; poor diffs for large sessions |
| SQLite | Overkill for single-user local storage |
| Binary format | Not human-readable; breaks session inspection |
| Streaming JSON (SSE) | Overly complex; JSONL is simpler |
