# Bavarix — Phase 0 Progress

## Task Dependencies

```
Task 1 (scaffolding) ─── blocks all ───┐
                                        ├── Task 2: DS2 framing
                                        ├── Task 3: KWP2000 framing
                                        ├── Task 4: PRG decryptor [local PRG files needed]
                                        │     └── Task 5: PRG job extractor
                                        │           └── Task 10: Batch PRG extraction
                                        ├── Task 6: SP-Daten parser
                                        ├── Task 7: Translation parser
                                        ├── Task 8: Voltage gate
                                        ├── Task 9: Backup vault
                                        └── Task 11: CI pipeline
```

## Status

### Wave 0 — Blocker
- [x] Task 1: Go project scaffolding

### Wave 1 — Parallel (no interdependencies)
- [ ] Task 2: DS2 protocol framing — `pkg/protocol/ds2/`
- [ ] Task 3: KWP2000 protocol framing — `pkg/protocol/kwp2000/`
- [ ] Task 6: SP-Daten parser — `pkg/parser/spdaten/`
- [ ] Task 7: Translation parser — `pkg/parser/translations/`
- [ ] Task 8: Voltage gate — `pkg/safety/`
- [ ] Task 9: Backup vault — `pkg/safety/vault/`
- [ ] Task 11: CI pipeline — `.github/workflows/`

### Wave 2 — Depends on Wave 0 (local PRG files required)
- [ ] Task 4: PRG file decryptor — `pkg/parser/prg/`
- [ ] Task 5: PRG job extractor — `pkg/parser/prg/`
- [ ] Task 10: Batch PRG extraction — `cmd/extract-prg/`

## Agent Assignment Log

| Agent | Task | Branch | Status | Started | Completed |
|-------|------|--------|--------|---------|-----------|
| — | — | — | — | — | — |

## Merge Log

| Branch | Merged To | Commit | Conflicts |
|--------|-----------|--------|-----------|
| — | — | — | — |

## Notes

- PRG files (807MB) are local-only at `data/ediabas/ecu/` — Tasks 4, 5, 10 must run on this machine
- SP-Daten, SGDAT, translations are in repo via Git LFS
- All tasks are TDD: test first, implement second, commit after green
