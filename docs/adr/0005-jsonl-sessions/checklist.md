# Checklist: ADR-005 JSONL Session Storage

## Save
- [ ] `Save()` writes `.jsonl` format (`session/session.go`)
- [ ] Line 1: metadata (id, model, timestamps, tokenUsage)
- [ ] Lines 2+: messages (one JSON object per line)

## Load
- [ ] `Load()` reads `.jsonl` file (`session/session.go`)
- [ ] Parse line 1 as metadata
- [ ] Parse lines 2+ as messages

## Migration
- [ ] `isLegacyJSON()` detects old format (`session/session.go`)
- [ ] `migrateToJSONL()` converts old sessions (`session/session.go`)
- [ ] Old `.json` file preserved
- [ ] New `.jsonl` file created

## Integration
- [ ] `SyncFromStore()` works with JSONL (`session/session.go`)
- [ ] `List()` finds `.jsonl` files (`session/session.go`)
- [ ] File extension updated everywhere

## Testing
- [ ] JSONL save/load roundtrip
- [ ] Legacy JSON migration
- [ ] Message integrity
- [ ] Token usage accumulation
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
