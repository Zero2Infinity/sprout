# Stories: ADR-005 JSONL Session Storage

## Problem

Session files use JSON arrays. Every save rewrites the entire file. Large sessions create noisy diffs. Partial writes can corrupt files.

## Goals

1. **Efficient** — Append-only writes (future optimization)
2. **Clean diffs** — Each message is a separate line
3. **Backward compatible** — Old JSON sessions auto-migrate

---

## Epic 1: JSONL Save

### Story 1.1: JSONL save
**As a** developer, **I want** sessions saved in JSONL format
**So that** saves are efficient and diffs are clean

**Acceptance Criteria:**
- [ ] `Save()` writes `.jsonl` file
- [ ] Line 1: metadata (id, model, timestamps, tokenUsage)
- [ ] Lines 2+: messages (one JSON object per line)
- [ ] File extension: `.jsonl`
- [ ] File path: `.sessions/<uuid>.jsonl`

**Files:** `session/session.go`

---

## Epic 2: JSONL Load

### Story 2.1: JSONL load
**As a** developer, **I want** sessions loaded from JSONL
**So that** sessions are restored correctly

**Acceptance Criteria:**
- [ ] `Load()` reads `.jsonl` file
- [ ] Parses line 1 as metadata
- [ ] Parses lines 2+ as messages
- [ ] Returns `Session` struct with all fields

**Files:** `session/session.go`

---

## Epic 3: Migration

### Story 3.1: Legacy JSON detection
**As a** developer, **I want** to detect old JSON format
**So that** migration works correctly

**Acceptance Criteria:**
- [ ] `isLegacyJSON()` checks for `"messages"` key
- [ ] `Load()` tries `.jsonl` first, then `.json`
- [ ] Returns clear error if neither exists

**Files:** `session/session.go`

---

### Story 3.2: JSON to JSONL migration
**As a** user, **I want** old sessions to migrate automatically
**So that** I don't lose history

**Acceptance Criteria:**
- [ ] `migrateToJSONL()` parses legacy JSON
- [ ] Writes new `.jsonl` file
- [ ] Preserves all messages and metadata
- [ ] Old `.json` file not deleted (preserved as backup)

**Files:** `session/session.go`

---

## Epic 4: Integration

### Story 4.1: SyncFromStore update
**As a** developer, **I want** SyncFromStore to work with JSONL
**So that** incremental saves work

**Acceptance Criteria:**
- [ ] `SyncFromStore()` copies store messages to session
- [ ] `Save()` writes updated session as JSONL
- [ ] Token usage accumulated correctly

**Files:** `session/session.go`

---

## Epic 5: Testing

### Story 5.1: Session tests
**As a** developer, **I want** tests for JSONL storage
**So that** I can verify correctness

**Acceptance Criteria:**
- [ ] Test JSONL save/load roundtrip
- [ ] Test legacy JSON migration
- [ ] Test message integrity (all fields preserved)
- [ ] Test token usage accumulation
- [ ] Test concurrent access (thread safety)

**Files:** `session/session_test.go`

---

## Definition of Done

- [ ] Sessions saved as `.jsonl`
- [ ] Sessions loaded from `.jsonl`
- [ ] Old `.json` sessions auto-migrate
- [ ] All existing tests pass
- [ ] New tests for JSONL roundtrip
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
