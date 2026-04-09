# Feature Research

**Domain:** Breach and Attack Simulation (BAS) — Realistic MITRE ATT&CK technique execution for SIEM validation
**Researched:** 2026-04-09
**Confidence:** HIGH (existing codebase read directly; BAS platform research verified via official GitHub and docs)

---

## Context: What Already Exists

LogNoJutsu v1.2 ships 57 MITRE ATT&CK techniques. Many are already realistic — they use real Windows APIs, Win32 P/Invoke, RunspaceFactory port scans, actual LDAP DirectorySearcher calls, and real process injection patterns. The v1.3 milestone is about:

1. Upgrading the remaining stub/echo techniques to authentic execution
2. Adding real network discovery that scans the live /24 subnet (ARP/ICMP + TCP)
3. Expanding the technique library with new patterns not yet covered

The executor supports `powershell` and `cmd` types only. All techniques run via PowerShell or CMD subprocesses — Go-native execution is not exposed via the YAML technique format. Win32 P/Invoke via `Add-Type` is the correct pattern to access native APIs while preserving PowerShell ScriptBlock logging (EID 4104).

---

## Existing Techniques: Realism Assessment

Based on direct codebase reading, the existing 57 techniques fall into three realism tiers:

**Tier 1 — Already Realistic (no upgrade needed):**
- T1003.001 LSASS: real MiniDump via comsvcs.dll + Win32 OpenProcess(0x1010)
- T1046 Network Scan: real RunspaceFactory parallel TCP scan generating Sysmon EID 3 burst
- T1087 Account Discovery: real net.exe, cmdkey, ADSI enumeration
- T1558.003 Kerberoasting: real KerberosRequestorSecurityToken, setspn.exe, LDAP SPN query
- T1021.001 RDP: real registry changes, firewall rules, cmdkey storage (with cleanup)
- T1110.001 Brute Force: real LogonUser Win32 API + net use IPC$
- T1550.002 Pass-the-Hash: real EID 4648 patterns via net use + Start-Process -Credential
- AZURE_ldap_recon: real DirectorySearcher, ldifde.exe, dsquery.exe
- T1016 Network Config: real ipconfig/arp/route/nbtstat/nslookup burst

**Tier 2 — Partially Realistic (uses real tools but incomplete realism):**
- T1046 scans only 127.0.0.1 and the gateway — not the full /24 subnet
- T1018 uses `net view` and `arp -a` but lacks active ping sweep and nltest DC discovery

**Tier 3 — Audit Required (possible stub/echo remnants):**
Techniques not directly reviewed must be catalogued before any upgrade work begins. The stub audit is the mandatory first deliverable of v1.3.

---

## Feature Landscape

### Table Stakes (Consultants Expect These for v1.3)

Features that, if missing, make v1.3 feel incomplete relative to what the milestone brief promises.

| Feature | Why Expected | Complexity | Safety Boundary | Notes |
|---------|--------------|------------|-----------------|-------|
| Real /24 subnet scan (ARP + ICMP + TCP) | T1046 currently only scans loopback and gateway — consultants need real network discovery artifacts from the live subnet | MEDIUM | SAFE: read-only network probes only. ARP requests and ICMP echo are standard operations that generate Sysmon EID 3 on the scanning host. No modification of remote systems. No cross-subnet traversal. | Replace 127.0.0.1 stub in T1046 with subnet-aware scan using the host's own /24 derived from primary adapter |
| Real T1018 Remote System Discovery | Current T1018 uses `net view` — should include ICMP ping sweep, ARP table read, and nltest DC discovery for full recon artifact set | LOW | SAFE: passive ARP read (`arp -a`) and ICMP echo requests are read-only. `nltest` and `net view` generate EID 4688 for legitimate built-in Windows tools. | `nltest /dclist:$env:USERDNSDOMAIN` for DC enumeration; `Test-Connection -Count 1` for ping sweep |
| Stub technique audit | 57 techniques must be verified — any remaining Write-Host echo stubs need replacement with real command execution that generates the declared expected_events | HIGH | SAFE: audit is read-only analysis. Each upgrade must generate authentic Windows events without destructive side effects. | First deliverable: a catalogue of which techniques are stubs vs. realistic, before any upgrade work begins |
| Diverse attack pattern expansion | Current library is concentrated in Discovery and Credential Access — lateral movement, collection, and exfiltration patterns are underrepresented | HIGH | VARIES per category (see tactic-by-tactic safety table below) | New techniques must follow the same pattern as existing Tier 1 techniques: real Windows API or tool, with cleanup for all write operations |

### Differentiators (What Sets LogNoJutsu Apart from Atomic Red Team)

Atomic Red Team requires PowerShell module installation and internet access at runtime. LogNoJutsu's constraint — single binary, no internet — is its primary distribution advantage. Differentiators build on this constraint rather than fighting it.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Subnet-aware network scan (auto-detect /24) | ART T1046 requires an explicit target IP. LogNoJutsu detects the host's primary adapter subnet automatically — no per-engagement configuration | LOW | Use `Get-NetIPConfiguration` to find the primary adapter; derive the /24 CIDR from the interface address |
| Burst timing signature calibration | Real scanners hit all ports simultaneously; real recon tools run in rapid succession within seconds. SIEM behavioral rules fire on temporal clustering. ART does not calibrate execution timing. | MEDIUM | T1046 already uses RunspaceFactory for parallel execution. T1016 already runs ipconfig/arp/route in rapid burst. New techniques should adopt the same burst pattern. |
| SIEM-specific technique variants | Per-SIEM coverage metadata already in the YAML format via the `siem_coverage` field. New techniques can target specific detection gaps: Exabeam's 22 Kerberoast rules, Sentinel hunting queries, Falcon behavioral AI patterns. | MEDIUM | Existing architecture supports this with no schema changes |
| Cleanup-by-default discipline | ART cleanup is opt-in per invocation. LogNoJutsu should enforce non-empty cleanup commands for all techniques that write to the system — registry, firewall, scheduled tasks, services, files. | LOW | Already correctly implemented for T1021.001 and AZURE_ldap_recon. Needs to be enforced as a mandatory pattern for all new write techniques. |

### Anti-Features (Explicitly Avoid)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Real exploitation (CVE-based attacks) | "More realistic" — mimics real attackers who exploit vulnerabilities | Causes actual damage to the engagement host; violates the core project constraint that the tool must be safe for production deployment targets | Generate the same Windows Event artifacts that post-exploitation behavior produces, without the exploitation step. The SIEM detects the event, not the underlying vulnerability. |
| Cross-subnet scanning | Consultants want to test larger environments | Generates network noise on production segments not authorized for testing; may trigger IDS alerts on out-of-scope systems | Limit subnet scan to the host's own /24 and document this boundary explicitly in technique output |
| Credential extraction (displaying actual hashes) | More authentic Kerberoasting / LSASS simulation | Exposing real password hashes or NTLM hashes during client engagements creates data handling liability | Current approach is correct: generate the Windows Events (EID 4769, Sysmon EID 10) without extracting or displaying actual credential material |
| Real lateral movement (psexec to remote hosts) | Tests SIEM lateral movement detection end-to-end | Requires authenticated access to remote systems; creates unauthorized access risk on engagement targets | Simulate the local-side artifacts of lateral movement: EID 4648, EID 4624 Type 3, Sysmon EID 3 connection attempts — all generated locally without remote execution |
| Go-native direct Win32 executor type | Bypasses PowerShell for faster execution | Removes PowerShell ScriptBlock logging (EID 4104), which is a primary SIEM detection surface. Adding a new executor type also breaks all existing YAML technique files. | Keep all technique execution routed through PowerShell. Win32 P/Invoke via `Add-Type` is the correct pattern — it generates EID 4104 for the script block while accessing native APIs. |
| C2 beacon to external infrastructure | Tests outbound C2 detection in SIEM | External connectivity violates the offline/standalone deployment requirement | Simulate C2 patterns locally: T1071.001 HTTP to 127.0.0.1, T1071.004 DNS queries to internal resolver, T1197 BITS jobs to localhost |
| Ransomware file encryption on user directories | Tests ransomware detection comprehensively | Even reversible encryption on production systems creates unrecoverable risk if cleanup fails | Scope T1486 to %TEMP% only, with mandatory immediate cleanup. Never touch user home directories. |

---

## Tactic-by-Tactic Safety Analysis

This table defines the safety boundary for each MITRE ATT&CK tactic, mapping which techniques can be made realistic without risk and which have hard constraints.

### Discovery (TA0007) — FULLY SAFE across all techniques

All discovery techniques are read-only enumeration. Real Windows tools generate EID 4688 (process creation) which is the primary SIEM detection target.

| Technique | Safe Realistic Implementation | Windows Events Generated |
|-----------|-------------------------------|--------------------------|
| T1016 Network Config | `ipconfig`, `arp -a`, `route print`, `nbtstat`, `nslookup` — already Tier 1 | EID 4688 burst |
| T1018 Remote System | Add: `nltest /dclist`, `Test-Connection -Count 1` ping sweep | EID 4688 + Sysmon EID 3 |
| T1046 Network Scan | Upgrade: RunspaceFactory TCP scan against real /24 hosts | Sysmon EID 3 burst |
| T1057 Process Discovery | `tasklist`, `Get-Process`, `wmic process` — read-only | EID 4688, EID 4104 |
| T1069 Group Discovery | `net localgroup`, `net group /domain` — read-only AD query | EID 4688 |
| T1082 System Info | `systeminfo`, `wmic os`, `Get-ComputerInfo` — read-only | EID 4688, EID 4104 |
| T1083 File Discovery | `Get-ChildItem` on standard paths — read-only | EID 4104 |
| T1087 Account Discovery | `net user`, `Get-LocalUser`, ADSI — already Tier 1 | EID 4688, EID 4104 |
| T1135 Network Share | `net share`, `net view \\localhost` — read-only | EID 4688 |
| T1482 Domain Trust | `nltest /domain_trusts` — read-only AD query | EID 4688 |

**Upgrade priority for v1.3:** T1046 (/24 scan upgrade) and T1018 (ICMP/ARP/nltest additions) are highest value and lowest risk — implement first.

### Credential Access (TA0006) — SAFE with existing patterns

Techniques generate authentication event artifacts without extracting actual credential material.

| Technique | Safe Realistic Implementation | Constraint |
|-----------|-------------------------------|------------|
| T1003.001 LSASS | Already Tier 1 — MiniDump + OpenProcess(0x1010) | Dump file created in %TEMP% and immediately removed |
| T1110.001 Brute Force | Already Tier 1 — LogonUser API + net use IPC$ | Use only 5 wrong passwords; stop before lockout threshold (typically 10 attempts) |
| T1110.003 Password Spray | Spray one password across local accounts — real EID 4625 burst | Same short-list constraint |
| T1558.003 Kerberoasting | Already Tier 1 — KerberosRequestorSecurityToken + LDAP | TGS tickets are cached by OS; no password cracking occurs |
| T1552.001 Credentials in Files | `Select-String` for credential patterns in standard paths — read-only | No file modification |
| T1003.006 DCSync | LDAP replication rights check via DirectorySearcher only | True DCSync calls DRSGetNCChanges and requires Domain Admin + replication right. Simulate reconnaissance phase only. |

### Persistence (TA0003) — SAFE only if cleanup is mandatory

Write operations — registry keys, scheduled tasks, services — are fully reversible. Cleanup must execute regardless of whether the technique body succeeds or fails.

| Technique | Safe Realistic Implementation | Cleanup Required |
|-----------|-------------------------------|-----------------|
| T1053.005 Scheduled Task | `schtasks /create` with benign command (e.g., `cmd /c whoami`) | Yes — `schtasks /delete /f /tn` in cleanup |
| T1547.001 Registry Persistence | Write to `HKCU:\Software\Microsoft\Windows\CurrentVersion\Run` | Yes — `Remove-ItemProperty` in cleanup |
| T1543.003 New Service | `sc create` a disabled service with benign binary path | Yes — `sc delete` in cleanup |
| T1197 BITS Jobs | `bitsadmin /create` pointing to localhost URL | Yes — `bitsadmin /cancel` in cleanup |
| T1546.003 WMI Event Subscription | Real WMI permanent event subscription | Yes — `Remove-WMIObject` in cleanup. THIS IS CRITICAL: WMI subscriptions survive reboots. If cleanup fails, the subscription persists on the engagement host. |

**Hard requirement:** Every new persistence technique must have a tested cleanup command. The WMI subscription (T1546.003) is the highest-risk existing technique and must be audited to confirm cleanup works.

### Defense Evasion (TA0005) — SAFE with reversibility enforcement

| Technique | Safe Realistic Implementation | Cleanup Required |
|-----------|-------------------------------|-----------------|
| T1027 Obfuscated Commands | `-EncodedCommand` PS invocation — generates EID 4104 ScriptBlock | No |
| T1036.005 Masquerading | Copy `cmd.exe` to `%TEMP%\svchost.exe`, execute once | Yes — remove fake binary from %TEMP% |
| T1070.001 Clear Event Logs | **AUDIT REQUIRED** — see note below | Depends on implementation |
| T1218.011 Rundll32 | Invoke benign exported function via rundll32 LOLBin | No |
| T1562.002 Disable Logging | Disable a specific audit subcategory via `auditpol`, then re-enable | Yes — immediately re-enable in same technique body |
| T1574.002 DLL Sideloading | Place a benign DLL in %TEMP% alongside a legitimate binary | Yes — remove DLL from %TEMP% |

**CRITICAL AUDIT: T1070.001 (Clear Event Logs).** If the current implementation runs `wevtutil cl Security` or `wevtutil cl System`, it is destructive on the engagement host — clearing the Security log destroys evidence and cannot be undone. The acceptable implementation is: use `wevtutil el` to enumerate log names (generates EID 4104), and document that EID 1102 would fire on a real log-clear. Never clear Security, System, or Application logs. A safe alternative: create a synthetic custom event log, populate it, then clear it — EID 1102 fires for any cleared log.

### Lateral Movement (TA0008) — SAFE for local-side simulation only

True lateral movement requires connectivity to remote systems. All implementations simulate the local artifacts of what an attacker would generate before or during lateral movement, without executing on remote hosts.

| Technique | Safe Realistic Implementation | What It Does NOT Do |
|-----------|-------------------------------|---------------------|
| T1021.001 RDP | Already Tier 1 — registry changes + firewall rule + cmdkey | Does NOT open an interactive RDP session to any host |
| T1021.002 SMB Shares | `net use \\localhost\IPC$` with explicit credentials | Does NOT copy files to remote shares |
| T1550.002 Pass-the-Hash | Already Tier 1 — EID 4648 patterns via net use + Start-Process | Does NOT inject NTLM hash into LSASS |
| T1021.006 WinRM (new) | `Invoke-Command -ComputerName 127.0.0.1` — loopback WinRM | Requires WinRM to be enabled; generate EID 4648 + EID 4624 Type 3 |

**New opportunity for v1.3:** T1021.006 WinRM against loopback generates EID 4648 and tests WinRM detection rules, which are a common gap in SIEM configurations. Safe because it targets 127.0.0.1 only.

### Collection (TA0009) — SAFE; read-only enumeration with %TEMP% staging

| Technique | Safe Realistic Implementation | Cleanup Required |
|-----------|-------------------------------|-----------------|
| T1005 Data from Local System | `Get-ChildItem` with file type filter across standard user paths | No — read-only |
| T1119 Automated Collection | Loop through %USERPROFILE% collecting file path list | No — read-only |
| T1560.001 Archive Collected Data | `Compress-Archive` on files already in %TEMP% only | Yes — remove zip from %TEMP% |

### Command and Control (TA0011) — SAFE for loopback/internal only

| Technique | Safe Realistic Implementation | Constraint |
|-----------|-------------------------------|------------|
| T1071.001 HTTP C2 | HTTP request to 127.0.0.1 — Sysmon EID 3 network connection | No external connectivity |
| T1071.004 DNS C2 | `Resolve-DnsName` to internal domain with unusual record types | Internal DNS resolver only |
| T1132 Data Encoding | Base64 encode dummy payload in PowerShell — EID 4104 | Read-only |

### Exfiltration (TA0010) — SAFE for loopback simulation

| Technique | Safe Realistic Implementation | Constraint |
|-----------|-------------------------------|------------|
| T1041 Exfiltration over C2 | HTTP POST with dummy payload to 127.0.0.1 — Sysmon EID 3 | No actual data leaves the host |
| T1048 Exfil over Alt Protocol | DNS TXT query with base64 dummy payload to internal resolver | Internal DNS only |

### Impact (TA0040) — SAFE only for %TEMP%-scoped operations

| Technique | Safe Realistic Implementation | Hard Limit |
|-----------|-------------------------------|------------|
| T1486 Data Encrypted for Impact | Encrypt a file created in %TEMP% only | NEVER touch user home directories or Documents. Cleanup must decrypt/delete immediately. |
| T1490 Inhibit Recovery | **AUDIT REQUIRED** — see note below | If current implementation calls `vssadmin delete shadows`, that is destructive and must be replaced |

**CRITICAL AUDIT: T1490 (Inhibit Recovery).** The real-world technique deletes VSS (Volume Shadow Copy) snapshots. If the current implementation runs `vssadmin delete shadows /all`, it destroys backups on the engagement host permanently. The safe implementation is read-only: `vssadmin list shadows` generates EID 4688 for vssadmin.exe, which is sufficient for SIEM detection testing. Deletion of shadows is an anti-feature for a SIEM validation tool deployed on client infrastructure.

---

## Feature Dependencies

```
Real /24 subnet scan (T1046 upgrade)
    └──requires──> Auto-detect primary adapter /24 (Get-NetIPConfiguration)

T1018 Remote System Discovery upgrade
    └──enhances──> Real /24 scan (uses discovered hosts for net view \\<host> recon)

Stub technique audit
    └──must precede──> All upgrade work (prevents duplicate effort)
    └──unblocks──> Expanded technique library

T1490 vssadmin audit
    └──P0 safety check (must happen before v1.3 ships)

T1070.001 log-clear audit
    └──P0 safety check (must happen before v1.3 ships)

T1546.003 WMI subscription cleanup audit
    └──P0 safety check (cleanup must be verified to work)

Cleanup-by-default enforcement
    └──required-by──> All persistence techniques
    └──required-by──> T1053.005, T1547.001, T1543.003, T1546.003, T1036.005
```

### Dependency Notes

- **T1018 upgrade pairs with T1046 upgrade:** Once the /24 scan discovers live hosts, T1018 can feed `net view \\<discovered-host>` for more authentic lateral discovery artifacts. Implement T1046 first.
- **Stub audit is the prerequisite gate:** No new technique work should start until existing techniques are classified. This prevents rewriting techniques that are already Tier 1 and missing techniques that are still Tier 3.
- **Safety audits are P0:** T1490 and T1070.001 could be destructive in their current form. These must be read and verified before the v1.3 milestone ships, not deferred.
- **WMI subscription cleanup is the highest-risk existing technique:** T1546.003 must be tested to confirm the cleanup command removes the subscription. A failed cleanup persists across reboots on the engagement host.

---

## MVP Definition for v1.3

### Launch With (v1.3 required)

The minimum set that delivers the realistic simulation promise.

- [ ] Safety audits: read T1490, T1070.001, T1546.003 — verify or fix destructive behavior
- [ ] Stub technique audit: classify all 57 techniques as Tier 1 / Tier 2 / Tier 3
- [ ] T1046 upgrade: real /24 subnet TCP scan with auto-detected subnet
- [ ] T1018 upgrade: add ICMP ping sweep + nltest DC discovery

### Add After Validation (v1.x)

- [ ] T1021.006 WinRM loopback simulation — test WinRM detection rules
- [ ] T1560.001 Archive staging upgrade — real Compress-Archive on %TEMP% files
- [ ] New collection techniques with real file enumeration patterns
- [ ] Kerberoasting explicit feedback: log clearly when no SPN accounts are found on non-domain-joined hosts

### Future Consideration (v2+)

- [ ] Campaign realistic chains: network scan feeds host list into credential access which feeds lateral movement simulation as a narrative sequence
- [ ] Technique parameterization via YAML: `subnet`, `scan_timeout_ms`, `port_list` configurable per engagement without recompiling
- [ ] Multi-host LDAP enumeration across discovered AD users (requires domain-joined host with DS access)

---

## Feature Prioritization Matrix

| Feature | Consultant Value | Implementation Cost | Priority |
|---------|-----------------|---------------------|----------|
| T1490 vssadmin safety audit | HIGH — prevents destroying client backups | LOW — single YAML read | P0 |
| T1070.001 log-clear safety audit | HIGH — prevents destroying Security log | LOW — single YAML read | P0 |
| T1546.003 cleanup verification | HIGH — prevents persistent WMI sub on client host | LOW — test cleanup command | P0 |
| Stub technique audit (catalogue) | HIGH — prerequisite to all upgrade work | LOW — read-only analysis | P1 |
| T1046 /24 subnet scan upgrade | HIGH — core v1.3 deliverable; T1046 YAML already has RunspaceFactory | LOW — extend existing subnet detection | P1 |
| T1018 ICMP/ARP/nltest upgrade | HIGH — pairs with T1046 for realistic Discovery phase | LOW — add Test-Connection + nltest | P1 |
| T1021.006 WinRM loopback | MEDIUM — tests common lateral movement gap | MEDIUM — conditional on WinRM state | P2 |
| New collection techniques | MEDIUM — adds coverage depth | MEDIUM — new YAML + PowerShell | P2 |
| Campaign realistic chains | HIGH long-term — shows attacker narrative | HIGH — requires engine changes | P3 |
| Technique parameterization | MEDIUM — per-engagement flexibility | HIGH — YAML schema change + engine | P3 |

---

## Competitor Feature Analysis

| Feature | Atomic Red Team | Caldera | Infection Monkey | LogNoJutsu |
|---------|-----------------|---------|-----------------|------------|
| MITRE ATT&CK coverage | 900+ atomics, full matrix | Full matrix, agent-based | Partial, network-focused | 57 curated techniques |
| Network scan target | Explicit IP required by caller | Agent discovers autonomously | Integrated autonomous spread | /24 upgrade needed for v1.3 |
| LDAP enumeration | T1087.002 DirectorySearcher | Via Caldera abilities | Limited | Full: DirectorySearcher, ldifde, dsquery |
| Single binary deployment | No — requires PS module install | No — server + agent | No — server required | Yes — core advantage |
| Offline operation | Partial | No | No | Yes — no internet at runtime |
| Cleanup commands | Yes — opt-in per invocation | Partial | No — relies on design reversibility | Yes — enforced for write operations |
| SIEM-specific coverage metadata | No | No | No | Yes — per-SIEM YAML field |
| Windows Event ID declarations | No | No | No | Yes — expected_events per technique |

---

## Sources

- Atomic Red Team GitHub (T1046, T1018, T1003, T1558 atomics): https://github.com/redcanaryco/atomic-red-team
- MITRE ATT&CK Enterprise Matrix: https://attack.mitre.org/techniques/enterprise/
- Caldera documentation: https://caldera.readthedocs.io/
- Infection Monkey BAS design: https://www.akamai.com/infectionmonkey
- Invoke-AtomicRedTeam cleanup documentation: https://www.atomicredteam.io/invoke-atomicredteam/docs/cleanup
- EVTX to MITRE ATT&CK event mapping: https://github.com/mdecrevoisier/EVTX-to-MITRE-Attack
- LogNoJutsu existing technique YAML files: direct codebase read (HIGH confidence)
- LogNoJutsu executor.go: direct codebase read (HIGH confidence)

---

*Feature research for: LogNoJutsu v1.3 Realistic Attack Simulation*
*Researched: 2026-04-09*
