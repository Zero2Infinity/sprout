# Checklist: ADR-005 JSONL Session Storage

## Save
- [ ] `Save()` writes `.jsonl` format (`session/session.go`)
- [ ] Line 1: metadata (id, model, timestamps, tokenUsage)
- [ ] Lines 2+: messages (one JSON object per line)

## Load
- [ ] `Load()` reads `.jsonl` file only (`session/session.go`)
- [ ] No fallback to `.json` (clean break)
- [ ] Parse line 1 as metadata
- [ ] Parse lines 2+ as messages
- [ ] Clear error message if session not found

## List
- [ ] `List()` finds `.jsonl` files (`session/session.go`)
- [ ] Returns sessions sorted by updatedAt desc

## Integration
- [ ] `SyncFromStore()` works with JSONL (`session/session.go`)
- [ ] File extension updated everywhere

## Testing
- [ ] JSONL save/load roundtrip
- [ ] Message integrity
- [ ] Token usage accumulation
- [ ] List sessions
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
