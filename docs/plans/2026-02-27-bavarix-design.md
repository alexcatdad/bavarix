# Bavarix Design Document

**Date:** 2026-02-27
**Status:** Draft — awaiting approval

## Vision

Open-source, cross-platform BMW diagnostics, coding, and flashing tool. Replaces the entire legacy Windows-only BMW toolchain (EDIABAS, INPA, NCS Expert, Tool32, WinKFP) with a single modern application that runs on Mac, Linux, and Windows.

## Target Audience

Any BMW owner with a DS2/KWP2000-era car. Community-driven, fully open source (MIT or GPL).

## Supported Chassis (initial data available)

E36, E38, E39, E46, E52, E53, E60, E65, E70, E83, E85, E89

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go |
| Web UI | React |
| Transport | Go interfaces (serial, WiFi, BT) |
| Database | SQLite (local) |
| LLM Assistant | Claude API + RAG |
| Distribution | Single binary per platform + static web assets |
| License | MIT or GPL (TBD) |
| i18n | English, Romanian at launch |

## Architecture

```
+----------------------------------------------------+
|                   Web UI (React)                    |
|         diagnostics . coding . flashing             |
|         LLM assistant . backup vault                |
+-----------------------+----------------------------+
                        |
+-----------------------v----------------------------+
|              ASP.NET  Go HTTP + WebSocket API       |
+----------------------------------------------------+
                        |
+-----------------------v----------------------------+
|              Safety Pipeline                        |
|  voltage -> backup -> validate -> confirm ->        |
|  write -> verify                                    |
+----------------------------------------------------+
                        |
+-----------------------v----------------------------+
|           Community Overlay Layer                   |
|   human-readable names . safety ratings . docs      |
|   i18n translations . compatibility tags            |
+----------------------------------------------------+
                        |
+-----------------------v----------------------------+
|             Protocol Engine                         |
|          DS2 . KWP2000 . PRG Parser                 |
+----------------------------------------------------+
                        |
+-----------------------v----------------------------+
|            Transport Layer (interface)              |
|   USB Serial (K+DCAN)  |  WiFi ELM327  |  BT ELM327|
|   [full read+write]    |  [read-only]   | [read-only]|
+----------------------------------------------------+
```

## Safety Pipeline

Every write operation (coding or flashing) must pass through all stages. No exceptions, no bypass.

### Stage 0: Voltage Check

Query battery voltage via OBD/DME before any write operation.

| Operation | Minimum Voltage | Behavior |
|-----------|----------------|----------|
| Single module coding | 12.0V hard block, 12.5V warning | Quick writes, lower risk |
| Multi-module coding | 12.5V hard block | Multiple writes, more time on bus |
| Flashing | 13.0V hard block | Charger required, no exceptions |

### Stage 1: Backup

Before ANY write, read and store the module's current state in the local backup vault (SQLite). Users accumulate a full history of every module state. Rollback to any previous state.

### Stage 2: Validate

Check new values against known-safe ranges from the community overlay. Unknown values get flagged. Dangerous combinations (e.g., disabling airbag via coding) get hard-blocked.

### Stage 3: Confirm

Show visual diff (old vs new) in plain English, not raw hex. User must explicitly approve.

### Stage 4: Write

Send to ECU with checksum verification.

### Stage 5: Verify

Immediately read back the module after writing and compare against what was written. If mismatch, alert the user. Never report success without verification.

### Additional Flashing Guards

- Firmware checksum validation before sending
- Progress tracking with no timeout surprises
- Connection drop detection with clear recovery instructions
- Refuse flash below 13.0V (charger required)

### Operation Logging

Every action timestamped and logged. Full audit trail for diagnostics if something goes wrong.

## Transport Layer

Designed as a Go interface so new adapters can be added:

```go
type Transport interface {
    Connect() error
    Disconnect() error
    SendFrame(frame []byte) error
    ReceiveFrame() ([]byte, error)
    SupportsWrite() bool
    ReadVoltage() (float64, error)
}
```

| Adapter | Read | Code | Flash | Connection |
|---------|------|------|-------|------------|
| K+DCAN USB (FTDI) | Full | Yes | Yes | USB serial |
| ENET cable | Full | Yes | Yes | Ethernet |
| WiFi ELM327 | Basic OBD-II | No | No | WiFi |
| BT ELM327 | Basic OBD-II | No | No | Bluetooth |

Write operations are **hard-blocked at the transport level** for adapters that don't support raw K-line access.

## Community Overlay Layer

Sits on top of parsed PRG/SP-Daten data and adds human context.

### Per-parameter data

```json
{
  "module": "LSZ",
  "byte": 7,
  "bit": 3,
  "name": "Angel eyes as DRL",
  "description": "Enables angel eye rings as daytime running lights",
  "safety": "safe",
  "reversible": true,
  "chassis": ["E46"],
  "hardware": ["LSZ.C28", "LSZ.C29", "LSZ.C31"],
  "requires": ["xenon_headlights"],
  "conflicts": [],
  "verified_by": ["@user1", "@user2"],
  "last_verified": "2026-01-15"
}
```

### Safety Ratings

| Rating | Meaning | UX Behavior |
|--------|---------|-------------|
| Safe | Cosmetic/comfort, no risk | Green, one-click |
| Caution | Works but verify after | Yellow, extra confirmation |
| Warning | Can cause electrical/functional issues if wrong | Orange, must read explanation |
| Dangerous | Can brick module or affect safety systems | Red, hard block unless expert mode |

### Unverified Parameters

Parameters with no community overlay are still shown (from PRG data) but marked "Unverified — use at your own risk" with a caution gate.

### File Structure

```
/overlays
  /E46
    /GM5.json
    /KMB.json
    /LSZ.json
    ...
  /E39
    /GM3.json
    ...
```

Community contributes via pull requests. JSON schema enforced in CI.

## Internationalization (i18n)

English and Romanian at launch. Structure supports unlimited languages.

```
/i18n
  /en.json
  /ro.json
  /de.json  (future)
  /ru.json  (future)
```

Overlay files stay in English as source of truth. Translations are a separate layer. Community contributes translations via PRs. CI checks for missing keys.

## LLM Assistant

Natural language interface for finding and applying coding changes.

### Flow

1. User types: "How do I enable fog light cornering?"
2. Context builder searches overlay DB by semantic similarity
3. Context builder pulls user's car profile (chassis, modules, current state)
4. Claude API called with: system prompt + relevant overlays + car context + question
5. Returns: which parameter, step-by-step instructions, safety warnings
6. "Apply Change" button routes through full safety pipeline

### Safety Rule

The LLM can **suggest** changes but never **execute** them directly. All changes go through the safety pipeline with explicit user confirmation.

## Testing Strategy

### Unit Tests (every commit)

- Protocol framing (DS2/KWP2000 checksums, frame building/parsing)
- Safety pipeline logic (voltage gates, value validation, diff generation)
- PRG parser (known files, expected output)
- Overlay loader (JSON parsing, schema validation)
- Backup vault (write/read/rollback)
- i18n completeness

### Integration Tests — Virtual ECU Simulator (every commit)

Simulated ECU bus in Go. Virtual ECUs respond to DS2/KWP2000 exactly like real modules.

- Full read/write cycle against virtual modules
- Safety pipeline with real protocol flow
- Backup creation verification
- Write verification catches mismatches
- Connection drop recovery
- Wrong module version detection
- ELM327 write blocking

### E2E Tests (every PR)

Full app (Go backend + React UI) against virtual car. Headless browser via Playwright.

- Connect, read modules, see values
- Change parameter, see diff, confirm, verify
- Backup vault populated
- LLM assistant returns correct parameters
- Language switching
- ELM327 write operations disabled in UI

### Chaos / Fault Injection (nightly)

| Scenario | Expected Behavior |
|----------|-------------------|
| Connection drops mid-write | Detect, alert, log, show recovery steps |
| ECU returns error after write | Rollback guidance, backup available |
| Voltage drops during coding | Abort before write, alert user |
| Invalid PRG file | Reject, don't expose broken jobs |
| Corrupted overlay JSON | Fall back to raw PRG data, warn |
| Checksum mismatch on read-back | Alert, show mismatch, don't report success |
| Write via ELM327 | Hard block at transport layer |
| Concurrent writes to same module | Queue, prevent concurrent access |

### Recording Proxy

First thing built. Sits between app and cable, logs every byte in/out. Real ECU conversations become golden test data for virtual ECUs.

One real session -> unlimited test runs forever.

### Hardware-in-Loop (manual, pre-release only)

Bench ECU (spare module from junkyard) wired with 12V supply. Full test suite against real hardware. Manual trigger by maintainers before major releases.

### CI Pipeline (GitHub Actions)

```yaml
on: [push, pull_request]

jobs:
  unit:        # every commit
  integration: # every commit
  e2e:         # every PR
  chaos:       # nightly
  lint:        # every commit (golangci-lint, JSON schema, i18n check)
```

## Phase 0: Data Extraction & Due Diligence

Before any serial connection or app code, extract and understand all available data sources.

### Data Sources On Disk

| Source | Location | Format | Records |
|--------|----------|--------|---------|
| PRG files | C:\EDIABAS\Ecu\ | Binary (BEST/2) | 1,192 files |
| SP-Daten | C:\NCSEXPER\DATEN\ | Binary | 10 chassis |
| SGDAT | C:\NCSEXPER\SGDAT\ | Binary (.ipo) | 641 files |
| NCS Dummy translations | C:\NCS Dummy\Translations.csv | CSV | 26,246 entries |
| NCS profiles | C:\NCSEXPER\PFL\ | Text | Profiles |
| CFGDAT | C:\NCSEXPER\CFGDAT\ | INI/Text | Config |

### External Sources

| Source | Value |
|--------|-------|
| EdiabasLib (GitHub) | PRG parser implementation, protocol reference |
| Deep OBD (GitHub) | Another PRG parser, real-world tested |
| BMW community wikis/forums | Coding guides, known-safe values |

### Extraction Steps

1. **Parse PRG files** — Port EdiabasLib's BEST/2 parser to Go. Extract every job name, parameters, return values. Output: structured JSON per module.
2. **Parse SP-Daten** — Reverse engineer binary format. Extract every coding parameter, valid values, byte positions. Output: structured JSON per chassis per module.
3. **Parse NCS Dummy translations** — Map German parameter names to English. Cross-reference with SP-Daten. Output: bilingual parameter database.
4. **Parse SGDAT profiles** — Extract module identification, hardware/software version mappings. Output: version compatibility matrix.
5. **Scrape EdiabasLib + Deep OBD source** — Extract protocol implementation details. Document DS2/KWP2000 frame formats. Output: protocol specification.
6. **Scrape community knowledge** — Coding guides, known-safe values. Output: seed data for community overlay.
7. **Build unified database** — Merge all sources into SQLite. Cross-reference PRG jobs, SP-Daten params, translations, community data. Validate consistency.
8. **Generate confidence report** — Per module per parameter: format understood? translations available? community verified? safe value ranges known? Flag all gaps.

### Database Schema (extensible)

Designed so new data sources slot in without restructuring. Each record tracks its provenance (which source it came from) and confidence level.

## Replaces

| Legacy Tool | Bavarix Equivalent |
|-------------|-------------------|
| EDIABAS | Transport + Protocol Engine |
| INPA | Diagnostics UI |
| NCS Expert | Coding UI + Safety Pipeline |
| NCS Dummy | Coding UI (friendly mode) |
| Tool32 | Raw Job Executor UI |
| WinKFP | Flashing UI (with safety guards) |
| BMW Coding Tool | The entire app |
