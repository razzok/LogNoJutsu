# Technique Classification

## Tier Definitions

- **Tier 1 (Realistic):** Generates the real Windows events a SIEM would detect in an actual attack. Uses real attacker tools/APIs with realistic parameters.
- **Tier 2 (Partial):** Generates some real events but uses simulation shortcuts (simulated credentials, localhost targets, custom channels, fake data).
- **Tier 3 (Stub):** Echo/stub/query-only that proves the technique runs but does not generate realistic SIEM events (pure enumeration, trivial commands).

## Verification Statuses

After execution, each technique receives a verification status indicating the outcome of post-execution event log verification:

| Status | Meaning |
|--------|---------|
| `pass` | Expected Windows events were found in the event log |
| `fail` | Expected events were NOT found in the event log |
| `not_run` | Verification was not attempted |
| `not_executed` | The technique command did not execute |
| `amsi_blocked` | AMSI (Antimalware Scan Interface) blocked the PowerShell execution before it could run |
| `elevation_required` | The technique requires administrator privileges but the current session is not elevated |

## Executor Types

- **powershell** -- Technique command is executed via PowerShell (the default for most techniques).
- **go** -- Technique logic is implemented as a native Go function compiled into the LogNoJutsu binary. The YAML `command` field is empty; execution is handled entirely in Go code.

## Confirmation Flag

Techniques with `requires_confirmation: true` prompt the operator for explicit approval before executing, typically because they perform destructive or high-impact actions. **No techniques currently have this flag set.**

## Technique Table

> **Elev** = `elevation_required` in the YAML (technique needs administrator privileges).
> **Exec** = Executor type (`ps` = powershell, `go` = native Go).

| Technique ID | Name | Tier | Exec | Elev | Has Cleanup | Rationale |
|---|---|:---:|:---:|:---:|:---:|---|
| AZURE_dcsync | Sentinel - Potential DCSync Replication Access | 2 | ps | No | No | LDAP ACL queries trigger EID 4662 but no actual replication tool (mimikatz/impacket) used |
| AZURE_kerberoasting | Sentinel - Potential Kerberoasting Detection | 1 | ps | No | No | Real TGS ticket requests via System.IdentityModel trigger real EID 4769 per SPN on DC |
| AZURE_ldap_recon | Sentinel - LDAP Directory Reconnaissance | 2 | ps | No | Yes | Real LDAP queries via DirectorySearcher but read-only recon, no attack artifact |
| FALCON_lateral_movement_psexec | Falcon Sensor - Lateral Movement PsExec Detection | 1 | ps | No | Yes | Real sc.exe service create/start/delete triggers EID 7045 + EID 4697; WMI process creation |
| FALCON_lsass_access | Falcon Sensor - LSASS Credential Theft Detection | 1 | ps | Yes | Yes | Real MiniDumpWriteDump via dbghelp.dll; PowerShell opening LSASS is authentic Falcon trigger |
| FALCON_process_injection | Falcon Sensor - Process Injection Detection | 1 | ps | Yes | Yes | Real OpenProcess + VirtualAllocEx + WriteProcessMemory + CreateRemoteThread API calls |
| T1003.001 | LSASS Memory Access -- Credential Dumping Simulation | 1 | ps | Yes | Yes | Real rundll32 + comsvcs.dll MiniDump -- exact LOLBin pattern used by real threat actors |
| T1003.006 | DCSync -- Domain Credential Replication Simulation | 2 | ps | No | No | Real nltest generates EID 4688 but no actual DCSync replication; LDAP queries only |
| T1005 | Data from Local System | 2 | ps | No | Yes | Real file enumeration + staging to temp dir but no actual C2 transfer; uses simulation staging path |
| T1016 | System Network Configuration Discovery | 3 | ps | No | No | Pure enumeration: ipconfig/all, netstat, net view -- trivial discovery commands |
| T1021.001 | Remote Desktop Protocol -- Lateral Movement Simulation | 2 | ps | Yes | Yes | cmdkey stores credential + registry/firewall changes but no actual RDP session established |
| T1021.002 | SMB/Windows Admin Shares | 3 | ps | No | Yes | net share + WMI share enumeration only; no actual admin share access or lateral movement |
| T1027 | Obfuscated Files or Information | 1 | ps | No | Yes | Real powershell.exe -EncodedCommand with full attacker flag set (-NoP -NonI -W Hidden -Exec Bypass) |
| T1036.005 | Match Legitimate Name or Location (Process Masquerading) | 1 | ps | No | Yes | Copies svchost.exe to TEMP and runs it; Sysmon EID 1 captures wrong image path |
| T1041 | Exfiltration Over C2 Channel (HTTP) | 2 | ps | No | Yes | Real Invoke-WebRequest HTTP POST but to .invalid TLD C2; connection fails but Sysmon EID 3 fires |
| T1046 | Network Service Discovery (Port Scan) | 3 | ps | No | No | PowerShell TCP connect to localhost ports only; no network scan; no Sysmon EID 3 from external scan |
| T1047 | WMI Remote Execution | 1 | ps | No | Yes | Real wmic.exe process call create; WmiPrvSE.exe spawns cmd.exe -- authentic WMI lateral movement pattern |
| T1049 | System Network Connections Discovery | 3 | ps | No | No | netstat -ano; pure network connection enumeration command |
| T1053.005 | Scheduled Task Persistence | 1 | ps | No | Yes | Real schtasks.exe /create with encoded PS payload + Register-ScheduledTask -- actual task written to Task Scheduler |
| T1057 | Process Discovery | 1 | **go** | No | No | Full attacker process enumeration via tasklist /v /svc, wmic process get CommandLine, Get-WmiObject Win32_Process with parent-child tree targeting lsass/csrss/winlogon |
| T1059.001 | PowerShell Execution | 1 | ps | No | No | Real powershell.exe invocation with full attacker flag set; no persistent artifacts from invocation itself |
| T1059.003 | Windows Command Shell (cmd.exe) | 1 | ps | No | Yes | Real cmd.exe chained recon commands with output redirection -- authentic attacker cmd.exe pattern |
| T1069 | Permission Groups Discovery | 1 | ps | No | No | Real net.exe 'localgroup Administrators' generates EID 4688 with attacker command line; whoami /groups, wmic group, .NET WindowsIdentity — no simulation shortcuts |
| T1070.001 | Clear Windows Event Logs | 2 | ps | Yes | Yes | Creates custom LogNoJutsu-Test channel + wevtutil cl; generates real EID 104 but not clearing real Security/System logs |
| T1071.001 | Application Layer Protocol -- Web Protocols | 2 | ps | No | No | Real Invoke-WebRequest to .invalid C2 host; Sysmon EID 3 + EID 22 fire but connection fails by design |
| T1071.004 | Application Layer Protocol -- DNS | 2 | ps | No | No | Real nslookup + DNS subdomain loop to .invalid domain; Sysmon EID 22 fires but no real C2 channel |
| T1082 | System Information Discovery | 1 | ps | No | No | Full attacker recon burst: systeminfo, wmic /format:csv, reg query, hostname, whoami — temporal clustering is Exabeam behavioral trigger; all real tools with real parameters |
| T1083 | File and Directory Discovery | 1 | ps | No | No | Real cmd.exe dir /s /b on user profile paths + tree /F + Get-ChildItem recursive on sensitive locations + ADS detection — genuine attacker file enumeration artifacts |
| T1087 | Account Discovery | 3 | ps | No | No | net user + net localgroup + ADSI queries; account enumeration only |
| T1098 | Account Manipulation | 1 | ps | Yes | Yes | Real net user /add + net localgroup /add + password changes trigger EID 4720, EID 4732 |
| T1110.001 | Password Brute Force | 1 | ps | No | No | Real Win32 LogonUser API calls generating EID 4625 bursts with correct SubStatus codes |
| T1110.003 | Password Spraying | 1 | ps | No | No | Real GetLocalUser + LogonUser spray pattern generating EID 4625 across multiple accounts |
| T1119 | Automated Collection | 3 | ps | No | Yes | Automated Get-ChildItem loop with file metadata collection; no actual data staging to C2 |
| T1134.001 | Token Impersonation/Theft | 3 | ps | No | Yes | whoami /priv + SeDebugPrivilege check; privilege enumeration only, no actual token theft |
| T1135 | Network Share Discovery | 2 | ps | No | No | Real net share/view + Get-SmbShare enumeration but admin share access (dir \\COMPUTERNAME\C$) is loopback SMB — simulation shortcut; EID 5140/5145 require Object Access audit policy |
| T1136.001 | Create Local Account | 1 | ps | Yes | Yes | Real net user /add + net localgroup /add trigger EID 4720 (account created) + EID 4732 (group add) |
| T1197 | BITS Jobs Persistence | 1 | ps | No | Yes | Real bitsadmin /setnotifycmdline with Windows-masquerading job name; EID 59 fires on completion |
| T1218.011 | Rundll32 Proxy Execution | 1 | ps | No | Yes | Real rundll32.exe LOLBin patterns (pcwutl.dll, javascript: URI, shell32.dll); Sysmon EID 1 + process chains |
| T1482 | Domain Trust Discovery | 1 | **go** | No | No | Real nltest /domain_trusts -- verbatim command from Ryuk/Trickbot/Cobalt Strike; standalone SIEM rule |
| T1486 | Data Encrypted for Impact (Ransomware Simulation) | 2 | ps | No | Yes | Real AES-256 encryption with ransom note + delete but only operates on simulation temp directory |
| T1490 | Inhibit System Recovery | 1 | ps | Yes | Yes | Real bcdedit /set recoveryenabled No + registry recovery policy changes; fully reversible by cleanup |
| T1543.003 | Create Windows Service | 1 | ps | Yes | Yes | Real sc.exe create + New-Service + registry; EID 7045 fires; service masquerades as Windows service |
| T1546.003 | WMI Event Subscription Persistence | 1 | ps | Yes | Yes | Real WMI EventFilter + CommandLineEventConsumer + FilterToConsumerBinding -- actual WMI persistence objects |
| T1547.001 | Registry Run Key Persistence | 1 | ps | No | Yes | Real reg add to HKCU\Run with hidden PS payload; Sysmon EID 13 fires on registry write |
| T1548.002 | UAC Bypass (Abuse Elevation Control Mechanism) | 1 | ps | No | Yes | Real fodhelper + eventvwr + sdclt UAC bypass techniques using registry hijacking |
| T1550.002 | Pass the Hash -- Credential Reuse Simulation | 2 | ps | No | No | Real net use IPC$ + cmdkey explicit credential logon triggers EID 4648/4624 but no actual NTLM hash |
| T1552.001 | Credentials in Files | 2 | ps | No | Yes | Creates fake credential files then searches for them; simulates attacker hunting but with fake data |
| T1558.003 | Kerberoasting -- Service Ticket Request | 1 | ps | No | No | Real setspn.exe enumeration + System.IdentityModel TGS requests; EID 4769 fires per ticket request |
| T1560.001 | Archive Collected Data -- Archive via Utility | 2 | ps | No | Yes | Real Compress-Archive on staging directory but only operates on simulation data, not real sensitive files |
| T1562.002 | Disable Windows Event Logging | 1 | ps | Yes | Yes | Real auditpol commands disable actual audit subcategories; EID 4719 fires; cleanup restores policy |
| T1574.002 | DLL Side-Loading | 2 | ps | No | Yes | Real Add-Type DLL compilation + LoadLibrary into process but benign DLL in simulation directory |
| UEBA-ACCOUNT-TAKEOVER | UEBA -- Account Takeover Chain | 2 | ps | No | No | Real LogonUser API failures + enumeration burst but simulated chain, not actual credential theft |
| UEBA-SPRAY-CHAIN | UEBA -- Credential Spray then Success Chain | 2 | ps | No | No | Real LogonUser x25 failures + success triggers Exabeam spray pattern; but controlled simulation |
| UEBA-DATA-STAGING | UEBA -- Data Staging + Exfiltration Chain | 2 | ps | No | Yes | Real file staging + HTTP POST attempt; .invalid C2 so connection fails but Sysmon EID 3 fires |
| UEBA-LATERAL-CHAIN | UEBA -- Lateral Movement Discovery Chain | 3 | ps | No | No | Rapid net user/ipconfig/netstat/arp burst; pure enumeration chain, no lateral movement |
| UEBA-LATERAL-NEW-ASSET | UEBA -- Lateral Movement + New Asset Access | 2 | ps | No | Yes | Real net use SMB + Sysmon EID 3 on port 445; localhost/loopback target, not real lateral movement |
| UEBA-OFFHOURS | UEBA -- Off-Hours Activity Simulation | 3 | ps | No | No | whoami + net user + Get-Process off-hours; trivial enumeration at current system time |
| UEBA-PRIV-ESC | UEBA -- Privilege Escalation Chain | 3 | ps | No | No | whoami /priv + net localgroup + WindowsIdentity .NET check; pure privilege enumeration, no escalation |

## Summary Statistics

| Category | Count |
|---|:---:|
| Total techniques | 58 |
| Tier 1 (Realistic) | 29 |
| Tier 2 (Partial) | 19 |
| Tier 3 (Stub) | 10 |
| Native Go executor | 2 |
| Elevation required | 11 |
| Has cleanup command | 33 |
| Requires confirmation | 0 |
