# Stories: ADR-005 JSONL Session Storage

## Problem

Session files use JSON arrays. Every save rewrites the entire file. Large sessions create noisy diffs. Partial writes can corrupt files.

## Goals

1. **Efficient** — Append-only writes (future optimization)
2. **Clean diffs** — Each message is a separate line
3. **Simple** — No migration logic, clean break from old format

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
- [ ] `Load()` reads `.jsonl` file only
- [ ] No fallback to `.json` (clean break)
- [ ] Parses line 1 as metadata
- [ ] Parses lines 2+ as messages
- [ ] Returns `Session` struct with all fields
- [ ] Clear error if `.jsonl` not found

**Files:** `session/session.go`

---

## Epic 3: Integration

### Story 3.1: SyncFromStore update
**As a** developer, **I want** SyncFromStore to work with JSONL
**So that** incremental saves work

**Acceptance Criteria:**
- [ ] `SyncFromStore()` copies store messages to session
- [ ] `Save()` writes updated session as JSONL
- [ ] Token usage accumulated correctly

**Files:** `session/session.go`

---

### Story 3.2: List sessions
**As a** user, **I want** `sprout ls` to find JSONL sessions
**So that** I can list and resume sessions

**Acceptance Criteria:**
- [ ] `List()` reads `.sessions/*.jsonl` files
- [ ] Returns sessions sorted by updatedAt desc
- [ ] Shows ID, age, message count, model

**Files:** `session/session.go`

---

## Epic 4: Testing

### Story 4.1: Session tests
**As a** developer, **I want** tests for JSONL storage
**So that** I can verify correctness

**Acceptance Criteria:**
- [ ] Test JSONL save/load roundtrip
- [ ] Test message integrity (all fields preserved)
- [ ] Test token usage accumulation
- [ ] Test concurrent access (thread safety)
- [ ] Test list sessions

**Files:** `session/session_test.go`

---

## Definition of Done

- [ ] Sessions saved as `.jsonl`
- [ ] Sessions loaded from `.jsonl`
- [ ] Old `.json` sessions not supported (clean break)
- [ ] All existing tests pass
- [ ] New tests for JSONL roundtrip
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
