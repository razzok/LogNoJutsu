# Technique Classification

Tier definitions:
- **Tier 1 (Realistic):** Generates the real Windows events a SIEM would detect in an actual attack. Uses real attacker tools/APIs with realistic parameters.
- **Tier 2 (Partial):** Generates some real events but uses simulation shortcuts (simulated credentials, localhost targets, custom channels, fake data).
- **Tier 3 (Stub):** Echo/stub/query-only that proves the technique runs but does not generate realistic SIEM events (pure enumeration, trivial commands).

| Technique ID | Name | Tier | Rationale | Has Cleanup | Writes Artifacts |
|-------------|------|------|-----------|-------------|-----------------|
| AZURE_dcsync | Sentinel - Potential DCSync Replication Access | 2 | LDAP ACL queries trigger EID 4662 but no actual replication tool (mimikatz/impacket) used | No | No |
| AZURE_kerberoasting | Sentinel - Potential Kerberoasting Detection | 1 | Real TGS ticket requests via System.IdentityModel trigger real EID 4769 per SPN on DC | No | No |
| AZURE_ldap_recon | Sentinel - LDAP Directory Reconnaissance | 2 | Real LDAP queries via DirectorySearcher but read-only recon, no attack artifact | Yes | Yes |
| FALCON_lateral_movement_psexec | Falcon Sensor - Lateral Movement PsExec Detection | 1 | Real sc.exe service create/start/delete triggers EID 7045 + EID 4697; WMI process creation | Yes | Yes |
| FALCON_lsass_access | Falcon Sensor - LSASS Credential Theft Detection | 1 | Real MiniDumpWriteDump via dbghelp.dll; PowerShell opening LSASS is authentic Falcon trigger | Yes | Yes |
| FALCON_process_injection | Falcon Sensor - Process Injection Detection | 1 | Real OpenProcess + VirtualAllocEx + WriteProcessMemory + CreateRemoteThread API calls | Yes | Yes |
| T1003.001 | LSASS Memory Access — Credential Dumping Simulation | 1 | Real rundll32 + comsvcs.dll MiniDump — exact LOLBin pattern used by real threat actors | Yes | Yes |
| T1003.006 | DCSync — Domain Credential Replication Simulation | 2 | Real nltest generates EID 4688 but no actual DCSync replication; LDAP queries only | No | No |
| T1005 | Data from Local System | 2 | Real file enumeration + staging to temp dir but no actual C2 transfer; uses simulation staging path | Yes | Yes |
| T1016 | System Network Configuration Discovery | 3 | Pure enumeration: ipconfig/all, netstat, net view — trivial discovery commands | No | No |
| T1021.001 | Remote Desktop Protocol — Lateral Movement Simulation | 2 | cmdkey stores credential + registry/firewall changes but no actual RDP session established | Yes | Yes |
| T1021.002 | SMB/Windows Admin Shares | 3 | net share + WMI share enumeration only; no actual admin share access or lateral movement | Yes | Yes |
| T1027 | Obfuscated Files or Information | 1 | Real powershell.exe -EncodedCommand with full attacker flag set (-NoP -NonI -W Hidden -Exec Bypass) | Yes | Yes |
| T1036.005 | Match Legitimate Name or Location (Process Masquerading) | 1 | Copies svchost.exe to TEMP and runs it; Sysmon EID 1 captures wrong image path | Yes | Yes |
| T1041 | Exfiltration Over C2 Channel (HTTP) | 2 | Real Invoke-WebRequest HTTP POST but to .invalid TLD C2; connection fails but Sysmon EID 3 fires | Yes | Yes |
| T1046 | Network Service Discovery (Port Scan) | 3 | PowerShell TCP connect to localhost ports only; no network scan; no Sysmon EID 3 from external scan | No | No |
| T1047 | WMI Remote Execution | 1 | Real wmic.exe process call create; WmiPrvSE.exe spawns cmd.exe — authentic WMI lateral movement pattern | Yes | Yes |
| T1049 | System Network Connections Discovery | 3 | netstat -ano; pure network connection enumeration command | No | No |
| T1053.005 | Scheduled Task Persistence | 1 | Real schtasks.exe /create with encoded PS payload + Register-ScheduledTask — actual task written to Task Scheduler | Yes | Yes |
| T1057 | Process Discovery | 3 | tasklist /v /svc + Get-Process — process enumeration only; not attacker-realistic | No | No |
| T1059.001 | PowerShell Execution | 1 | Real powershell.exe invocation with full attacker flag set; no persistent artifacts from invocation itself | No | No |
| T1059.003 | Windows Command Shell (cmd.exe) | 1 | Real cmd.exe chained recon commands with output redirection — authentic attacker cmd.exe pattern | Yes | Yes |
| T1069 | Permission Groups Discovery | 3 | net localgroup enumeration only; no exploitation | No | No |
| T1070.001 | Clear Windows Event Logs | 2 | Creates custom LogNoJutsu-Test channel + wevtutil cl; generates real EID 104 but not clearing real Security/System logs | Yes | Yes |
| T1071.001 | Application Layer Protocol — Web Protocols | 2 | Real Invoke-WebRequest to .invalid C2 host; Sysmon EID 3 + EID 22 fire but connection fails by design | No | No |
| T1071.004 | Application Layer Protocol — DNS | 2 | Real nslookup + DNS subdomain loop to .invalid domain; Sysmon EID 22 fires but no real C2 channel | No | No |
| T1082 | System Information Discovery | 3 | systeminfo + hostname + whoami /all — trivial system info discovery | No | No |
| T1083 | File and Directory Discovery | 3 | Get-ChildItem recursive enumeration; file discovery only | No | No |
| T1087 | Account Discovery | 3 | net user + net localgroup + ADSI queries; account enumeration only | No | No |
| T1098 | Account Manipulation | 1 | Real net user /add + net localgroup /add + password changes trigger EID 4720, EID 4732 | Yes | Yes |
| T1110.001 | Password Brute Force | 1 | Real Win32 LogonUser API calls generating EID 4625 bursts with correct SubStatus codes | No | No |
| T1110.003 | Password Spraying | 1 | Real GetLocalUser + LogonUser spray pattern generating EID 4625 across multiple accounts | No | No |
| T1119 | Automated Collection | 3 | Automated Get-ChildItem loop with file metadata collection; no actual data staging to C2 | Yes | Yes |
| T1134.001 | Token Impersonation/Theft | 3 | whoami /priv + SeDebugPrivilege check; privilege enumeration only, no actual token theft | Yes | Yes |
| T1135 | Network Share Discovery | 3 | net share + Get-SmbShare + WMI Win32_Share enumeration; discovery only | No | No |
| T1136.001 | Create Local Account | 1 | Real net user /add + net localgroup /add trigger EID 4720 (account created) + EID 4732 (group add) | Yes | Yes |
| T1197 | BITS Jobs Persistence | 1 | Real bitsadmin /setnotifycmdline with Windows-masquerading job name; EID 59 fires on completion | Yes | Yes |
| T1218.011 | Rundll32 Proxy Execution | 1 | Real rundll32.exe LOLBin patterns (pcwutl.dll, javascript: URI, shell32.dll); Sysmon EID 1 + process chains | Yes | Yes |
| T1482 | Domain Trust Discovery | 1 | Real nltest /domain_trusts — verbatim command from Ryuk/Trickbot/Cobalt Strike; standalone SIEM rule | No | No |
| T1486 | Data Encrypted for Impact (Ransomware Simulation) | 2 | Real AES-256 encryption with ransom note + delete but only operates on simulation temp directory | Yes | Yes |
| T1490 | Inhibit System Recovery | 1 | Real bcdedit /set recoveryenabled No + registry recovery policy changes; fully reversible by cleanup | Yes | Yes |
| T1543.003 | Create Windows Service | 1 | Real sc.exe create + New-Service + registry; EID 7045 fires; service masquerades as Windows service | Yes | Yes |
| T1546.003 | WMI Event Subscription Persistence | 1 | Real WMI EventFilter + CommandLineEventConsumer + FilterToConsumerBinding — actual WMI persistence objects | Yes | Yes |
| T1547.001 | Registry Run Key Persistence | 1 | Real reg add to HKCU\Run with hidden PS payload; Sysmon EID 13 fires on registry write | Yes | Yes |
| T1548.002 | UAC Bypass (Abuse Elevation Control Mechanism) | 1 | Real fodhelper + eventvwr + sdclt UAC bypass techniques using registry hijacking | Yes | Yes |
| T1550.002 | Pass the Hash — Credential Reuse Simulation | 2 | Real net use IPC$ + cmdkey explicit credential logon triggers EID 4648/4624 but no actual NTLM hash | No | No |
| T1552.001 | Credentials in Files | 2 | Creates fake credential files then searches for them; simulates attacker hunting but with fake data | Yes | Yes |
| T1558.003 | Kerberoasting — Service Ticket Request | 1 | Real setspn.exe enumeration + System.IdentityModel TGS requests; EID 4769 fires per ticket request | No | No |
| T1560.001 | Archive Collected Data — Archive via Utility | 2 | Real Compress-Archive on staging directory but only operates on simulation data, not real sensitive files | Yes | Yes |
| T1562.002 | Disable Windows Event Logging | 1 | Real auditpol commands disable actual audit subcategories; EID 4719 fires; cleanup restores policy | Yes | Yes |
| T1574.002 | DLL Side-Loading | 2 | Real Add-Type DLL compilation + LoadLibrary into process but benign DLL in simulation directory | Yes | Yes |
| UEBA-ACCOUNT-TAKEOVER | UEBA -- Account Takeover Chain | 2 | Real LogonUser API failures + enumeration burst but simulated chain, not actual credential theft | No | No |
| UEBA-SPRAY-CHAIN | UEBA — Credential Spray then Success Chain | 2 | Real LogonUser x25 failures + success triggers Exabeam spray pattern; but controlled simulation | No | No |
| UEBA-DATA-STAGING | UEBA -- Data Staging + Exfiltration Chain | 2 | Real file staging + HTTP POST attempt; .invalid C2 so connection fails but Sysmon EID 3 fires | Yes | Yes |
| UEBA-LATERAL-CHAIN | UEBA — Lateral Movement Discovery Chain | 3 | Rapid net user/ipconfig/netstat/arp burst; pure enumeration chain, no lateral movement | No | No |
| UEBA-LATERAL-NEW-ASSET | UEBA -- Lateral Movement + New Asset Access | 2 | Real net use SMB + Sysmon EID 3 on port 445; localhost/loopback target, not real lateral movement | Yes | Yes |
| UEBA-OFFHOURS | UEBA — Off-Hours Activity Simulation | 3 | whoami + net user + Get-Process off-hours; trivial enumeration at current system time | No | No |
| UEBA-PRIV-ESC | UEBA -- Privilege Escalation Chain | 3 | whoami /priv + net localgroup + WindowsIdentity .NET check; pure privilege enumeration, no escalation | No | No |
