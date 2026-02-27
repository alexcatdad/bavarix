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
- [x] Task 2: DS2 protocol framing — `pkg/protocol/ds2/` (10/10 tests pass)
- [x] Task 3: KWP2000 protocol framing — `pkg/protocol/kwp2000/` (7/7 tests pass, fixed checksum 0x9A→0x9C)
- [x] Task 6: SP-Daten parser — `pkg/parser/spdaten/` (4/4 tests pass, adapted to real binary format)
- [x] Task 7: Translation parser — `pkg/parser/translations/` (6/6 tests pass)
- [x] Task 8: Voltage gate — `pkg/safety/` (10/10 tests pass, fixed boundary at 12.0V)
- [x] Task 9: Backup vault — `pkg/safety/vault/` (6/6 tests pass, fixed SQLite timestamp ordering)
- [x] Task 11: CI pipeline — `.github/workflows/` (committed)

### Wave 2 — Depends on Wave 0 (local PRG files required)
- [x] Task 4: PRG file decryptor — `pkg/parser/prg/` (5/5 tests pass)
- [x] Task 5: PRG job extractor — `pkg/parser/prg/` (9/9 tests pass, reverse-engineered binary format)
- [ ] Task 10: Batch PRG extraction — `cmd/extract-prg/` (agent running)

## Agent Assignment Log

| Agent | Task | Branch | Status | Tests |
|-------|------|--------|--------|-------|
| Agent | Task | Branch | Status | Tests |
|-------|------|--------|--------|-------|
| adf544 | Task 2: DS2 | worktree-agent-adf5445a | Merged | 10/10 pass |
| ab18f4 | Task 3: KWP2000 | worktree-agent-ab18f4d5 | Merged | 7/7 pass |
| a25692 | Task 7: Translations | worktree-agent-a25692e7 | Merged | 6/6 pass |
| a8d91e | Task 6: SP-Daten | worktree-agent-a8d91e5e | Merged | 4/4 pass |
| ad7bdc | Task 8: Voltage | worktree-agent-ad7bdc4e | Merged | 10/10 pass |
| ab80d5 | Task 9: Vault | worktree-agent-ab80d563 | Merged | 6/6 pass |
| abc132 | Task 11: CI | worktree-agent-abc13250 | Merged | N/A |

## Merge Log

| Branch | Merged To | Commit | Conflicts |
|--------|-----------|--------|-----------|
| — | — | — | — |

## Notes

- PRG files (807MB) are local-only at `data/ediabas/ecu/` — Tasks 4, 5, 10 must run on this machine
- SP-Daten, SGDAT, translations are in repo via Git LFS
- All tasks are TDD: test first, implement second, commit after green
