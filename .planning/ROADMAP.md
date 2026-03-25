# Roadmap: LogNoJutsu

## Overview

Milestone 1 — Verified & Expanded: Make LogNoJutsu trustworthy for professional client delivery. Add verification so silent technique failures are caught, improve code quality so changes are safe, and expand technique and SIEM platform coverage (CrowdStrike, Sentinel).

## Phases

- [x] **Phase 1: Events Manifest & Verification Engine** - Add expected Event IDs per technique and post-run pass/fail verification against local Windows Event Log (completed 2026-03-24)
- [x] **Phase 2: Code Structure & Test Coverage** - Refactor package-level globals to a struct, split into packages, add unit tests for engine, handlers, and verification logic (completed 2026-03-25)
- [ ] **Phase 3: Additional Techniques** - Add 5+ new MITRE ATT&CK techniques and 3+ Exabeam UEBA scenarios, all with events manifest entries
- [ ] **Phase 4: CrowdStrike SIEM Coverage** - Add CrowdStrike Falcon detection mappings and Falcon-sensor-specific techniques to the events manifest and HTML report
- [ ] **Phase 5: Microsoft Sentinel Coverage** - Add Microsoft Sentinel / Azure AD detection mappings and Azure-specific techniques to the events manifest and HTML report

## Phase Details

### Phase 1: Events Manifest & Verification Engine
**Goal**: Each technique declares which Event IDs it should produce. After a run, the tool queries the local Windows Event Log and reports pass/fail per technique, shown in the HTML report.
**Depends on**: Nothing (first phase)
**Requirements**: VERIF-01, VERIF-02, VERIF-03, VERIF-04, VERIF-05
**Success Criteria** (what must be TRUE):
  1. Every technique struct has an `ExpectedEvents` field listing Event IDs and log channels
  2. After a simulation run, verification queries Windows Event Log and stores pass/fail status per technique
  3. HTML report shows a verification column: expected events, found events, and pass/fail indicator per technique
  4. A technique that fails to execute is flagged as "not verified" — distinct from "events not found in log"
  5. All existing techniques are populated with their correct expected Event IDs
**Plans:** 3/3 plans complete
Plans:
- [x] 01-01-PLAN.md — Core types (EventSpec, VerificationStatus, VerifiedEvent), verifier package, engine integration
- [ ] 01-02-PLAN.md — Migrate all 43 technique YAML files to structured EventSpec format
- [x] 01-03-PLAN.md — HTML report verification column with pass/fail badges and per-event breakdown

### Phase 2: Code Structure & Test Coverage
**Goal**: Refactor global state so HTTP handlers are testable. Add unit tests for engine, handlers, and verification logic. Split code into packages.
**Depends on**: Phase 1
**Requirements**: QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05
**Success Criteria** (what must be TRUE):
  1. Package-level globals in server.go replaced by an `AppState` struct
  2. Code split into at least: engine/, server/, techniques/, reporter/, verifier/ packages
  3. `go test ./... -race` passes with no failures or race conditions
  4. HTTP handler tests use httptest.NewRecorder — no global state required for test setup
  5. Engine state machine transitions covered by unit tests
**Plans:** 3/3 plans complete
Plans:
- [x] 02-01-PLAN.md — Fix vet warning, Server struct refactor (QUAL-01), RunnerFunc injection on Engine
- [x] 02-02-PLAN.md — Engine unit tests (QUAL-03), verifier D-11 named tests (QUAL-05)
- [x] 02-03-PLAN.md — HTTP handler tests with httptest (QUAL-04), full suite validation

### Phase 3: Additional Techniques
**Goal**: Expand technique library with at least 5 new ATT&CK techniques and 3 new Exabeam UEBA scenarios. All new techniques include events manifest entries.
**Depends on**: Phase 1
**Requirements**: TECH-01, TECH-02, TECH-03
**Success Criteria** (what must be TRUE):
  1. At least 5 new MITRE ATT&CK techniques added (Discovery or Attack phase)
  2. At least 3 new Exabeam UEBA scenarios added
  3. All new techniques have ExpectedEvents populated (pass verification from Phase 1)
  4. README updated to document new techniques
**Plans:** 2/3 plans executed
Plans:
- [x] 03-01-PLAN.md — 5 new ATT&CK YAML files: Collection (T1005, T1560.001, T1119) and C2 (T1071.001, T1071.004)
- [x] 03-02-PLAN.md — 4 new UEBA scenario YAML files + TestExpectedEvents loader test for TECH-03 validation
- [ ] 03-03-PLAN.md — README documentation for all 9 new techniques in German

### Phase 4: CrowdStrike SIEM Coverage
**Goal**: Add CrowdStrike Falcon detection mappings and techniques that generate Falcon sensor events. HTML report shows a CrowdStrike coverage column.
**Depends on**: Phase 1
**Requirements**: CROW-01, CROW-02, CROW-03
**Success Criteria** (what must be TRUE):
  1. Each technique has a SIEMCoverage map with CrowdStrike detection rule names where applicable
  2. At least 3 techniques specifically target and generate Falcon sensor events
  3. HTML report shows a CrowdStrike coverage column listing mapped detection rules per technique
  4. Documentation covers CrowdStrike-specific prerequisites and setup
**Plans**: TBD

### Phase 5: Microsoft Sentinel Coverage
**Goal**: Add Microsoft Sentinel / Azure AD detection mappings and techniques targeting Azure-specific log sources. HTML report shows a Sentinel coverage column.
**Depends on**: Phase 1
**Requirements**: SENT-01, SENT-02, SENT-03
**Success Criteria** (what must be TRUE):
  1. Each technique has Sentinel detection rule mappings in SIEMCoverage map where applicable
  2. At least 3 techniques target Azure AD / Microsoft Defender / Sentinel log sources
  3. HTML report shows a Sentinel coverage column listing mapped analytic rule names per technique
  4. Documentation covers Sentinel-specific prerequisites and setup
**Plans**: TBD
