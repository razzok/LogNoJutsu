# Pitfalls Research

**Domain:** Adding realistic attack simulation to a professional SIEM validation tool (consulting engagement context — v1.3)
**Researched:** 2026-04-09
**Confidence:** HIGH (executor/playbook patterns from codebase direct inspection; AV/EDR behaviour from multiple confirmed sources; Windows privilege requirements from official Microsoft docs)

---

## Critical Pitfalls

### Pitfall 1: Windows Defender Quarantines the Go Binary Itself

**What goes wrong:**
Windows Defender's ML heuristic (`Trojan:Win32/Wacatac.C!ml` family) flags Go binaries as malware — even clean, benign ones — when they match generic threat patterns in the heuristic database. This is a known, recurring issue across the Go ecosystem (go/issues#44323, microsoft/go#1255, wailsapp/wails#3308). Adding realistic attack-pattern strings — `comsvcs.dll`, `MiniDump`, `lsass`, `0x1010`, `OpenProcess`, `net user /add` — to the binary's embedded YAML playbooks or compiled strings makes quarantine far more likely. The binary gets removed silently from the client machine before the consultant runs it, stalling the engagement.

**Why it happens:**
Go binaries compile all embedded data (YAML playbooks, string constants) directly into the binary image. The more realistic the technique command strings, the higher the heuristic similarity score to known malware. The `-ldflags -s -w` production stripping that generates a clean binary actually increases false positive rates because stripped executables structurally resemble malware that also strips debug information to evade static analysis. The existing lognojutsu binary already embeds T1003.001 LSASS commands and `comsvcs.dll MiniDump` in the YAML — the v1.3 expansion makes this worse.

**How to avoid:**
1. Before each client delivery, scan the build artifact with `MpCmdRun.exe -Scan -ScanType 3 -File lognojutsu.exe` and document the result in the engagement package.
2. Include a path-based Defender exclusion (not process exclusion) in the pre-engagement checklist: `Add-MpPreference -ExclusionPath "C:\Engagements\LogNoJutsu\"`. Clients apply this before extracting the binary.
3. Consider signing the binary with a code-signing certificate — this substantially reduces ML-based false positive rates.
4. Do NOT add attack-technique strings as Go `const` or package-level `var` — keep them in YAML playbooks, which can be neutrally named.
5. During development, test builds against Defender before tagging releases.

**Warning signs:**
- Consultant reports "the exe disappeared" or "Defender blocked it" after copying to the client machine
- Binary appears in `%ProgramData%\Microsoft\Windows Defender\Quarantine\`
- Binary vanishes 5–15 seconds after extraction from a zip

**Phase to address:**
Earliest phase that introduces new attack-pattern strings in YAML or Go source. Add a CI step that runs `MpCmdRun.exe` against the build output on a Windows runner. Add the Defender exclusion path requirement to the engagement handoff checklist.

---

### Pitfall 2: AMSI Silently Kills PowerShell Techniques Without a Useful Error

**What goes wrong:**
AMSI (Anti-Malware Scan Interface) intercepts every PowerShell script block at runtime, scanning content against current AV signatures. When realistic technique commands contain strings like `MiniDump`, `OpenProcess`, `comsvcs.dll`, or LDAP enumeration patterns matching known malicious scripts, AMSI terminates the PowerShell session mid-execution. The current executor (`executor.go:194-216`) returns `Success: false` with an empty `ErrorOutput` — because AMSI kills the process before it can write output. The consultant cannot distinguish "AMSI blocked the technique" from "technique executed but no SIEM event was generated."

**Why it happens:**
`runCommand` uses `-ExecutionPolicy Bypass`, which bypasses script signing requirements but does NOT bypass AMSI. AMSI operates below the execution policy layer, intercepting script content before any policy check. The existing echo-stub commands (`Write-Host "[T1046] simulation"`) never triggered AMSI because they contained no malware-like patterns. Real attack commands — `comsvcs.dll MiniDump`, `OpenProcess(0x1010, ...)`, `DirectorySearcher.FindAll()` with broad filters — do.

**How to avoid:**
1. Detect AMSI-blocked failures in the verifier: when a PowerShell technique returns `Success: false`, `Output == ""`, and `ErrorOutput == ""` within under 2 seconds, classify the result as `"blocked_endpoint_security"` rather than `"execution_error"`. Surface this to the consultant in the report.
2. Document that clients must add a Defender AMSI process exclusion for `lognojutsu.exe` and `powershell.exe` (when launched by lognojutsu) in the pre-engagement checklist.
3. Never attempt to bypass AMSI programmatically in the tool — this transforms a simulation tool into offensive tooling and creates severe legal and reputational risk.
4. For techniques that are frequently AMSI-blocked, consider whether the technique can be implemented as a native Go call (e.g., Win32 API via `syscall`/`golang.org/x/sys/windows`) instead of shelling out to PowerShell. Native Go calls are not scanned by AMSI.

**Warning signs:**
- Techniques that pass in the consultant's test lab fail on all client machines
- `ErrorOutput` is empty but `Success` is false on PowerShell techniques
- Client has MDE (Microsoft Defender for Endpoint) with cloud-delivered protection enabled

**Phase to address:**
Phase implementing realistic technique commands. Add AMSI failure detection to the verifier/reporter before any new realistic technique is deployed. Also add it as a classification to the HTML report — "Blocked by endpoint security" is a distinct finding from "SIEM missed the event."

---

### Pitfall 3: Network Scan Triggers IDS/NDR Alerts That Contaminate SIEM Validation Results

**What goes wrong:**
The planned real ARP/ICMP + TCP port scan on the local /24 subnet generates exactly the burst pattern that IDS/NDR appliances are tuned to detect (threshold-based: N connections from one source IP in T seconds). Many client environments have NDR appliances (Darktrace, ExtraHop, Vectra) or IDS (Snort/Suricata with Nmap detection rules) that automatically isolate the scanning host, block outbound connections, or generate so many alerts that they drown out the SIEM validation results. Worse: the technique may show as "fail" in the verification engine because IDS blocked the TCP connection before the endpoint could generate a Sysmon EID 3 event — not because the SIEM rule is missing.

**Why it happens:**
A scan from a single PID hitting 17+ ports simultaneously (as the current T1046 runspace implementation does) is a textbook IDS trigger signature. The SIEM and the IDS/NDR are independent systems — the IDS can block traffic that the SIEM would have detected. There is currently no way in the verification engine to distinguish "IDS blocked the scan" from "SIEM missed the event."

**How to avoid:**
1. Add an explicit pre-execution warning in the Web UI before any network discovery technique: "Network scanning generates IDS/NDR alerts. Confirm the scanning host is in the IDS allowlist OR the IDS is in monitor-only mode."
2. Make the scan target range an explicit required configuration field — do NOT auto-detect and auto-scan the /24. The consultant must enter the target range.
3. Implement configurable scan rate: a "slow" mode (1 connection per port per 200ms) for basic coverage testing, and a "burst" mode for behavioral detection testing. Default to slow.
4. Limit the default port list to the 15–20 ports that generate meaningful Sysmon events, not broad discovery sweeps.
5. Document IDS allowlist requirements in the pre-engagement checklist alongside the Defender exclusion requirements.

**Warning signs:**
- Client network team raises an incident ticket before SIEM validation completes
- Network scan technique shows "fail" even though Sysmon is confirmed running on the host
- Consultant's workstation is isolated from the network mid-engagement

**Phase to address:**
Phase implementing real network discovery. Target range configuration UI and rate-limiting must be designed and built before the scan implementation, not added afterward.

---

### Pitfall 4: Raw ICMP / ARP Scan Requires Elevation and Fails Silently Without It

**What goes wrong:**
Real host discovery using ICMP echo requests or raw ARP packets requires `SOCK_RAW` on Windows. Microsoft's Win32 API documentation explicitly states: "To use a socket of type SOCK_RAW requires administrative privileges." The current tool requires local admin for the Attack phase but the v1.3 requirement places real network discovery in the Discovery phase, which is documented as not requiring elevation. If the implementation uses `net/icmp` or `golang.org/x/net/icmp` without elevation, the call returns `WSAEACCES (10013)`. If this error is not explicitly handled, the scan silently runs in TCP-connect-only mode, generating a different (less realistic) Sysmon EID 3 pattern, and the verification engine may still report "pass."

**Why it happens:**
Go's `net/icmp` package follows the same Windows privilege rules as the Win32 API. The documentation for `icmp.ListenPacket` on Windows states it requires an elevated token. Non-elevated fallback to TCP connect succeeds but produces a categorically different scan artifact.

**How to avoid:**
1. At startup, detect the current elevation level (`IsUserAnAdmin()` via `golang.org/x/sys/windows`) and expose it in the Web UI status bar.
2. Implement two explicit code paths:
   - Elevated: raw ICMP + ARP scan (full realistic artifact)
   - Non-elevated: TCP connect scan (reduced artifact, different Sysmon pattern)
3. Label the execution result with the mode used: `"network_scan_mode": "icmp+arp"` vs. `"network_scan_mode": "tcp_connect"`.
4. Do NOT silently fall back — display "Running TCP-only scan (no elevation)" in the technique output.

**Warning signs:**
- ICMP scan returns zero hosts while TCP connect finds active hosts
- `WSAEACCES` errors swallowed in error handling with `_ = err`
- Network scan technique reports "pass" using a fundamentally different artifact than what was tested during development

**Phase to address:**
Phase implementing real network discovery. Elevation detection logic must be written and tested before the raw socket implementation.

---

### Pitfall 5: Legitimate Tool Execution Triggers EDR Behavioral Rules Independent of the SIEM

**What goes wrong:**
Using real tools — `ldifde.exe`, `dsquery.exe`, `net.exe`, `nltest.exe`, `rundll32.exe` (already used in T1003.001), `schtasks.exe` — triggers EDR behavioral detection rules that operate completely independently of the SIEM. CrowdStrike Falcon, SentinelOne, and MDE all have built-in behavioral detections for "AD enumeration tooling spawned by unknown parent process." The EDR may kill the child process, send an alert to the client's SOC (creating an unexpected incident during a controlled test), or quarantine the `lognojutsu.exe` parent process. The process tree `lognojutsu.exe → powershell.exe → ldifde.exe` is unusual and matches attacker patterns regardless of intent.

**Why it happens:**
The executor always shells out via `exec.Command("powershell.exe", ...)` — every technique runs as a child of the `lognojutsu.exe` process. EDR tools track process trees. An unknown binary spawning PowerShell which spawns LDAP tooling is a high-fidelity attack pattern. The existing AZURE_ldap_recon.yaml already uses `ldifde.exe` and `dsquery.exe`; v1.3 adds more such techniques.

**How to avoid:**
1. Generate a dynamic EDR exclusion checklist during the Preparation phase: parse all loaded technique YAML files, extract all executable names referenced in `executor.command` fields (regex for `.exe` names), and present the list in the Preparation tab of the Web UI.
2. Include in the pre-engagement checklist: exclude `lognojutsu.exe` as a parent process in the client's EDR, OR suppress alerts for child processes spawned by `lognojutsu.exe` during the engagement window.
3. Add EDR-kill detection: if a technique involving a known LOLBin command (executable names that match a hardcoded list: `ldifde.exe`, `dsquery.exe`, `net.exe`, `nltest.exe`, `schtasks.exe`, `rundll32.exe`, `reg.exe`, `wmic.exe`) fails with empty output in under 2 seconds, log "possible EDR process termination" in the result.
4. For techniques that do not require PowerShell as a wrapper (e.g., direct Win32 API calls), use `exec.Command` to call the tool directly rather than via PowerShell — shorter process tree, fewer EDR trigger points.

**Warning signs:**
- Techniques pass in the consultant's lab but fail on client machines with similar configs
- Client's EDR console shows alerts during the agreed engagement window
- Techniques involving LOLBins fail with 0-byte output and exit code non-zero

**Phase to address:**
Phase implementing realistic technique commands. The dynamic exclusion checklist generation should be built into the Preparation phase before the first client deployment of v1.3.

---

### Pitfall 6: Brute-Force or Credential Simulation Locks Out Real Client Accounts

**What goes wrong:**
Any technique simulating credential attacks (T1110 brute force, T1110.003 password spray, repeated failed logons to generate Event ID 4625 patterns) that iterates over multiple attempts against a real domain account can trigger the client's lockout policy (typically 3–5 failed attempts → 15–30 min lockout). Locking out a service account or a domain admin account mid-engagement causes real operational disruption: printing stops, scheduled tasks fail, helpdesk calls flood in. This is an engagement-ending incident.

**Why it happens:**
Generating one Event ID 4625 is sufficient to prove the SIEM detects failed logons. But a realistic "password spray" SIEM rule keys on a burst of failures across multiple accounts — developers add a loop to generate the burst pattern, accidentally using real AD accounts enumerated from the tool's userstore or from the LDAP recon results. The line between "pattern that triggers the rule" and "actual lockout trigger" is one attempt.

**How to avoid:**
1. Hard-limit all credential failure simulations to exactly 1 failed attempt per account. The SIEM rule fires on a pattern of events — the tool needs only 1 to prove the event type is being generated and forwarded.
2. Create a dedicated `lognojutsu-test` local account during Preparation phase setup and use it as the sole target for all credential attack simulations. Never use accounts from the userstore or from LDAP enumeration as attack targets.
3. Generate Event ID 4625 via `LogonUser` Win32 API with a deliberately wrong password against the test account only — not against any discovered domain accounts.
4. Add a prominent confirmation prompt in the Web UI before any credential attack technique: "This generates failed logon events against [account]. Confirm account lockout policy will not trigger."

**Warning signs:**
- Technique YAML shows a loop iterating over `$users` from a previous LDAP recon result
- No dedicated test account created in the Preparation phase setup steps
- Password spray implementation attempts more than 1 failed logon per account

**Phase to address:**
Phase adding credential attack techniques. Enforce the single-attempt constraint by code review and/or a YAML schema field `max_auth_attempts` validated before execution.

---

### Pitfall 7: Network Scan Scope Creep — Tool Accidentally Touches Out-of-Scope Systems

**What goes wrong:**
A /24 TCP scan on a client network almost certainly includes systems outside the agreed validation scope: OT/SCADA controllers, VoIP phones, printers, network switches, medical devices, building management systems. A TCP connect attempt to a PLC or network-connected medical device can:
- Cause the device to fault or reboot (embedded firmware may not handle unexpected TCP connections gracefully)
- Generate a mandatory reportable incident (healthcare clients under HIPAA, financial clients under GLBA)
- Violate the engagement's Rules of Engagement, creating legal liability for the consulting firm

**Why it happens:**
The tool's planned /24 sweep is a reasonable IT network assumption but is dangerous as a default on any production environment. The consultant running the tool may not know what devices are on the network before the scan runs. If the tool auto-detects the adapter's /24 and starts scanning without explicit confirmation, scope creep is the default outcome.

**How to avoid:**
1. Make the scan target range a required configuration field in the Web UI — no default, no auto-detect. Require explicit entry before the technique can execute.
2. Add a mandatory confirmation dialog: "You are about to scan [user-entered range]. This range must be within the agreed engagement scope. Confirm?"
3. Consider a pre-scan ARP-only pass that lists discovered hosts for the consultant to review before any port scanning begins.
4. In the engagement handoff document, add a field: "Approved IP ranges for network scanning." If this field is empty, the network scan technique does not run.

**Warning signs:**
- Default target range is auto-populated from the network adapter without prompting
- No confirmation step exists between "start simulation" and "port scan begins"
- Engagement scoping document has no "approved scan targets" field

**Phase to address:**
Phase implementing real network discovery. The target range UI and confirmation dialog are prerequisites to the scan implementation, not polish to add afterward.

---

### Pitfall 8: Temp File Artifacts Left on Client Machine After Abort or AV Quarantine

**What goes wrong:**
Several existing techniques write files to `%TEMP%`: LSASS dumps (`lognojutsu_lsass.dmp`), LDIF exports (`lognojutsu_ldap_export.ldf`), and RunAs temp scripts (`lnj_*.ps1`). Cleanup commands remove these during normal execution, but cleanup does NOT run if:
- The simulation is aborted mid-technique
- Windows Defender quarantines the lognojutsu process before cleanup executes
- A technique times out and the engine moves on

A `.dmp` file or `.ldf` file left on a client machine is a compliance violation — GDPR and SOC 2 auditors treat uncontrolled AD exports or process memory files as data leakage events. A lognojutsu LSASS dump file on a client machine is particularly serious because LSASS dumps contain process memory that may include credential material.

**Why it happens:**
`RunWithCleanup` in `executor.go:30-40` runs cleanup only after execution completes. `RunCleanupOnly` exists for abort scenarios but depends on the engine calling it. If the Go process is killed (Defender quarantine, system shutdown, consultant force-kills the UI), no deferred cleanup runs. The `lnj_*.ps1` temp script files in `runCommandAs` use `defer os.Remove(scriptFile)` — Go deferred cleanup only executes on normal process exit.

**How to avoid:**
1. Add a startup artifact scan: at process start, scan `%TEMP%` for any `lognojutsu_*` and `lnj_*` files from previous runs and delete them before the first technique executes.
2. Use `os.MkdirTemp` to create a dedicated `%TEMP%\lognojutsu-[session-id]\` subdirectory for all temp files, and call `os.RemoveAll` on that directory as a deferred on `main()`. Entire session artifacts are cleaned even on normal exit.
3. For LDAP techniques that currently write LDIF to disk, pipe the output through stdout capture instead — avoid writing AD data to disk entirely where possible.
4. Never write data from LDAP queries (user lists, computer lists) to disk in any form that persists beyond the technique execution window.

**Warning signs:**
- `RunWithCleanup` not used for all technique types that write files to disk
- Cleanup commands use `-ErrorAction SilentlyContinue` without logging the skip
- No startup artifact scan exists in `main.go`

**Phase to address:**
Phase implementing realistic technique commands — specifically, before any technique that writes real data to disk. The startup cleanup routine is a hard prerequisite.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| `-ExecutionPolicy Bypass` covers all execution restrictions | No per-technique policy configuration | Does NOT bypass AMSI; silently fails on hardened clients with no diagnostic output | Never acceptable as the sole approach for realistic technique commands |
| Auto-detect /24 from network adapter for scan target | Zero setup for consultant | Scans OT/SCADA/medical devices without consent; creates liability | Never acceptable — always require explicit confirmation |
| `exec.Command("powershell.exe", ...)` for all techniques | Consistent code path | Every technique shares the same suspicious process tree; EDR blocks the pattern | Acceptable for techniques that genuinely need PowerShell; avoid for techniques that can be implemented as direct exec or Go native calls |
| Broad LDAP `objectClass=*` with high SizeLimit | Simulates realistic attacker enumeration | May time out or return 50k+ objects on large domains; may trip LDAP query cost alerts on DC | Only use with low SizeLimit (already 1000 in AZURE_ldap_recon); never remove the limit |
| Leave lsass.dmp on disk during technique execution | Simplest cleanup approach | If process dies between dump creation and cleanup, credential-adjacent process memory sits on client disk | Never acceptable to leave without immediate cleanup; use `defer Remove` at minimum |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Windows Defender / AMSI | Assume `-ExecutionPolicy Bypass` also bypasses AMSI | AMSI is independent of execution policy; it intercepts script content before execution policy applies. Treat AMSI blocks as a separate failure classification in the result |
| CrowdStrike Falcon / SentinelOne (client EDR) | Assume the SIEM team has whitelisted the tool | Falcon and SentinelOne have independent behavioral rules separate from the SIEM; the EDR and SIEM are different systems. Pre-engagement checklist must cover both |
| Active Directory LDAP | Run `objectClass=*` queries without SizeLimit | Large AD environments (50k+ objects) will exhaust memory and may trigger LDAP query cost alerts (EID 1644) that generate unexpected SIEM noise during validation |
| Event ID 4662 (AD object access) | Expect EID 4662 without pre-configuring audit policy | EID 4662 requires Directory Service Access audit policy AND SACLs configured on AD objects — neither exists by default. Technique executes successfully but generates zero verifiable events |
| Raw ICMP via `net/icmp` or `golang.org/x/net/icmp` | Assume the package works without elevation on Windows | Requires Administrator token on Windows; falls back to TCP-connect on error, which produces a different Sysmon event pattern |
| MDE cloud-delivered protection | Test on local Defender, deploy to client with MDE active | MDE's cloud protection has faster signature updates than local Defender; a technique that passed local testing may be blocked by MDE's real-time cloud analysis |
| `exec.Command` for LOLBin execution | Assume direct exec avoids PowerShell process tree issues | The parent process is still `lognojutsu.exe`; EDR tracks the full tree. The process tree `lognojutsu.exe → ldifde.exe` is unusual regardless of PowerShell intermediary |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Synchronous /24 TCP scan without per-host timeout | Scan hangs indefinitely on hosts dropping packets (no RST) | Use `WaitOne(300ms, false)` per connection (already in T1046); add outer goroutine timeout for the full scan | Any network with RFC 3330 reserved space or hosts with host-based firewalls silently dropping SYNs |
| LDAP queries without SizeLimit | Query hangs or returns 50k+ objects exhausting memory | Always enforce `SizeLimit` (already set in AZURE_ldap_recon.yaml to 1000); do not remove this cap when adding new LDAP techniques | Domain with > 10k objects and unrestricted SizeLimit |
| Raw ICMP probe for all 254 /24 hosts in parallel | 254 simultaneous ICMP sends; some AV products flag burst ICMP | Rate-limit: no more than 16 concurrent ICMP probes; /24 scan should take at least 10 seconds | Any environment with volume-based ICMP detection |
| Generating Sysmon EID 3 burst from runspace pool with too many threads | OOM on low-memory target machines under load | Cap runspace pool to 10 (already done in T1046); do not increase for "more realistic" scans | Systems with < 4GB RAM under high concurrent load |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Attempting to bypass AMSI programmatically | Transforms simulation tool into offensive tooling; violates engagement terms; legal liability for consulting firm | Never implement AMSI bypass — document it as a client pre-configuration requirement |
| Writing credentials to temp `.ps1` files that survive process death | Credentials at rest in TEMP if cleanup fails (already exists in `runCommandAs`) | Add startup cleanup; use pipes instead of temp files for RunAs where possible |
| Scanning networks beyond agreed scope | Device disruption, mandatory incident reporting, legal liability | Require explicit target range input; no auto-detect; mandatory confirmation dialog |
| Leaving lsass.dmp files on client machines | File may contain credential-adjacent memory; GDPR/SOC 2 finding | Startup cleanup for leftover `lognojutsu_lsass*.dmp` files; use `defer Remove` in technique execution |
| Running credential spray with real domain user list from LDAP recon | Account lockout causing production outage | Dedicated test account only; single-attempt limit enforced |
| Implementing LOLBin techniques without EDR exclusion guidance | EDR creates incident ticket in client SOC during controlled test | Pre-engagement checklist with dynamic exclusion list generated from technique YAML |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No pre-scan network range confirmation | Consultant scans OT/SCADA devices; engagement becomes an incident | Require explicit IP range entry with confirmation dialog before any network scan technique |
| No dynamic EDR exclusion checklist in Preparation tab | Consultant runs techniques that get blocked; entire report shows failures; wasted engagement | Parse technique YAML at load time; generate process exclusion checklist in Preparation tab |
| AMSI-blocked failure looks identical to "technique execution error" | Consultant assumes SIEM gap when real cause is endpoint security | Add result classification: `blocked_by_endpoint_security` distinct from `execution_error`, shown in report |
| No indication of network scan mode (ICMP vs TCP-only) | Report says "T1046 detected" but technique ran a degraded non-elevated TCP-only pattern | Show scan mode in technique output and in the HTML report detail |
| Cleanup failures suppressed by `-ErrorAction SilentlyContinue` | Client machine retains artifacts; compliance risk | Log cleanup failures in technique result; surface as warning in the HTML report |
| Network scan target range auto-populated from network adapter | Consultant doesn't realise they're about to scan production OT | Always blank; require active consultant input |

---

## "Looks Done But Isn't" Checklist

- [ ] **Network scan target range:** Verify the scan does not start without explicit consultant input — no auto-detect, no default value
- [ ] **EDR exclusion checklist:** Verify the Preparation tab generates a process exclusion list derived from loaded technique YAML files
- [ ] **AMSI failure classification:** Verify technique results distinguish "blocked by endpoint security" from "execution error" in both the result struct and the HTML report
- [ ] **Startup artifact cleanup:** Verify `main.go` or engine startup scans and removes `lognojutsu_*` and `lnj_*` temp files from previous runs before the first technique executes
- [ ] **Credential attack single-attempt enforcement:** Verify no technique attempts more than 1 failed authentication against any account — no loops over userstore or LDAP-discovered accounts
- [ ] **Elevation mode detection for network scan:** Verify ICMP mode is only attempted when elevated, and non-elevated mode is explicitly labelled in the output

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Defender quarantines binary on client | HIGH | Restore from quarantine with IT help; apply path exclusion; re-deliver binary; add Defender scan to CI to prevent recurrence |
| Network scan triggers IDS incident | HIGH | Stop simulation immediately; produce engagement authorisation letter for client security team; coordinate IDS allowlist addition; reschedule network scan phase |
| Account lockout from credential simulation | HIGH | Client AD admin unlocks affected accounts; document as tool bug; implement single-attempt limit before any re-engagement |
| AMSI blocks technique mid-execution | MEDIUM | Document as "endpoint security configuration gap" in the report — if the technique can't execute, the SIEM rule can never fire, which is itself a SIEM coverage finding |
| EDR quarantines child process (ldifde, dsquery) | MEDIUM | Add process to EDR exclusion list; re-run specific technique; or reclassify as "EDR-blocked, SIEM validation not possible without exclusion" — this is a valid engagement finding |
| Leftover temp files after abort | MEDIUM | Run startup cleanup routine or manual `Remove-Item $env:TEMP\lognojutsu_*`; add startup cleanup to prevent recurrence |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Defender quarantines Go binary | Before every client delivery of a new binary version | Run `MpCmdRun.exe -Scan` in CI; document VirusTotal scan result in release notes |
| AMSI blocks PowerShell techniques | Phase implementing realistic technique commands | Add AMSI failure detection: test that a technique returning `Success: false`, empty output, < 2 sec runtime is classified as `blocked_endpoint_security` |
| Network scan triggers IDS alerts | Phase implementing real network discovery | UI requires explicit target range input; rate-limiting implemented and configurable; pre-scan confirmation dialog verified in integration test |
| Raw ICMP requires silent elevation | Phase implementing real network discovery | Test on non-elevated process; verify TCP fallback runs and is labelled correctly in technique output |
| EDR kills child processes | Phase implementing realistic technique commands | Dynamic exclusion checklist in Preparation tab; lab test with Defender in block mode; LOLBin failure classification in result |
| Brute force locks out accounts | Phase adding credential attack techniques | Code review enforces single-attempt constraint; no technique iterates over userstore/AD accounts as targets |
| Scope creep to OT/SCADA systems | Phase implementing real network discovery | Integration test verifies network scan does not execute without explicit consultant-entered target range |
| Temp artifacts left on client | Phase implementing realistic technique commands (first write-to-disk technique) | Startup cleanup test: create `lognojutsu_test.tmp` in TEMP, start engine, verify it is removed before first technique runs |

---

## Sources

- [Microsoft: Go binary false positive detections — microsoft/go#1255](https://github.com/microsoft/go/issues/1255)
- [Go: Virus detected in Go 1.16 Windows binaries — golang/go#44323](https://github.com/golang/go/issues/44323)
- [Wails: Defender false detects empty Go project — wailsapp/wails#3308](https://github.com/wailsapp/wails/issues/3308)
- [Microsoft Learn: TCP/IP raw sockets on Windows require Administrator privilege](https://learn.microsoft.com/en-us/windows/win32/winsock/tcp-ip-raw-sockets-2)
- [Microsoft Learn: AMSI integration with Microsoft Defender Antivirus](https://learn.microsoft.com/en-us/defender-endpoint/amsi-on-mdav)
- [Atomic Red Team: Issue #3158 — PowerShell security protections blocking tests](https://github.com/redcanaryco/atomic-red-team/issues/3158)
- [Red Canary: Running Atomic Red Team with Microsoft Defender for Endpoint](https://redcanary.com/blog/microsoft/atomic-red-team-mde/)
- [Vaadata: Active Directory Monitoring — LDAP Log Analysis and ELK Rules](https://www.vaadata.com/blog/active-directory-monitoring-ldap-log-analysis-and-elk-rules/)
- [BlackLanternSecurity: Detecting LDAP Reconnaissance](https://blog.blacklanternsecurity.com/p/detecting-ldap-recoannaissance)
- [UltimateWindowsSecurity: Event ID 4662 — Object operation on AD object](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4662)
- [Nmap Book: Subverting Intrusion Detection Systems](https://nmap.org/book/subvert-ids.html)
- [Penetration Testing Authority: Rules of Engagement](https://penetrationtestingauthority.com/rules-of-engagement-penetration-testing)
- [OWASP: Testing for Weak Lock Out Mechanism](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/04-Authentication_Testing/03-Testing_for_Weak_Lock_Out_Mechanism)
- Codebase direct inspection: `internal/executor/executor.go`, `internal/playbooks/types.go`, technique YAML files (`T1046_network_scan.yaml`, `AZURE_ldap_recon.yaml`, `T1003_001_lsass.yaml`, `T1021_002_smb_shares.yaml`)

---
*Pitfalls research for: LogNoJutsu v1.3 Realistic Attack Simulation*
*Researched: 2026-04-09*
