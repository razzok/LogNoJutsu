# LogNoJutsu — SIEM Validation & ATT&CK Simulation Tool

## Table of Contents

1. [What is LogNoJutsu?](#what-is-lognojutsu)
2. [Origin and Motivation](#origin-and-motivation)
3. [How It Works](#how-it-works)
4. [System Requirements](#system-requirements)
5. [Quick Start](#quick-start)
6. [Web Interface](#web-interface)
7. [System Preparation](#system-preparation)
8. [Simulation Phases](#simulation-phases)
9. [Multi-User Simulation](#multi-user-simulation)
10. [Exabeam Use Case Coverage](#exabeam-use-case-coverage)
11. [Techniques — Complete Reference](#techniques--complete-reference)
    - [Phase 1: Discovery (Enumeration)](#phase-1-discovery-enumeration)
    - [Phase 2: Attack](#phase-2-attack)
    - [UEBA Scenarios (Exabeam)](#ueba-scenarios-exabeam)
12. [Campaigns / Playbooks](#campaigns--playbooks)
13. [Logging and Reporting](#logging-and-reporting)
14. [Cleanup Mechanism](#cleanup-mechanism)
15. [Command-Line Options](#command-line-options)

---

## What is LogNoJutsu?

LogNoJutsu is a **SIEM validation tool** that simulates real attacker behavior on a Windows system. Its goal is to automatically verify, after onboarding a SIEM solution, whether all relevant detections and use cases fire correctly.

The tool does **not perform real attacks** and does not extract credentials or sensitive data. Instead, it generates the actions and system artifacts that a real attacker would leave behind — specifically: the Windows Event Logs, Sysmon events, and PowerShell logs by which a SIEM should detect an attack.

**Core principle:** If LogNoJutsu executes a technique and the SIEM does *not* detect it, there is a problem in the SIEM onboarding, log forwarding, or rule configuration.

---

## Origin and Motivation

LogNoJutsu was developed as an open-source project inspired by **Magneto**, an internal tool by SIEM vendor Exabeam. Magneto is presented at Exabeam events and is available exclusively internally. It was written in PowerShell, launched via an executable, offered a simple web interface, and executed small anomaly and enumeration actions at configurable time intervals before running full attack simulations.

LogNoJutsu implements this idea as a standalone, extensible tool — with a particular focus on **Exabeam UEBA use cases**, but designed so that standard ATT&CK techniques can also be detected with other SIEM solutions (Splunk, Microsoft Sentinel, IBM QRadar, etc.).

---

## How It Works

```
┌─────────────────────────────────────────────────────────┐
│  lognojutsu.exe                                         │
│                                                         │
│  1. Starts HTTP server (localhost:8080)                 │
│  2. Opens Web UI in browser                            │
│  3. Waits for user interaction — NOTHING runs           │
│     automatically                                       │
│                                                         │
│  After clicking "Start Simulation":                     │
│                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │ PREP PHASE   │ →  │ PHASE 1      │ →  │ PHASE 2  │  │
│  │ (manual)     │    │ Discovery    │    │ Attack   │  │
│  │              │    │ (after T1)   │    │ (after T2)│  │
│  └──────────────┘    └──────────────┘    └──────────┘  │
│                                                         │
│  Every action is logged to a .log file                 │
│  At the end: cleanup of all created artifacts           │
└─────────────────────────────────────────────────────────┘
```

The tool is a **single `.exe` file** (~10 MB). No additional runtime environments, external software, or internet connection are required (except for the automatic Sysmon download during preparation).

---

## System Requirements

| Requirement | Details |
|---|---|
| Operating System | Windows 10 / Windows 11 / Windows Server 2016+ |
| Permissions | Regular user for Discovery techniques; Administrator for Attack techniques and Preparation |
| PowerShell | Version 5.1+ (available on all modern Windows versions) |
| Network | Only required for automatic Sysmon download during the Preparation phase |
| SIEM Agent | A log forwarder (Exabeam Agent, Winlogbeat, NXLog, etc.) should be configured and active before starting |

---

## Quick Start

```
# Standard — locally accessible only:
lognojutsu.exe

# With network access (SIEM engineer configures from own laptop):
lognojutsu.exe -host 0.0.0.0 -port 8080

# With password protection:
lognojutsu.exe -host 0.0.0.0 -port 8080 -password "MyPassword123"
```

After starting, open `http://localhost:8080` in your browser. **No simulation is running yet.**

---

## Web Interface

The web UI consists of seven sections:

| Tab | Function |
|---|---|
| **Dashboard** | Current simulation status, phase display, execution timeline, expected SIEM events |
| **Preparation** | One-time system preparation: Audit Policy, PowerShell Logging, Sysmon installation |
| **Playbooks** | Overview of all available campaigns and individual techniques |
| **Configure & Run** | Choose mode (Quick / PoC), campaign, timing, tactic filter, WhatIf mode, user rotation, start/stop simulation |
| **Results** | Detailed results for each executed technique with output, cleanup status, and executing user |
| **Simulation Log** | Live log stream of all actions with color-coded view by event type |
| **Users** | Manage user profiles (local/AD), store credentials, discovery, credential test |

---

## System Preparation

Before the first simulation, three configuration steps must be performed once. These require **administrator rights** and are initiated via the web UI.

### 1. PowerShell ScriptBlock Logging

**What is configured:**
Sets three registry keys under `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell`:

| Key | Value | Purpose |
|---|---|---|
| `ScriptBlockLogging\EnableScriptBlockLogging` | `1` | Enable Event 4104 |
| `ModuleLogging\EnableModuleLogging` | `1` | Enable Event 4103 |
| `Transcription\EnableTranscripting` | `1` | Enable PowerShell transcription |

**Why important:** Without ScriptBlock Logging (4104), PowerShell-based attacks are practically invisible in the SIEM. This is the most important setting for detecting T1059.001 (79 Exabeam rules), T1027 (47 rules), and all other PowerShell-heavy techniques.

### 2. Windows Audit Policy

**What is configured:**
12 audit subcategories are enabled via `auditpol.exe` and command-line logging in Event 4688 is enabled via a registry key:

| Subcategory | Generated Events |
|---|---|
| Logon | 4624 (Success), 4625 (Failure), 4634 (Logoff) |
| Account Lockout | 4740 |
| Logon — Special Logon | 4672 |
| Process Creation | 4688 (incl. command line) |
| Scheduled Task | 4698, 4699 |
| Security Group Management | 4728, 4732 |
| User Account Management | 4720, 4726 |
| Audit Policy Change | 4719 |
| Directory Service Access | 4662 |
| Sensitive Privilege Use | 4673, 4674 |
| Other Object Access | 4698 |
| Kerberos Authentication | 4768, 4769, 4771, 4776 |

**Why important:** Event 4688 with command-line logging is a prerequisite for detecting LOLBin techniques (T1218, T1059.003). Events 4624/4625/4648 are core prerequisites for all credential and lateral movement use cases.

### 3. Sysmon Installation

Sysmon (System Monitor) from Sysinternals is automatically downloaded and installed with an optimized configuration. The following Event IDs are configured:

| Sysmon Event ID | Description | Important for |
|---|---|---|
| **1** | Process Create (with hashes, parent, cmdline) | Almost all techniques |
| **3** | Network Connection | T1046 (Port Scan), T1135 (Share Discovery) |
| **7** | Image Loaded (DLL) | T1218.011 (Rundll32) |
| **8** | CreateRemoteThread | Injection techniques |
| **10** | ProcessAccess | T1003.001 (LSASS access) |
| **11** | FileCreate | T1486 (Ransomware), T1036.005 (Masquerading) |
| **12/13** | RegistryEvent | T1547.001, T1548.002 (UAC Bypass) |
| **19/20/21** | WMI Event | T1546.003 (WMI Subscription) |
| **22** | DNS Query | C2 communication, T1197 (BITS) |

---

## Simulation Phases

### Simulation Modes

LogNoJutsu offers two operating modes:

**Quick Mode** — One-time simulation within minutes/hours:
- T1: Wait time before Phase 1 (0 – 7200s)
- T2: Pause between Phase 1 and Phase 2 (0 – 7200s)
- Ideal for quick SIEM tests and technical validation

**PoC Multi-Day Mode** — Multi-day simulation for 4-week PoC with Exabeam UEBA:
- Phase 1 runs N days (e.g., 7–14), 2–5 Discovery techniques daily for UEBA baseline building
- Followed by a pause (gap, configurable in days)
- Phase 2 runs N days (e.g., 7–14), one complete attack campaign per day
- Execution occurs daily at the configured time (e.g., 09:00)
- Dashboard shows countdown to next execution

### Phase 1 — Discovery ("Low & Slow")

Starts after the configured wait time T1 (default: 5 seconds in test, in production e.g., 10–30 minutes). Executes all **10 enumeration techniques** (filtered by active tactic filter). Goal: disrupt UEBA baseline and generate recon behavior for anomaly detections.

### Phase 2 — Attack ("Full Attack")

Starts after Phase 1 completes + wait time T2 (default: 30 seconds in test). Executes the selected campaign or all **33 attack techniques** (filtered by active tactic filter). This is where the serious artifacts are created (Persistence, Credential Access, Defense Evasion, Exfiltration).

### Cleanup

After the simulation ends (or on manual abort), all created artifacts are automatically removed. Details: see [Cleanup Mechanism](#cleanup-mechanism).

---

## Multi-User Simulation

LogNoJutsu can execute techniques in the security context of **other users** — both local Windows accounts and Active Directory domain users. This is particularly valuable for Exabeam UEBA validation, as Exabeam builds behavioral baselines per user and detects anomalies when a user performs unusual actions.

### Concept: User Profiles

A **user profile** in LogNoJutsu consists of:

| Field | Description |
|---|---|
| **Username** | Windows username (without domain) |
| **Domain** | Domain name for AD users; empty for local accounts |
| **Password** | Password — stored encrypted (DPAPI) |
| **User Type** | `local` (local account), `domain` (AD account), `current` (current user) |
| **Display Name** | Optional display name for the UI |

Profiles are stored in `lognojutsu_users.json` in the working directory. The file is protected with file system permissions `0600`.

### Password Security: Windows DPAPI

Passwords are **never stored in plaintext**. Instead, LogNoJutsu uses the **Windows Data Protection API (DPAPI)**:

```
Store:  PowerShell ConvertFrom-SecureString (without key = DPAPI Machine+User context)
Read:   PowerShell ConvertTo-SecureString + SecureStringToBSTR
```

DPAPI binds the encryption to the Windows user and machine. This means: the `lognojutsu_users.json` cannot be decrypted on a different system or with a different user.

> **Fallback:** If DPAPI is not available (e.g., in sandbox environments), LogNoJutsu stores the password with the prefix `PLAIN:`.

### Execution Mechanism: ProcessStartInfo with Credentials

When a technique is to be executed as a different user, LogNoJutsu uses `System.Diagnostics.ProcessStartInfo` with explicit credentials:

```powershell
$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = "powershell.exe"
$psi.UserName = "DOMAIN\username"    # or ".\username" for local accounts
$psi.Password = $securePassword
$psi.UseShellExecute = $false
$psi.RedirectStandardOutput = $true
$proc = [System.Diagnostics.Process]::Start($psi)
```

**Generated Windows Events:**

| Event ID | Log | Description |
|---|---|---|
| **4648** | Security | *A logon was attempted using explicit credentials* — core indicator for RunAs behavior, central Exabeam UEBA signal |
| **4624** | Security | Successful logon of the target user |
| **4688** | Security | Process creation with command-line logging |
| **1** | Sysmon | Process Create with ParentProcess = LogNoJutsu |

### User Rotation

| Mode | Behavior |
|---|---|
| **None** | All techniques run as the current user (no profiles needed) |
| **Sequential** | Profiles are assigned in order (Technique 1 → User A, Technique 2 → User B, …) |
| **Random** | Each technique randomly receives one of the configured profiles |

### Complete Workflow

```
1. Tab "Users" → Add users (or use Discovery)
2. Perform credential test (green checkmark = OK)
3. Tab "Configure & Run"
4. Choose User Rotation Mode: Sequential or Random
5. Select desired profiles in the multi-select (Ctrl+Click)
6. Start simulation
```

---

## Execution Options

### WhatIf Mode (Preview)

The **WhatIf mode** does not execute techniques, but only shows what would be executed. Ideal for:
- Planning before a simulation
- Demos without artifacts
- Verifying the tactic filter

Activation: Checkbox "WhatIf mode" in Configure & Run. The simulation log shows `[WhatIf] Would run: T1082 — ...` for each technique. At the end, JSON and HTML reports are generated (marked with WhatIf badge).

### Pause Between Techniques

Configurable wait time (seconds) after each executed technique. Useful for:
- Realistic operational tempo (attacker pacing)
- Maintaining SIEM correlation windows
- Avoiding overly dense event bursts

Recommendation: 5–30 seconds for real PoC simulations.

### Tactic Filter

Checkboxes for all 10 ATT&CK tactics allow targeted inclusion/exclusion:

| Use Case | Filter Configuration |
|---|---|
| Discovery events only | Deselect all, only `discovery` active |
| No destructive techniques | Deselect `impact` |
| Credential tests only | `credential-access` + `privilege-escalation` active |
| Exfiltration validation | `exfiltration` + `collection` active |

The filter applies in both modes (Quick + PoC Multi-Day).

---

## Exabeam Use Case Coverage

LogNoJutsu covers all three Exabeam TDIR Use Case Packages with a total of 21 use cases. The following table shows the mapping:

| Exabeam Use Case Package | Use Case | Covering Techniques |
|---|---|---|
| **Compromised Insiders** | Compromised Credentials | T1110.001, T1110.003, UEBA-SPRAY-CHAIN |
| | Lateral Movement | T1021.001, T1550.002, T1558.003, T1046, T1135, T1482 |
| | Privilege Escalation | T1548.002, T1098, T1136.001, T1087, T1069 |
| | Privileged Activity | T1003.001, T1003.006, T1059.001 |
| | Account Manipulation | T1136.001, T1098 |
| | Data Exfiltration | T1552.001, T1083 |
| | Evasion | T1027, T1036.005, T1218.011, T1562.002, T1070.001, T1548.002 |
| **Malicious Insiders** | Audit Tampering | T1562.002, T1070.001 |
| | Data Leak | T1552.001, T1083 |
| | Privilege Abuse | T1098, T1003.001 |
| | Abnormal Auth & Access | UEBA-OFFHOURS, UEBA-LATERAL-CHAIN |
| **External Threats** | Ransomware | T1490, T1486, T1059.001, T1562.002, T1070.001 |
| | Malware | T1547.001, T1053.005, T1543.003, T1197, T1546.003 |
| | Brute Force | T1110.001, T1110.003, UEBA-SPRAY-CHAIN |

**Exabeam rule coverage by technique (from content documentation):**

| Technique | Exabeam Rules | Priority |
|---|---|---|
| T1078 Valid Accounts | 304 | via Multi-User + 4648 |
| T1059 Command & Scripting | 144 | T1059.001 + T1059.003 |
| T1021 Remote Services | 118 | T1021.001 |
| T1218 System Binary Proxy | 116 | T1218.011 (Rundll32) |
| T1098 Account Manipulation | 57 | T1098 |
| T1048 Exfiltration Alt Protocol | 68 | T1552.001, T1083 |
| T1027 Obfuscated Files | 47 | T1027 |
| T1003 Credential Dumping | 49 | T1003.001, T1003.006 |
| T1558 Kerberos Tickets | 36 | T1558.003 |
| T1136 Create Account | 35 | T1136.001 |
| T1550 Alternate Auth Material | 38 | T1550.002 |
| T1083 File Discovery | 38 | T1083 |
| T1543 System Process | 38 | T1543.003 |
| T1036 Masquerading | 27 | T1036.005 |
| T1053 Scheduled Task | 27 | T1053.005 |
| T1087 Account Discovery | 25 | T1087 |
| T1550.002 Pass the Hash | 23 | T1550.002 |
| T1558.003 Kerberoasting | 22 | T1558.003 |
| T1047 WMI | 18 | T1047 |
| T1562 Impair Defenses | 18 | T1562.002 |
| T1070 Indicator Removal | 18 | T1070.001 |
| T1135 Network Share Discovery | 12 | T1135 |
| T1197 BITS Jobs | 6 | T1197 |
| T1546.003 WMI Subscription | 6 | T1546.003 |
| T1110.003 Password Spraying | **1** | T1110.003 (Gap test) |

> **Gap validation:** Techniques with few Exabeam rules (T1110.003 = 1 rule, T1197 = 6, T1546.003 = 6) are intentionally included in the tool to make gaps in SIEM configuration visible.

---
