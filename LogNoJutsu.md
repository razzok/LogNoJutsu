# LogNoJutsu — SIEM Validation & ATT&CK Simulation Tool

## Inhaltsverzeichnis

1. [Was ist LogNoJutsu?](#was-ist-lognojutsu)
2. [Entstehung und Motivation](#entstehung-und-motivation)
3. [Funktionsweise](#funktionsweise)
4. [Systemvoraussetzungen](#systemvoraussetzungen)
5. [Schnellstart](#schnellstart)
6. [Web-Oberfläche](#web-oberfläche)
7. [Systemvorbereitung (Preparation)](#systemvorbereitung-preparation)
8. [Simulationsphasen](#simulationsphasen)
9. [Multi-User-Simulation](#multi-user-simulation)
10. [Exabeam Use Case Abdeckung](#exabeam-use-case-abdeckung)
11. [Techniken — Vollständige Referenz](#techniken--vollständige-referenz)
    - [Phase 1: Discovery (Enumeration)](#phase-1-discovery-enumeration)
    - [Phase 2: Attack](#phase-2-attack)
    - [UEBA-Szenarien (Exabeam)](#ueba-szenarien-exabeam)
12. [Kampagnen / Playbooks](#kampagnen--playbooks)
13. [Logging und Reporting](#logging-und-reporting)
14. [Cleanup-Mechanismus](#cleanup-mechanismus)
15. [Kommandozeilen-Optionen](#kommandozeilen-optionen)

---

## Was ist LogNoJutsu?

LogNoJutsu ist ein **SIEM-Validierungswerkzeug**, das reales Angreiferverhalten auf einem Windows-System simuliert. Ziel ist es, nach dem Onboarding einer SIEM-Lösung automatisiert zu überprüfen, ob alle relevanten Detektionen und Use Cases korrekt anschlagen.

Das Tool führt **keine echten Angriffe** durch und extrahiert keine Credentials oder sensiblen Daten. Stattdessen werden die Aktionen und Systemartefakte erzeugt, die ein realer Angreifer hinterlassen würde — sprich: die Windows Event Logs, Sysmon-Events und PowerShell-Logs, anhand derer ein SIEM einen Angriff erkennen soll.

**Kernprinzip:** Wenn LogNoJutsu eine Technik ausführt und das SIEM diese *nicht* erkennt, liegt ein Problem im SIEM-Onboarding, in der Log-Weiterleitung oder in der Regelkonfiguration vor.

---

## Entstehung und Motivation

LogNoJutsu entstand als Open-Source-Eigenentwicklung inspiriert durch **Magneto**, ein internes Tool des SIEM-Herstellers Exabeam. Magneto wird auf Exabeam-Events vorgestellt und steht ausschließlich intern zur Verfügung. Es war in PowerShell geschrieben, startete über eine Executable, bot eine einfache Weboberfläche und führte nach konfigurierbaren Zeitintervallen zunächst kleine Anomalie- und Enumerationsaktionen aus, bevor vollständige Angriffssimulationen abliefen.

LogNoJutsu setzt diese Idee als eigenständiges, erweiterbares Tool um — mit besonderem Fokus auf **Exabeam UEBA-Use-Cases**, aber so gestaltet, dass Standard-ATT&CK-Techniken auch mit anderen SIEM-Lösungen (Splunk, Microsoft Sentinel, IBM QRadar etc.) detektiert werden können.

---

## Funktionsweise

```
┌─────────────────────────────────────────────────────────┐
│  lognojutsu.exe                                         │
│                                                         │
│  1. Startet HTTP-Server (localhost:8080)                │
│  2. Öffnet Web-UI im Browser                           │
│  3. Wartet auf Benutzerinteraktion — NICHTS läuft       │
│     automatisch                                         │
│                                                         │
│  Nach Klick auf "Start Simulation":                     │
│                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │ PREP PHASE   │ →  │ PHASE 1      │ →  │ PHASE 2  │  │
│  │ (manuell)    │    │ Discovery    │    │ Attack   │  │
│  │              │    │ (nach T1)    │    │ (nach T2)│  │
│  └──────────────┘    └──────────────┘    └──────────┘  │
│                                                         │
│  Jede Aktion wird in .log-Datei protokolliert          │
│  Am Ende: Cleanup aller angelegten Artefakte            │
└─────────────────────────────────────────────────────────┘
```

Das Tool ist eine **einzelne `.exe`-Datei** (~10 MB). Es werden keine zusätzlichen Laufzeitumgebungen, keine externe Software und keine Internetverbindung benötigt (außer beim automatischen Sysmon-Download während der Vorbereitung).

---

## Systemvoraussetzungen

| Anforderung | Details |
|---|---|
| Betriebssystem | Windows 10 / Windows 11 / Windows Server 2016+ |
| Berechtigungen | Normaler Benutzer für Discovery-Techniken; Administrator für Attack-Techniken und Preparation |
| PowerShell | Version 5.1+ (auf allen modernen Windows-Versionen vorhanden) |
| Netzwerk | Nur für automatischen Sysmon-Download in der Preparation-Phase erforderlich |
| SIEM-Agent | Vor dem Start sollte ein Log-Forwarder (Exabeam Agent, Winlogbeat, NXLog etc.) konfiguriert und aktiv sein |

---

## Schnellstart

```
# Standard — nur lokal erreichbar:
lognojutsu.exe

# Mit Netzwerkzugriff (SIEM-Engineer konfiguriert von eigenem Laptop):
lognojutsu.exe -host 0.0.0.0 -port 8080

# Mit Passwortschutz:
lognojutsu.exe -host 0.0.0.0 -port 8080 -password "MeinPasswort123"
```

Nach dem Start öffnet man `http://localhost:8080` im Browser. **Es läuft noch keine Simulation.**

---

## Web-Oberfläche

Die Web-UI besteht aus sieben Bereichen:

| Tab | Funktion |
|---|---|
| **Dashboard** | Aktueller Simulationsstatus, Phasen-Anzeige, Ausführungs-Timeline, erwartete SIEM-Events |
| **Preparation** | Einmalige Systemvorbereitung: Audit Policy, PowerShell-Logging, Sysmon-Installation |
| **Playbooks** | Übersicht aller verfügbaren Kampagnen und Einzeltechniken |
| **Configure & Run** | Kampagne auswählen, Zeitverzögerungen konfigurieren, Benutzer-Rotation, Simulation starten/stoppen |
| **Results** | Detaillierte Ergebnisse jeder ausgeführten Technik mit Output, Cleanup-Status und ausführendem Benutzer |
| **Simulation Log** | Live-Log-Stream aller Aktionen mit farbkodierter Ansicht nach Ereignistyp |
| **Users** | Benutzerprofile verwalten (local/AD), Credentials hinterlegen, Discovery, Credential-Test |

---

## Systemvorbereitung (Preparation)

Vor der ersten Simulation müssen einmalig drei Konfigurationsschritte durchgeführt werden. Diese erfordern **Administratorrechte** und werden über die Web-UI gestartet.

### 1. PowerShell ScriptBlock Logging

**Was wird konfiguriert:**
Setzt drei Registry-Schlüssel unter `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell`:

| Schlüssel | Wert | Zweck |
|---|---|---|
| `ScriptBlockLogging\EnableScriptBlockLogging` | `1` | Event 4104 aktivieren |
| `ModuleLogging\EnableModuleLogging` | `1` | Event 4103 aktivieren |
| `Transcription\EnableTranscripting` | `1` | PowerShell-Transkription aktivieren |

**Warum wichtig:** Ohne ScriptBlock Logging (4104) sind PowerShell-basierte Angriffe im SIEM praktisch unsichtbar. Dies ist die wichtigste Einstellung für die Erkennung von T1059.001 (79 Exabeam-Regeln), T1027 (47 Regeln) und aller weiteren PowerShell-lastigen Techniken.

### 2. Windows Audit Policy

**Was wird konfiguriert:**
12 Audit-Subkategorien werden über `auditpol.exe` aktiviert sowie Command-Line-Logging in Event 4688 über einen Registry-Schlüssel:

| Subkategorie | Erzeugte Events |
|---|---|
| Logon | 4624 (Erfolg), 4625 (Fehler), 4634 (Logoff) |
| Account Lockout | 4740 |
| Logon — Special Logon | 4672 |
| Process Creation | 4688 (inkl. Kommandozeile) |
| Scheduled Task | 4698, 4699 |
| Security Group Management | 4728, 4732 |
| User Account Management | 4720, 4726 |
| Audit Policy Change | 4719 |
| Directory Service Access | 4662 |
| Sensitive Privilege Use | 4673, 4674 |
| Other Object Access | 4698 |
| Kerberos Authentication | 4768, 4769, 4771, 4776 |

**Warum wichtig:** Event 4688 mit Kommandozeilen-Logging ist Voraussetzung für die Erkennung von LOLBin-Techniken (T1218, T1059.003). Event 4624/4625/4648 sind Kernvoraussetzungen für alle Credential- und Lateral-Movement-Use-Cases.

### 3. Sysmon Installation

Sysmon (System Monitor) von Sysinternals wird automatisch heruntergeladen und mit einer optimierten Konfiguration installiert. Folgende Event IDs werden konfiguriert:

| Sysmon Event ID | Beschreibung | Wichtig für |
|---|---|---|
| **1** | Process Create (mit Hashes, Parent, Cmdline) | Fast alle Techniken |
| **3** | Network Connection | T1046 (Port Scan), T1135 (Share Discovery) |
| **7** | Image Loaded (DLL) | T1218.011 (Rundll32) |
| **8** | CreateRemoteThread | Injection-Techniken |
| **10** | ProcessAccess | T1003.001 (LSASS-Zugriff) |
| **11** | FileCreate | T1486 (Ransomware), T1036.005 (Masquerading) |
| **12/13** | RegistryEvent | T1547.001, T1548.002 (UAC Bypass) |
| **19/20/21** | WMI Event | T1546.003 (WMI Subscription) |
| **22** | DNS Query | C2-Kommunikation, T1197 (BITS) |

---

## Simulationsphasen

### Phase 1 — Discovery ("Low & Slow")

Startet nach der konfigurierten Wartezeit T1 (Standard: 5 Sekunden im Test, in Produktion z.B. 10–30 Minuten). Führt alle 10 Enumeration-Techniken aus. Das Ziel ist es, die UEBA-Baseline zu stören und Recon-Verhalten zu erzeugen, das Anomalie-Detektionen im SIEM triggert.

### Phase 2 — Attack ("Full Attack")

Startet nach Abschluss von Phase 1 + Wartezeit T2 (Standard: 30 Sekunden im Test). Führt die ausgewählte Kampagne oder alle 26 Attack-Techniken aus. Hier entstehen die schwerwiegenden Artefakte (Persistence, Credential Access, Defense Evasion).

### Cleanup

Nach Ende der Simulation (oder bei manuellem Abbruch) werden alle angelegten Artefakte automatisch entfernt. Details: siehe [Cleanup-Mechanismus](#cleanup-mechanismus).

---

## Multi-User-Simulation

LogNoJutsu kann Techniken im Sicherheitskontext **anderer Benutzer** ausführen — sowohl lokale Windows-Konten als auch Active-Directory-Domänenbenutzer. Dies ist besonders wertvoll für die Exabeam-UEBA-Validierung, da Exabeam verhaltensbasierte Basislinien pro Benutzer aufbaut und Anomalien erkennt, wenn ein Benutzer ungewöhnliche Aktionen durchführt.

### Konzept: Benutzerprofile

Ein **Benutzerprofil** in LogNoJutsu besteht aus:

| Feld | Beschreibung |
|---|---|
| **Username** | Windows-Benutzername (ohne Domain) |
| **Domain** | Domain-Name für AD-Benutzer; leer für lokale Konten |
| **Password** | Passwort — wird verschlüsselt gespeichert (DPAPI) |
| **User Type** | `local` (lokaler Account), `domain` (AD-Konto), `current` (aktueller Benutzer) |
| **Display Name** | Optionaler Anzeigename für die UI |

Profile werden in `lognojutsu_users.json` im Arbeitsverzeichnis gespeichert. Die Datei ist mit Dateisystem-Berechtigungen `0600` geschützt.

### Passwort-Sicherheit: Windows DPAPI

Passwörter werden **niemals im Klartext** gespeichert. Stattdessen verwendet LogNoJutsu die **Windows Data Protection API (DPAPI)**:

```
Speichern:   PowerShell ConvertFrom-SecureString (ohne Schlüssel = DPAPI Machine+User-Kontext)
Lesen:       PowerShell ConvertTo-SecureString + SecureStringToBSTR
```

DPAPI bindet die Verschlüsselung an den Windows-Benutzer und die Maschine. Das bedeutet: Die `lognojutsu_users.json` kann nicht auf einem anderen System oder mit einem anderen Benutzer entschlüsselt werden.

> **Fallback:** Falls DPAPI nicht verfügbar ist (z.B. in Sandbox-Umgebungen), speichert LogNoJutsu das Passwort mit dem Präfix `PLAIN:`.

### Ausführungsmechanismus: ProcessStartInfo mit Credentials

Wenn eine Technik als anderer Benutzer ausgeführt werden soll, verwendet LogNoJutsu `System.Diagnostics.ProcessStartInfo` mit expliziten Credentials:

```powershell
$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = "powershell.exe"
$psi.UserName = "DOMAIN\username"    # oder ".\username" für lokale Konten
$psi.Password = $securePassword
$psi.UseShellExecute = $false
$psi.RedirectStandardOutput = $true
$proc = [System.Diagnostics.Process]::Start($psi)
```

**Erzeugte Windows Events:**

| Event ID | Log | Beschreibung |
|---|---|---|
| **4648** | Security | *A logon was attempted using explicit credentials* — Kernindikator für RunAs-Verhalten, zentrales Exabeam-UEBA-Signal |
| **4624** | Security | Erfolgreicher Logon des Zielbenutzers |
| **4688** | Security | Prozesserstellung mit Command-Line-Logging |
| **1** | Sysmon | Process Create mit ParentProcess = LogNoJutsu |

### Benutzer-Rotation

| Modus | Verhalten |
|---|---|
| **None** | Alle Techniken laufen als aktueller Benutzer (keine Profile nötig) |
| **Sequential** | Profile werden der Reihe nach zugewiesen (Technik 1 → User A, Technik 2 → User B, …) |
| **Random** | Jede Technik bekommt zufällig einen der konfigurierten Profile |

### Vollständiger Workflow

```
1. Tab "Users" → Benutzer hinzufügen (oder Discovery nutzen)
2. Credential-Test durchführen (grüner Haken = OK)
3. Tab "Configure & Run"
4. User Rotation Mode wählen: Sequential oder Random
5. Gewünschte Profile im Mehrfach-Select markieren (Ctrl+Klick)
6. Simulation starten
```

---

## Exabeam Use Case Abdeckung

LogNoJutsu deckt alle drei Exabeam TDIR Use Case Packages mit insgesamt 21 Use Cases ab. Die folgende Tabelle zeigt die Zuordnung:

| Exabeam Use Case Package | Use Case | Abdeckende Techniken |
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

**Exabeam-Regelabdeckung nach Technik (aus Content-Doc):**

| Technik | Exabeam-Regeln | Priorität |
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
| T1110.003 Password Spraying | **1** | T1110.003 (Gap-Test) |

> **Gap-Validierung:** Techniken mit wenigen Exabeam-Regeln (T1110.003 = 1 Regel, T1197 = 6, T1546.003 = 6) sind bewusst im Tool enthalten, um Lücken in der SIEM-Konfiguration sichtbar zu machen.

---

## Techniken — Vollständige Referenz

### Phase 1: Discovery (Enumeration)

Discovery-Techniken laufen in Phase 1 und erzeugen ausschließlich lesende Zugriffe. Sie dienen dazu, Recon-Verhalten für UEBA-Basislinien zu stören und Enumeration-Detektionen zu testen.

---

#### T1082 — System Information Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1082](https://attack.mitre.org/techniques/T1082/) |
| Taktik | Discovery |
| Exabeam-Regeln | 10 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Hostname, Username, Domain, OS-Version via PowerShell
$env:COMPUTERNAME; $env:USERNAME; $env:USERDOMAIN
(Get-CimInstance Win32_OperatingSystem).Caption

# Systeminformationen via Built-in-Tool
systeminfo | Select-String "OS Name|OS Version|System Type|Domain|Logon Server"
```

**Warum das Angreifer tun:** Systeminformationen sind der erste Schritt nach einer Kompromittierung. Der Angreifer ermittelt Betriebssystemversion (für Exploit-Auswahl), Domain-Zugehörigkeit (für Lateral Movement), und Logon-Server (für DC-Identifikation).

**Erwartete SIEM-Events:**
- `4688` — `systeminfo.exe` und `powershell.exe` Prozesserstellung
- `Sysmon 1` — Prozesserstellung mit vollständiger Kommandozeile und Hash

---

#### T1087 — Account Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1087.001](https://attack.mitre.org/techniques/T1087/001/) |
| Taktik | Discovery |
| Exabeam-Regeln | 25 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
net user                          # Alle lokalen Benutzerkonten
net localgroup administrators     # Mitglieder der Admin-Gruppe
net localgroup                    # Alle lokalen Gruppen
whoami /all                       # Aktueller Benutzer mit Rechten und SID
```

**Warum das Angreifer tun:** Account Discovery ist Voraussetzung für Privilege Escalation und Lateral Movement. Der Angreifer sucht nach privilegierten Konten (Domain Admins, lokale Admins), Dienstkonten und inaktiven Konten zum Missbrauch.

**Erwartete SIEM-Events:**
- `4688` — `net.exe` mehrfach mit verschiedenen Argumenten
- `Sysmon 1` — Prozesschain mit Kommandozeilen-Argumenten

---

#### T1049 — System Network Connections Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1049](https://attack.mitre.org/techniques/T1049/) |
| Taktik | Discovery |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
netstat -ano                                      # Alle aktiven Verbindungen mit PID
Get-NetTCPConnection -State Established           # Etablierte Verbindungen
Get-NetTCPConnection -State Listen                # Lauschende Ports
```

**Warum das Angreifer tun:** Aktive Netzwerkverbindungen zeigen dem Angreifer, welche Server das System kennt (Datenbankserver, Domain Controller, Share-Server) — potenzielle Lateral-Movement-Ziele.

**Erwartete SIEM-Events:**
- `4688` — `netstat.exe` Prozesserstellung
- `Sysmon 1` — `netstat.exe` mit `-ano`

---

#### T1016 — System Network Configuration Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1016](https://attack.mitre.org/techniques/T1016/) |
| Taktik | Discovery |
| Exabeam-Regeln | 5 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
ipconfig /all           # Alle Netzwerkadapter mit Details (IP, MAC, Gateway, DNS)
route print             # Routing-Tabelle (zeigt Subnetz-Struktur)
ipconfig /displaydns    # Lokaler DNS-Cache (zeigt bekannte Hostnamen)
```

**Warum das Angreifer tun:** Die Netzwerkkonfiguration verrät die Subnetz-Topologie, Gateway-Adressen für weiteres Pivoting, und der DNS-Cache zeigt kürzlich kontaktierte Systeme — wertvolle Recon-Information für die Angriffsplanung.

**Erwartete SIEM-Events:**
- `4688` — `ipconfig.exe` mit `/all` und `/displaydns`
- `Sysmon 1` — Prozesserstellung mit Argumenten

---

#### T1057 — Process Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1057](https://attack.mitre.org/techniques/T1057/) |
| Taktik | Discovery |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
tasklist /v             # Alle laufenden Prozesse mit Details

# Gezielte Suche nach Security-Tools (Attacker Awareness)
# Sysmon, Splunk, CrowdStrike, Defender, Carbon Black, Cylance,
# Sentinel, Qualys, Tanium, SentinelOne, Cybereason ...
Get-Process | Where-Object { $_.Name -match "sysmon|splunk|crowd..." }
```

**Warum das Angreifer tun:** Angreifer suchen nach Endpoint-Security-Tools, um diese zu umgehen oder zu deaktivieren. Die Suche nach Sicherheitssoftware-Prozessen ist ein charakteristisches "Attacker Awareness"-Verhalten, das Exabeam als T1057-Aktivität klassifiziert.

**Erwartete SIEM-Events:**
- `4688` — `tasklist.exe` mit `/v`
- `Sysmon 1` — Prozesserstellung

---

#### T1083 — File and Directory Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1083](https://attack.mitre.org/techniques/T1083/) |
| Taktik | Discovery |
| Exabeam-Regeln | 38 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Verzeichnisstruktur des User-Profils
Get-ChildItem -Path $env:USERPROFILE -Depth 1

# Gezielte Suche nach interessanten Dateitypen
# *.txt, *.docx, *.xlsx, *.pdf, *.ps1, *.bat, *.config,
# *.xml, *.json, *.key, *.pem, *.pfx (Zertifikate/Credentials)
Get-ChildItem -Path "C:\" -Recurse -ErrorAction SilentlyContinue
```

**Warum das Angreifer tun:** Angreifer suchen nach sensitiven Dateien — Konfigurationsdateien mit Passwörtern, SSH-Keys, Zertifikaten, Scripts mit hardcodierten Credentials, und Backup-Dateien. Die Suche nach `.pem`, `.pfx`, `.key` ist ein starkes Indiz für Credential-Hunting.

**Erwartete SIEM-Events:**
- `4688` — `powershell.exe` mit `Get-ChildItem`
- `Sysmon 1` — Prozesserstellung

---

#### T1069 — Permission Groups Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1069.001](https://attack.mitre.org/techniques/T1069/001/) |
| Taktik | Discovery |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
net localgroup                        # Alle lokalen Gruppen
net localgroup "Administrators"       # Mitglieder der Admin-Gruppe
net localgroup "Remote Desktop Users" # RDP-berechtigte Benutzer

# Domain-Abfrage (nur wenn System domain-joined):
net group /domain
```

**Warum das Angreifer tun:** Gruppenabfragen zeigen, welche Accounts privilegierten Zugriff haben. "Remote Desktop Users" ist ein häufiges Ziel für Lateral Movement via RDP. Domain-Gruppen wie "Domain Admins" oder "Enterprise Admins" sind die ultimativen Ziele.

**Erwartete SIEM-Events:**
- `4688` — `net.exe` mehrfach mit Gruppenargumenten
- Bei Domain-Abfrage: `4661` (LDAP-Suche)

---

#### T1046 — Network Service Discovery (Port Scan)

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1046](https://attack.mitre.org/techniques/T1046/) |
| Taktik | Discovery |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Scan der wichtigsten Angreifer-Ports auf localhost und Gateway
$ports = @(21, 22, 23, 25, 53, 80, 135, 139, 443, 445,
           1433, 3306, 3389, 5985, 8080)
foreach ($port in $ports) {
    Test-NetConnection -ComputerName "127.0.0.1" -Port $port
}
```

**Warum das Angreifer tun:** Port-Scanning nach der initialen Kompromittierung dient der Identifikation laufender Services für weitere Angriffsvektoren. Port 3389 (RDP), 445 (SMB), 1433 (SQL Server) und 5985 (WinRM) sind Hauptziele für Lateral Movement.

**Erwartete SIEM-Events:**
- `Sysmon 3` — NetworkConnect-Events für jede Verbindungsattempt
- `4688` — `powershell.exe` Prozesserstellung

---

#### T1135 — Network Share Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1135](https://attack.mitre.org/techniques/T1135/) |
| Taktik | Discovery |
| Exabeam-Regeln | 12 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
net share                                   # Lokale Freigaben
net view \\$env:COMPUTERNAME                # Freigaben auf lokalem Host
Get-SmbShare                                # PowerShell SMB-Enumeration

# Zugriff auf Admin-Shares (generiert Event 5140)
foreach ($share in @("C$", "IPC$", "ADMIN$")) {
    Test-Path "\\$env:COMPUTERNAME\$share"
}
```

**Warum das Angreifer tun:** Netzwerkfreigaben sind primäre Exfiltrationsziele. Admin-Shares (`C$`, `ADMIN$`) ermöglichen Remote-Code-Execution. Event 5140 (Network Share Object Access) ist ein wichtiges Exabeam-Signal für ungewöhnlichen Share-Zugriff.

**Erwartete SIEM-Events:**
- `4688` — `net.exe` mit `view` und `share`
- `5140` — Network share object accessed
- `Sysmon 3` — SMB-Verbindungen (Port 445)

---

#### T1482 — Domain Trust Discovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1482](https://attack.mitre.org/techniques/T1482/) |
| Taktik | Discovery |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# nltest — häufigstes Angreifer-Tool für Trust-Enumeration
nltest /domain_trusts           # Alle Trust-Beziehungen
nltest /dclist:$env:USERDOMAIN  # Domain Controller auflisten

# PowerShell .NET Trust-Enumeration
$domain = [System.DirectoryServices.ActiveDirectory.Domain]::GetCurrentDomain()
$domain.GetAllTrustRelationships()
```

**Warum das Angreifer tun:** Domain-Trusts sind die Brücken für Forest-übergreifendes Lateral Movement. Ein Angreifer, der eine Domain-Vertrauensbeziehung kennt, kann sich in vertraute Domains bewegen. `nltest.exe` mit `/domain_trusts` ist ein so charakteristisches Angreifer-Muster, dass viele EDR-Lösungen diesen Aufruf direkt als Indicator of Compromise werten.

**Erwartete SIEM-Events:**
- `4688` — `nltest.exe` mit `/domain_trusts`
- `Sysmon 1` — `nltest.exe` Prozesserstellung

---

### Phase 2: Attack

Attack-Techniken simulieren die eigentlichen Angriffs- und Post-Exploitation-Aktionen. Viele dieser Techniken erfordern Administratorrechte und legen Artefakte an, die im Cleanup entfernt werden.

---

#### T1059.001 — PowerShell Execution

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1059.001](https://attack.mitre.org/techniques/T1059/001/) |
| Taktik | Execution |
| Exabeam-Regeln | **79** |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# 1. Encoded Command — typisches Angreifer-Obfuscation-Muster
$command = "Write-Host 'LogNoJutsu: Simulated payload'; Get-Date; whoami"
$encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
powershell.exe -NonInteractive -EncodedCommand $encoded

# 2. Invoke-Expression — simuliertes Download-Cradle (IEX) Muster
$simulatedPayload = { Get-Process | Select-Object -First 5 }
Invoke-Expression ($simulatedPayload.ToString())
```

**Warum das Angreifer tun:** PowerShell ist das wichtigste Angreifer-Tool auf Windows-Systemen. Die `-EncodedCommand`-Flag ist das Standardmuster für Obfuscation. `Invoke-Expression` (IEX) kombiniert mit Download-Cradles ist das Muster für dateilose Malware. Mit 79 dedizierten Exabeam-Regeln ist T1059.001 eine der wichtigsten zu testenden Techniken.

**Erwartete SIEM-Events:**
- `4688` — `powershell.exe` mit `-EncodedCommand` in der Kommandozeile
- `4104` — ScriptBlock-Logging des dekodierten Befehls
- `4103` — Module-Logging
- `Sysmon 1` — Prozesserstellung mit Base64-Payload im Argument

---

#### T1059.003 — Windows Command Shell

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1059.003](https://attack.mitre.org/techniques/T1059/003/) |
| Taktik | Execution |
| Exabeam-Regeln | 34 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```cmd
cmd.exe /C "whoami /all"
cmd.exe /C "net user"
cmd.exe /C "net localgroup administrators"
cmd.exe /C "systeminfo | findstr /B /C:`"OS Name`" /C:`"Domain`""
cmd.exe /C "dir C:\Users /AD"
```

**Warum das Angreifer tun:** `cmd.exe` ist auf jedem Windows-System vorhanden und wird von Angreifern für schnelle System-Recon und als Shell nach Exploitation genutzt. Die charakteristischen Befehle (`whoami`, `net user`, `systeminfo`) sind starke SIEM-Signale, da normale Benutzer diese selten ausführen.

**Erwartete SIEM-Events:**
- `4688` — `cmd.exe` mit `/C` und verdächtigen Argumenten (mehrfach)
- `Sysmon 1` — Prozesserstellung mit vollständiger Kommandozeile

---

#### T1027 — Obfuscated Files or Information

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1027](https://attack.mitre.org/techniques/T1027/) |
| Taktik | Defense Evasion |
| Exabeam-Regeln | 47 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# 1. Base64-encoded Command (häufigstes Muster in der Praxis)
$encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($payload))
powershell.exe -EncodedCommand $encoded

# 2. String-Concatenation Obfuscation (umgeht einfache Signaturen)
$a = "Get-"; $b = "Pro"; $c = "cess"
Invoke-Expression ($a + $b + $c)

# 3. Tick-Mark Obfuscation (PowerShell Escape-Zeichen als Obfuscation)
G`et-`Hos`tN`ame

# 4. Double-Encoded Command (triggert Exabeam-Anomalie für verschachtelte Encoding)
powershell.exe -EncodedCommand <base64(powershell.exe -EncodedCommand <base64>)>
```

**Warum das Angreifer tun:** Obfuscation ist der primäre Mechanismus, um signaturbasierte Erkennungen zu umgehen. Exabeam hat 47 dedizierte Regeln für diese Technik, weil es ein universelles Angreifer-Verhalten ist. Besonders Double-Encoding ist ein starkes Signal, da kein legitimes Skript so vorgehen würde.

**Erwartete SIEM-Events:**
- `4104` — ScriptBlock-Logging zeigt obfuszierten Code
- `4688` — `powershell.exe` mit `-EncodedCommand` oder `-Enc` Flag
- `Sysmon 1` — Prozessargument enthält Base64-String

---

#### T1218.011 — Rundll32 Proxy Execution

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1218.011](https://attack.mitre.org/techniques/T1218/011/) |
| Taktik | Defense Evasion |
| Exabeam-Regeln | 27 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Rundll32 shell32.dll — generiert Sysmon Event 1 und 7 (DLL loaded)
Start-Process "rundll32.exe" -ArgumentList "shell32.dll,Control_RunDLL"

# Rundll32 advpack.dll — häufig in Malware für INF-Execution
Start-Process "rundll32.exe" -ArgumentList "advpack.dll,DelNodeRunDLL32 test.inf"

# Rundll32 url.dll — phishing-typisches Muster
Start-Process "rundll32.exe" -ArgumentList "url.dll,FileProtocolHandler"
```

**Warum das Angreifer tun:** `rundll32.exe` ist eine signierte Windows-Binärdatei (LOLBin), die beliebige DLL-Funktionen ausführen kann. Angreifer nutzen sie, um Application-Whitelisting zu umgehen, da `rundll32.exe` selbst als vertrauenswürdig gilt. Das Exabeam-Regelwerk für T1218 (116 Regeln gesamt) ist eines der umfangreichsten.

**Erwartete SIEM-Events:**
- `4688` — `rundll32.exe` mit ungewöhnlichen Argumenten
- `Sysmon 1` — Prozesserstellung mit DLL-Argument
- `Sysmon 7` — ImageLoaded — DLL wird durch Rundll32 geladen

---

#### T1047 — WMI Execution

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1047](https://attack.mitre.org/techniques/T1047/) |
| Taktik | Execution |
| Exabeam-Regeln | 18 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# WMIC Prozess-Enumeration
wmic process list brief

# WMIC lokale Prozesserstellung (Kindprozess von WmiPrvSE.exe)
wmic process call create "cmd.exe /C whoami"

# PowerShell Invoke-WmiMethod (generiert Sysmon Event 20)
Invoke-WmiMethod -Class Win32_Process -Name Create -ArgumentList "cmd.exe /C hostname"

# WMI Systeminformationen (Recon über WMI-Interface)
Get-WmiObject -Class Win32_OperatingSystem
Get-WmiObject -Class Win32_ComputerSystem
```

**Warum das Angreifer tun:** WMI ist ein nativer Windows-Mechanismus, der Prozesse ohne direkten `CreateProcess()`-Aufruf starten kann. Prozesse, die über WMI gestartet werden, haben `WmiPrvSE.exe` als Parent-Prozess statt `cmd.exe` oder `powershell.exe` — eine klassische Defense-Evasion-Technik. WMI-basierte Execution ist schwer zu erkennen ohne Sysmon Event 20.

**Erwartete SIEM-Events:**
- `4688` — `wmic.exe` Prozesserstellung
- `Sysmon 1` — Kindprozesse mit `WmiPrvSE.exe` als Parent
- `Sysmon 20` — WMI Activity Events

---

#### T1110.001 — Password Brute Force

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1110.001](https://attack.mitre.org/techniques/T1110/001/) |
| Taktik | Credential Access |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# 10 fehlgeschlagene NTLM-Authentifizierungsversuche
# Ziel-Accounts existieren NICHT — keine echten Konten werden gesperrt
for ($i = 1; $i -le 10; $i++) {
    $ctx = New-Object System.DirectoryServices.AccountManagement.PrincipalContext(...)
    $ctx.ValidateCredentials("lognojutsu_nonexistent_$i", "WrongPassword$i!")
    Start-Sleep -Milliseconds 200
}
```

**Warum das Angreifer tun:** Brute-Force-Angriffe auf Passwörter sind der klassischste Credential-Access-Vektor. Die Simulation erzeugt 10 Event-4625-Einträge in schneller Abfolge — das Basis-Erkennungsmuster für Brute-Force in nahezu jedem SIEM.

**Erwartete SIEM-Events:**
- `4625` × 10 — "Account failed to log on" in schneller Abfolge
- `4740` — Account Lockout (wenn Lockout-Policy greift)

---

#### T1110.003 — Password Spraying

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1110.003](https://attack.mitre.org/techniques/T1110/003/) |
| Taktik | Credential Access |
| Exabeam-Regeln | **1** (Gap-Validierung!) |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Enumerate alle aktivierten lokalen Accounts
$accounts = Get-LocalUser | Where-Object { $_.Enabled }

# Ein einziges Passwort gegen alle Accounts — verhindert Lockout
foreach ($account in $accounts | Select-Object -First 5) {
    $ctx.ValidateCredentials($account.Name, "Password1_SPRAY_SIM_INVALID")
    Start-Sleep -Milliseconds 500  # Langsame Rate = Low-and-Slow-Muster
}
```

**Warum das Angreifer tun:** Password Spraying umgeht Account-Lockout-Richtlinien, indem nur ein Passwort pro Account versucht wird. Dies ist eine der häufigsten Initial-Access-Techniken in echten Incidents (Microsoft, Okta, SolarWinds alle betroffen). **Mit nur 1 Exabeam-Regel ist dies eine der wichtigsten Gap-Validierungen.**

**Erwartete SIEM-Events:**
- `4625` — Fehlgeschlagene Anmeldungen über mehrere Accounts verteilt
- `4771` — Kerberos pre-auth failed (auf Domain-Systemen)

---

#### T1003.001 — LSASS Memory Access

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1003.001](https://attack.mitre.org/techniques/T1003/001/) |
| Taktik | Credential Access |
| Exabeam-Regeln | 18 |
| Admin erforderlich | **Ja** |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Windows-API: Handle auf lsass.exe mit PROCESS_QUERY_INFORMATION (0x0400)
# Das ist exakt der erste Schritt, den Mimikatz ausführt
Add-Type -MemberDefinition $sig -Name WinAPI -Namespace LogNoJutsu
$handle = [LogNoJutsu.WinAPI]::OpenProcess(0x0400, $false, $lsassPID)

# Handle sofort schließen — KEINE Credential-Extraktion
[LogNoJutsu.WinAPI]::CloseHandle($handle)
```

**Warum das Angreifer tun:** LSASS (Local Security Authority Subsystem Service) speichert Passwort-Hashes und Kerberos-Tickets im Speicher. Tools wie Mimikatz, Procdump und Task Manager können den LSASS-Prozess dumpen. Der Handle-Zugriff auf `lsass.exe` ist der primäre Sysmon-10-Trigger — dasselbe Signal, das Mimikatz auslöst.

**Erwartete SIEM-Events:**
- `Sysmon 10` — ProcessAccess: `TargetImage = lsass.exe` — **primäres Credential-Dumping-Signal**
- `4656` — Handle-Anfrage auf `lsass.exe` (bei aktivierter Objekt-Auditierung)

---

#### T1003.006 — DCSync Simulation

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1003.006](https://attack.mitre.org/techniques/T1003/006/) |
| Taktik | Credential Access |
| Exabeam-Regeln | 4 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Enumerate Domain Controller (attacker recon vor DCSync)
nltest /dclist:$env:USERDOMAIN
nltest /dsgetdc:$env:USERDOMAIN

# Accounts mit DS-Replication-Rechten suchen (DCSync-Ziel-Identifikation)
$searcher = [ADSISearcher]"(&(objectClass=user)(userAccountControl:...:=512))"

# Domain Admins / Enterprise Admins auflisten (DCSync-fähige Gruppen)
foreach ($group in @("Domain Admins", "Enterprise Admins")) {
    $groupObj = [ADSI]"LDAP://CN=$group,CN=Users,$domainDN"
}

# Domain-Objekt ACL auf Replication-Rechte prüfen
Get-Acl "AD:\$((Get-ADDomain).DistinguishedName)"
```

**Warum das Angreifer tun:** DCSync missbraucht das MS-DRSR-Protokoll (Directory Replication Service Remote Protocol) um Passwort-Hashes direkt vom Domain Controller zu replizieren — ohne Code auf dem DC auszuführen. Das einzige sichtbare Signal ist Event 4662 (Directory Service object access) mit den Replication-Rechte-GUIDs. Mimikatz-Befehl: `lsadump::dcsync /domain:corp /user:Administrator`.

**Erwartete SIEM-Events:**
- `4662` — Directory Service object access (Replication rights)
- `4688` — `nltest.exe` Prozesserstellung
- `4769` — Kerberos TGS für DRSUAPI-Service

---

#### T1552.001 — Credentials in Files

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1552.001](https://attack.mitre.org/techniques/T1552/001/) |
| Taktik | Credential Access |
| Exabeam-Regeln | 2 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Durchsucht bekannte Credential-Speicher-Pfade
$paths = @($env:USERPROFILE, $env:APPDATA, "C:\inetpub", "C:\xampp")
$patterns = @("password", "passwd", "secret", "apikey", "connectionstring")
$extensions = @("*.xml", "*.ini", "*.config", "*.txt", "*.ps1", "*.bat")

foreach ($path in $paths) {
    Get-ChildItem $path -Filter $ext -Recurse |
        Where-Object { (Get-Content $_.FullName) -match $pattern }
}

# findstr als cmd-Variante (generiert 4688 für findstr.exe)
cmd.exe /C "findstr /si password $env:USERPROFILE\*.xml *.ini *.txt"
```

**Warum das Angreifer tun:** Konfigurationsdateien, Deployment-Skripte und Anwendungs-Configs enthalten häufig Passwörter im Klartext. Web-Server-Konfigurationen (IIS, Apache), Datenbankverbindungsstrings und PowerShell-Skripte sind die häufigsten Fundorte. `findstr /si password` ist ein bekanntes Angreifer-Kommando.

**Erwartete SIEM-Events:**
- `4104` — ScriptBlock-Logging: Dateisystem-Traversal mit Credential-Suchmustern
- `4688` — `findstr.exe` mit `password`-Argument

---

#### T1558.003 — Kerberoasting

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1558.003](https://attack.mitre.org/techniques/T1558/003/) |
| Taktik | Credential Access |
| Exabeam-Regeln | 22 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# 1. LDAP-Abfrage: Accounts mit Service Principal Names (Kerberoasting-Ziele)
$spnAccounts = ([ADSISearcher]"(&(objectCategory=user)(servicePrincipalName=*))").FindAll()

# 2. Kerberos Service Tickets für gefundene SPNs anfordern (generiert Event 4769)
Add-Type -AssemblyName System.IdentityModel
foreach ($spn in $discoveredSPNs) {
    $ticket = New-Object System.IdentityModel.Tokens.KerberosRequestorSecurityToken($spn)
    # Ticket wird im Speicher gecacht — normalerweise dann offline geknackt
}

# 3. Gecachte Kerberos-Tickets anzeigen
klist
```

**Warum das Angreifer tun:** Kerberoasting ermöglicht es, Passwort-Hashes von Service-Accounts offline zu knacken, ohne Admin-Rechte zu benötigen. Der Angreifer fordert Service-Tickets für Accounts mit SPNs an (normales Kerberos-Verhalten), extrahiert den verschlüsselten Teil und knackt ihn offline mit Hashcat. Event 4769 mit RC4-Verschlüsselung (etype 23) statt AES ist das Erkennungssignal.

**Erwartete SIEM-Events:**
- `4769` — Kerberos Service Ticket Request — **primäres Kerberoasting-Signal**
- `4768` — Kerberos TGT Request
- `4104` — ScriptBlock: SPN-Enumeration via LDAP

---

#### T1550.002 — Pass the Hash Pattern

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1550.002](https://attack.mitre.org/techniques/T1550/002/) |
| Taktik | Lateral Movement |
| Exabeam-Regeln | 23 |
| Admin erforderlich | Nein |
| Cleanup | Ja — Netzwerkverbindungen |

**Was wird ausgeführt:**

```powershell
# NTLM-basierter Netzwerk-Share-Zugriff (generiert Event 4776/4624 type 3)
net use \\$localHost\IPC$ /user:$env:USERNAME ""

# Lateral Attempts zu mehreren Hosts (PtH-Spray-Muster)
foreach ($target in @($env:COMPUTERNAME, "127.0.0.1", "localhost")) {
    net use \\$target\IPC$
}

# Explicit Credential Use (generiert Event 4648 — core PtH-Signal)
cmdkey /add:$env:COMPUTERNAME /user:"$domain\$username" /pass:"..."
```

**Warum das Angreifer tun:** Pass-the-Hash verwendet den NTLM-Hash eines Passworts statt des Klartexts für Authentifizierung. Das primäre Erkennungssignal sind Event 4648 (explicit credential use) kombiniert mit Event 4624 Typ 3 (network logon) via NTLM. Exabeam hat 23 dedizierte Regeln für dieses Muster. Der tatsächliche PtH-Angriff erfordert Mimikatz (`sekurlsa::pth`).

**Erwartete SIEM-Events:**
- `4648` — Logon mit expliziten Credentials — **primäres PtH-Signal**
- `4624` Typ 3 — Netzwerk-Logon via NTLM
- `4776` — NTLM Credential-Validierung

---

#### T1136.001 — Create Local Account

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1136.001](https://attack.mitre.org/techniques/T1136/001/) |
| Taktik | Persistence |
| Exabeam-Regeln | 10 |
| Admin erforderlich | **Ja** |
| Cleanup | **Ja** — Account wird gelöscht |

**Was wird angelegt:**

```
Benutzername: lnj_test_acct
Passwort:     LogNoJutsu!Temp2024
Kommentar:    LogNoJutsu SIEM validation test account
```

**Was wird ausgeführt:**

```cmd
net user lnj_test_acct LogNoJutsu!Temp2024 /add /comment:"LogNoJutsu test"
```

**Cleanup:**

```powershell
net user lnj_test_acct /delete
```

**Warum das Angreifer tun:** Das Anlegen eines backdoor-Benutzerkontos ist eine der persistentesten Hintertüren, die ein Angreifer hinterlassen kann. Event 4720 ist das direkte Signal. Exabeam prüft zusätzlich, ob das neue Konto ungewöhnliche Eigenschaften hat (z.B. kein Passwort-Ablaufdatum, unbekannte Naming Convention).

**Erwartete SIEM-Events:**
- `4720` — User account created — **Kern-Event für Account Manipulation Use Case**
- `4722` — User account enabled
- `4688` — `net.exe` mit `/add` Argument

---

#### T1098 — Account Manipulation (Add to Administrators)

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1098](https://attack.mitre.org/techniques/T1098/) |
| Taktik | Persistence, Privilege Escalation |
| Exabeam-Regeln | **57** |
| Admin erforderlich | **Ja** |
| Cleanup | **Ja** — Gruppen-Mitgliedschaft und Account entfernt |

**Was wird ausgeführt:**

```powershell
# Test-Account anlegen (wenn nicht vorhanden)
net user lnj_test_acct "LogNoJutsu!Temp2024" /add

# Zu lokalen Administrators hinzufügen (generiert Event 4732)
net localgroup Administrators lnj_test_acct /add
```

**Cleanup:**

```powershell
net localgroup Administrators lnj_test_acct /delete
net user lnj_test_acct /delete
```

**Warum das Angreifer tun:** Ein bereits kompromittiertes Nicht-Admin-Konto wird in die Administrators-Gruppe aufgenommen, um persistente Adminrechte zu sichern. Mit 57 dedizierten Exabeam-Regeln ist T1098 einer der best-detektierten Techniken — ein erfolgreicher Test zeigt, dass die Gruppenänderungs-Auditierung korrekt funktioniert.

**Erwartete SIEM-Events:**
- `4732` — Member added to security-enabled local group — **Kern-Event**
- `4728` — Member added to global security group (Domain-Kontext)

---

#### T1548.002 — UAC Bypass via Event Viewer

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1548.002](https://attack.mitre.org/techniques/T1548/002/) |
| Taktik | Privilege Escalation, Defense Evasion |
| Exabeam-Regeln | 10 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — Registry-Key entfernt |

**Was wird angelegt:**

```
Registry: HKCU\Software\Classes\mscfile\shell\open\command
Wert:     cmd.exe /K echo LogNoJutsu-UAC-Bypass-Simulation
```

**Was wird ausgeführt:**

```powershell
# Registry-Hijack setzen (Sysmon Event 12/13)
$regPath = "HKCU:\Software\Classes\mscfile\shell\open\command"
New-Item -Path $regPath -Force | Out-Null
Set-ItemProperty -Path $regPath -Name "(default)" -Value "cmd.exe /K echo UAC-Bypass-Sim"

# eventvwr.exe triggern (liest den manipulierten Registry-Key)
Start-Process "eventvwr.exe" -WindowStyle Hidden
```

**Cleanup:**

```powershell
Remove-Item -Path "HKCU:\Software\Classes\mscfile" -Recurse -Force
```

**Warum das Angreifer tun:** Die eventvwr-UAC-Bypass-Methode ermöglicht es, Code als Administrator auszuführen, ohne einen UAC-Dialog zu erzeugen. `eventvwr.exe` liest den `mscfile`-Shell-Handler aus der Registry — durch Überschreiben in `HKCU` (ohne Admin-Rechte möglich) kann beliebiger Code mit erhöhten Rechten ausgeführt werden. Das Sysmon-13-Event auf diesen spezifischen Registry-Pfad ist das Erkennungssignal.

**Erwartete SIEM-Events:**
- `Sysmon 12` — RegistryEvent (Key created): `HKCU\Software\Classes\mscfile\...`
- `Sysmon 13` — RegistryEvent (Value set) — **UAC Bypass Indicator**
- `4688` — `eventvwr.exe` Prozesserstellung

---

#### T1547.001 — Registry Run Key Persistence

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1547.001](https://attack.mitre.org/techniques/T1547/001/) |
| Taktik | Persistence |
| Exabeam-Regeln | 10 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — Registry-Eintrag wird entfernt |

**Was wird angelegt:**

```
Pfad:  HKCU\Software\Microsoft\Windows\CurrentVersion\Run
Name:  LogNoJutsu_Persistence_Test
Wert:  C:\Windows\System32\notepad.exe
```

**Was wird ausgeführt:**

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

**Warum das Angreifer tun:** Run-Keys sind der einfachste Persistenz-Mechanismus auf Windows — sie werden bei jedem Benutzer-Login ausgeführt. Angreifer verwenden sie um Backdoors nach Reboots zu erhalten. Sysmon Event 13 auf `CurrentVersion\Run` ist ein direktes Erkennungssignal.

**Erwartete SIEM-Events:**
- `Sysmon 13` — RegistryEvent (Value Set) auf `CurrentVersion\Run`
- `4688` — `reg.exe` oder `powershell.exe`

---

#### T1053.005 — Scheduled Task Persistence

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1053.005](https://attack.mitre.org/techniques/T1053/005/) |
| Taktik | Persistence |
| Exabeam-Regeln | 27 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — Task wird deregistriert |

**Was wird angelegt:**

```
Task-Name:    LogNoJutsu_Task_Test
Ausführung:   notepad.exe
Trigger:      Bei Benutzer-Anmeldung
Einstellung:  Versteckt (Hidden)
```

**Was wird ausgeführt:**

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

**Warum das Angreifer tun:** Scheduled Tasks sind persistenter als Run-Keys, da sie auch unter anderen Benutzerkontexten ausgeführt werden können. Das `Hidden`-Flag ist eine Standard-Angreifer-Technik. Event 4698 ist das direkte Erkennungssignal.

**Erwartete SIEM-Events:**
- `4698` — "A scheduled task was created" im Security-Log
- `4688` — `schtasks.exe` Prozesserstellung

---

#### T1543.003 — Create Windows Service

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1543.003](https://attack.mitre.org/techniques/T1543/003/) |
| Taktik | Persistence |
| Exabeam-Regeln | 38 |
| Admin erforderlich | **Ja** |
| Cleanup | **Ja** — Service wird gelöscht |

**Was wird angelegt:**

```
Service-Name:    LogNoJutsuTestSvc
Display-Name:    LogNoJutsu Test Service
Binary-Pfad:     C:\Windows\System32\notepad.exe
Start-Typ:       Manuell (demand)
```

**Was wird ausgeführt:**

```cmd
sc.exe create LogNoJutsuTestSvc binPath= "C:\Windows\System32\notepad.exe" ^
    DisplayName= "LogNoJutsu Test Service" start= demand
```

**Cleanup:**

```cmd
sc.exe delete LogNoJutsuTestSvc
```

**Warum das Angreifer tun:** Malware-Services überleben Reboots und laufen typischerweise als SYSTEM. Die Kombination aus Event 7045 (Service installed) und einem unbekannten Binary-Pfad ist ein starkes Exabeam-Signal. APTs wie Cobalt Strike und Empire nutzen Service-Installation als primären Persistence-Mechanismus.

**Erwartete SIEM-Events:**
- `7045` — "A new service was installed" im System-Log
- `4697` — "A service was installed" im Security-Log

---

#### T1197 — BITS Jobs Persistence

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1197](https://attack.mitre.org/techniques/T1197/) |
| Taktik | Persistence, Defense Evasion |
| Exabeam-Regeln | 6 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — BITS-Job wird abgebrochen |

**Was wird ausgeführt:**

```cmd
bitsadmin /create LogNoJutsu_BITS_Test
bitsadmin /setnotifycmdline LogNoJutsu_BITS_Test "cmd.exe" "/C echo BITS-Persistence"
bitsadmin /info LogNoJutsu_BITS_Test
```

**Cleanup:**

```cmd
bitsadmin /cancel LogNoJutsu_BITS_Test
```

**Warum das Angreifer tun:** BITS (Background Intelligent Transfer Service) ist ein legitimer Windows-Service für Dateiübertragungen. Angreifer missbrauchen ihn für stealthy Downloads und Persistence via Notification-Commands. BITS-Jobs überleben Reboots, laufen als Systemservice, und werden von vielen AV-Lösungen nicht überwacht. Erkennungsevents befinden sich im `Microsoft-Windows-Bits-Client/Operational` Log — viele SIEMs ignorieren diese Quelle.

**Erwartete SIEM-Events:**
- `BITS-Client Event 3` — BITS Job created
- `BITS-Client Event 59` — BITS Job transfer started
- `4688` — `bitsadmin.exe` Prozesserstellung

---

#### T1546.003 — WMI Event Subscription

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1546.003](https://attack.mitre.org/techniques/T1546/003/) |
| Taktik | Persistence |
| Exabeam-Regeln | 6 |
| Admin erforderlich | **Ja** |
| Cleanup | **Ja** — Filter, Consumer und Binding entfernt |

**Was wird angelegt:**

```
WMI Filter:   LogNoJutsu_WMI_Filter  (SystemUpTime >= 200s)
WMI Consumer: LogNoJutsu_WMI_Consumer (cmd.exe /C echo ...)
WMI Binding:  Filter → Consumer
```

**Was wird ausgeführt:**

```powershell
# WMI Event Filter (Sysmon Event 19)
$wmiFilter = Set-WmiInstance -Namespace root\subscription -Class __EventFilter `
    -Arguments @{ Name = "LogNoJutsu_WMI_Filter"; Query = $filterQuery }

# WMI CommandLine Consumer (Sysmon Event 20)
$wmiConsumer = Set-WmiInstance -Namespace root\subscription -Class CommandLineEventConsumer `
    -Arguments @{ Name = "LogNoJutsu_WMI_Consumer"; CommandLineTemplate = "cmd.exe /C ..." }

# Filter-Consumer-Binding (Sysmon Event 21)
Set-WmiInstance -Namespace root\subscription -Class __FilterToConsumerBinding `
    -Arguments @{ Filter = $wmiFilter; Consumer = $wmiConsumer }
```

**Cleanup:**

```powershell
# Binding, Consumer und Filter entfernen
Get-WmiObject -Namespace root\subscription -Class __FilterToConsumerBinding | Remove-WmiObject
Get-WmiObject -Namespace root\subscription -Class CommandLineEventConsumer | Remove-WmiObject
Get-WmiObject -Namespace root\subscription -Class __EventFilter | Remove-WmiObject
```

**Warum das Angreifer tun:** WMI Event Subscriptions sind eine der stealth-fähigsten Persistence-Mechanismen auf Windows. Sie existieren ausschließlich in der WMI-Datenbank, nicht als Dateien oder Registry-Einträge. APTs wie APT29 (Cozy Bear) nutzen diese Technik intensiv. Sysmon Events 19/20/21 sind die einzige zuverlässige Erkennungsquelle.

**Erwartete SIEM-Events:**
- `Sysmon 19` — WmiEvent (EventFilter created)
- `Sysmon 20` — WmiEvent (EventConsumer created)
- `Sysmon 21` — WmiEvent (Filter-Consumer-Binding) — **Trifecta = Sicherer IOC**

---

#### T1021.001 — Remote Desktop Protocol

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1021.001](https://attack.mitre.org/techniques/T1021/001/) |
| Taktik | Lateral Movement |
| Exabeam-Regeln | 6 |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# RDP-Status prüfen (Registry-Lese-Zugriff)
Get-ItemProperty "HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server" -Name "fDenyTSConnections"

# Aktive RDP-Sessions abfragen
query session

# 5 fehlgeschlagene RDP-Authentifizierungsversuche
1..5 | ForEach-Object {
    $ctx.ValidateCredentials("rdp_target_$_", "WrongRDPPass$_!")
    Start-Sleep -Milliseconds 500
}
```

**Warum das Angreifer tun:** RDP ist der am häufigsten für Lateral Movement genutzte Windows-Protokoll. Event 4624 Typ 10 (Remote Interactive) ist der spezifische Logon-Typ für RDP. Fehlgeschlagene RDP-Versuche (4625) in Kombination mit dem Quell-Host sind ein Schlüsselindikator für RDP-Brute-Force.

**Erwartete SIEM-Events:**
- `4625` × 5 — Fehlgeschlagene Anmeldungen mit RDP-Kontext
- `4624` Typ 10 — Remote Interactive Logon (echte RDP-Sessions)
- `Sysmon 1` — `query.exe` Prozesserstellung

---

#### T1036.005 — Process Masquerading

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1036.005](https://attack.mitre.org/techniques/T1036/005/) |
| Taktik | Defense Evasion |
| Exabeam-Regeln | 27 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — Temporäre Dateien werden entfernt |

**Was wird angelegt:**

```
%TEMP%\LogNoJutsu_Masq\svchost.exe       (Kopie von cmd.exe)
%TEMP%\LogNoJutsu_Masq\explorer.exe      (Kopie von powershell.exe)
%TEMP%\LogNoJutsu_Masq\invoice_Q4.pdf.exe (Double-Extension)
```

**Was wird ausgeführt:**

```powershell
# cmd.exe als svchost.exe aus Temp-Verzeichnis ausführen
Copy-Item "C:\Windows\System32\cmd.exe" -Destination "$tempDir\svchost.exe"
Start-Process "$tempDir\svchost.exe" -ArgumentList "/C echo Masquerade-svchost"

# PowerShell als explorer.exe
Copy-Item "powershell.exe" -Destination "$tempDir\explorer.exe"
Start-Process "$tempDir\explorer.exe" -ArgumentList "-Command ..."

# Double-Extension Datei
Copy-Item "cmd.exe" -Destination "$tempDir\invoice_Q4_2024.pdf.exe"
```

**Cleanup:**

```powershell
Remove-Item "$env:TEMP\LogNoJutsu_Masq" -Recurse -Force
```

**Warum das Angreifer tun:** Prozess-Masquerading täuscht Security-Tools und SOC-Analysten durch vertraute Prozessnamen. Exabeam und andere SIEM-Lösungen erkennen dies durch Vergleich des Prozessnamens mit dem Ausführungspfad. `svchost.exe` außerhalb von `C:\Windows\System32\` ist ein sofortiger IOC.

**Erwartete SIEM-Events:**
- `Sysmon 1` — Prozess mit bekanntem Windows-Namen aus unbekanntem Pfad
- `Sysmon 11` — FileCreate für kopierte Binärdateien
- `4688` — Prozesserstellung mit Pfad-/Name-Mismatch

---

#### T1486 — Data Encrypted for Impact (Ransomware Simulation)

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1486](https://attack.mitre.org/techniques/T1486/) |
| Taktik | Impact |
| Exabeam-Regeln | 3 |
| Admin erforderlich | Nein |
| Cleanup | **Ja** — Simulation-Verzeichnis vollständig entfernt |

**Was wird angelegt:**

```
%TEMP%\LogNoJutsu_RansomSim\
  ├── document.docx.lnj_encrypted
  ├── spreadsheet.xlsx.lnj_encrypted
  ├── presentation.pptx.lnj_encrypted
  ├── ...  (10 Dateien)
  └── README_RESTORE_FILES.txt
```

**Was wird ausgeführt:**

```powershell
# Testdateien erstellen (Sysmon 11 — Mass FileCreate)
foreach ($name in $fileNames) {
    Set-Content -Path "$targetDir\$name" -Value "Test file content"
}

# Dateien mit fake-verschlüsselter Erweiterung umbenennen
Get-ChildItem $targetDir | ForEach-Object {
    Rename-Item $_.FullName -NewName ($_.Name + ".lnj_encrypted")
}

# Ransom-Note erstellen (Sysmon 11 — README_RESTORE_FILES.txt)
Set-Content -Path "$targetDir\README_RESTORE_FILES.txt" -Value $note
```

**Cleanup:**

```powershell
Remove-Item "$env:TEMP\LogNoJutsu_RansomSim" -Recurse -Force
```

**Warum das Angreifer tun:** Ransomware benennt verschlüsselte Dateien mit einer neuen Erweiterung um und erstellt eine Ransom-Note. Das Muster aus massenhafter Datei-Umbenennung + unbekannte Erweiterung + README-Datei-Erstellung ist das primäre Verhaltenssignal für Ransomware-Erkennung in SIEM/EDR. **Keine echte Verschlüsselung findet statt.**

**Erwartete SIEM-Events:**
- `Sysmon 11` — Massen-FileCreate und Datei-Rename mit unbekannter Extension
- `4688` — `powershell.exe` Prozesserstellung

---

#### T1490 — Inhibit System Recovery

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1490](https://attack.mitre.org/techniques/T1490/) |
| Taktik | Impact |
| Exabeam-Regeln | 6 |
| Admin erforderlich | **Ja** |
| Cleanup | Keiner — Shadow Copies werden **nicht** gelöscht |

**Was wird ausgeführt:**

```cmd
# Shadow Copies AUFLISTEN (nicht löschen)
vssadmin list shadows
wmic shadowcopy list brief
```

Das Tool zeigt informatorisch, welche Befehle ein Angreifer ausführen würde:
```
vssadmin delete shadows /all /quiet       ← WIRD NICHT AUSGEFÜHRT
wmic shadowcopy delete                    ← WIRD NICHT AUSGEFÜHRT
bcdedit /set {default} recoveryenabled no ← WIRD NICHT AUSGEFÜHRT
```

**Warum das Angreifer tun:** Das Löschen von Volume Shadow Copies (VSS) und Deaktivierung der Windows-Wiederherstellung verhindert, dass Opfer ihre Daten wiederherstellen können — klassischer Ransomware-Vorläufer. `vssadmin.exe` und `wmic.exe` mit Shadow-Copy-Argumenten sind in Exabeam als "Disable Windows recovery mode" Correlation Rule mit Severity 2 hinterlegt.

**Erwartete SIEM-Events:**
- `4688` — `vssadmin.exe` und `wmic.exe` mit Shadow-Copy-Argumenten
- `Sysmon 1` — Prozesserstellung mit vollständigen Argumenten

---

#### T1562.002 — Disable Windows Event Logging

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1562.002](https://attack.mitre.org/techniques/T1562/002/) |
| Taktik | Defense Evasion |
| Exabeam-Regeln | 3 |
| Admin erforderlich | **Ja** |
| Cleanup | **Ja** — Auditierung wird wiederhergestellt |

**Was wird ausgeführt:**

```cmd
# Aktuelle Logon-Audit-Policy abfragen
auditpol /get /subcategory:"Logon" /r

# Logon-Auditierung temporär deaktivieren (wird sofort in Cleanup wiederhergestellt)
auditpol /set /subcategory:"Logon" /success:disable /failure:disable
```

**Cleanup:**

```cmd
auditpol /set /subcategory:"Logon" /success:enable /failure:enable
```

**Warum das Angreifer tun:** Durch Deaktivierung der Auditierung werden keine Event-4624/4625-Einträge für nachfolgende Aktionen generiert. Event 4719 ("System audit policy was changed") ist das Erkennungssignal — ironischerweise kann ein Angreifer dieses Signal durch schnelle Aktivitäten zwischen Policy-Änderung und Wiederherstellung minimieren.

**Erwartete SIEM-Events:**
- `4719` — "System audit policy was changed" — **das Erkennungssignal**

---

#### T1070.001 — Clear Windows Event Logs

| Eigenschaft | Wert |
|---|---|
| MITRE ATT&CK | [T1070.001](https://attack.mitre.org/techniques/T1070/001/) |
| Taktik | Defense Evasion |
| Exabeam-Regeln | 8 |
| Admin erforderlich | **Ja** |
| Cleanup | Keiner erforderlich |

**Was wird ausgeführt:**

```cmd
# Nur das Application-Log wird geleert (kein Security-Log)
# Ein echter Angreifer würde wevtutil cl Security ausführen
wevtutil cl Application
```

**Warum das Angreifer tun:** Nach einer Kompromittierung versuchen Angreifer, Spuren zu verwischen, indem sie Event-Logs löschen. Event 1102 (Audit Log Cleared) und `wevtutil cl` mit dem Security-Log-Argument sind direkte IOCs. **Das Security-Log wird von LogNoJutsu bewusst nicht gelöscht**, um die forensische Integrität der Simulation zu bewahren.

**Erwartete SIEM-Events:**
- `4688` — `wevtutil.exe` mit `cl` Argument
- `Sysmon 1` — `wevtutil.exe cl Application`
- `1102` — Audit log cleared (nur wenn Security-Log geleert)

---

### UEBA-Szenarien (Exabeam)

Diese Szenarien sind speziell für die Validierung von **Exabeam UEBA-Use-Cases** konzipiert. UEBA (User and Entity Behavior Analytics) erkennt keine einzelnen Events, sondern **Verhaltensmuster über Zeit**. Exabeam nutzt 750+ vortrainierte Verhaltensmodelle (Kategorial, Numerisch-Clustered, Zeitbasiert), die pro Benutzer, Peer-Group und Organisation kalibriert werden.

---

#### UEBA-SPRAY-CHAIN — Credential Spray → Success Chain

| Eigenschaft | Wert |
|---|---|
| UEBA Use Case | Brute Force / Credential Stuffing (Exabeam Package: Compromised Insiders) |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# 25 fehlgeschlagene Auth-Versuche in schneller Folge (100ms Abstand)
for ($i = 1; $i -le 25; $i++) {
    $ctx.ValidateCredentials("spray_victim_user", "WrongPass$i!")
    Start-Sleep -Milliseconds 100
}
```

**UEBA-Erkennungslogik:** Exabeam erkennt einen deutlichen Anstieg von 4625-Events innerhalb eines kurzen Zeitfensters. Die Kombination aus hohem Volumen (>10 Versuche in 60 Sekunden) und konstantem Quell-Host triggert den "Brute Force" Use Case. Die Schwelle für VPN-Brute-Force ist bei Exabeam auf 10+ fehlgeschlagene Logins pro Minute dokumentiert.

**Erwartete SIEM-Events:**
- `4625` × 25 in ~3 Sekunden
- Exabeam: Brute Force / Credential Stuffing Use Case

---

#### UEBA-OFFHOURS — Off-Hours Activity Simulation

| Eigenschaft | Wert |
|---|---|
| UEBA Use Case | Abnormal Activity Time (Exabeam: Compromised Credentials & Abnormal Auth) |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

```powershell
# Standard-Recon-Aktivitäten — Content ist zweitrangig, Zeitstempel entscheidet
whoami /all; net user; ipconfig /all
Get-Process | Select-Object -First 10
Get-ChildItem $env:USERPROFILE
```

**UEBA-Erkennungslogik:** Exabeam nutzt ein **Numerical Time-of-Week Model** — es lernt die normalen Arbeitszeiten eines Benutzers aus historischen 4624-Events (typisch 08:00–18:00 Mo–Fr). Aktivität außerhalb dieser Baseline erhöht den Session-Risikowert. Das Modell behandelt Sonntag 23:00 als ähnlich zu Montag 00:00 (zyklisches Zeitmodell).

> **Hinweis:** Für maximale Erkennungswirkung außerhalb der regulären Geschäftszeiten ausführen. Das Tool gibt die aktuelle Uhrzeit im Log aus.

**Erwartete SIEM-Events:**
- `4688` / `Sysmon 1` — Prozesse außerhalb der Baseline-Arbeitszeiten
- Exabeam: "Abnormal activity time for user"

---

#### UEBA-LATERAL-CHAIN — Lateral Movement Discovery Chain

| Eigenschaft | Wert |
|---|---|
| UEBA Use Case | Reconnaissance / First-Time-Seen Behavior (Exabeam: Lateral Movement) |
| Admin erforderlich | Nein |
| Cleanup | Keiner |

**Was wird ausgeführt:**

12 Enumeration-Befehle in schneller Abfolge (~300ms Abstand):

```powershell
whoami /groups          # Benutzer-Gruppen
net user                # Lokale Benutzer
net localgroup administrators
net config workstation  # Domain-Info
ipconfig /all           # Netzwerkkonfiguration
netstat -ano            # Aktive Verbindungen
arp -a                  # ARP-Cache (Nachbar-Hosts)
route print             # Routing-Tabelle
ipconfig /displaydns    # DNS-Cache
net share               # Freigegebene Ressourcen
net session             # Aktive Sessions
tasklist /v             # Laufende Prozesse
```

**UEBA-Erkennungslogik:** Exabeam erkennt dieses Muster auf zwei Ebenen:
1. **First-Time-Seen (Categorical Model):** Wenn ein Benutzer erstmalig `netstat.exe`, `arp.exe` oder `net.exe` ausführt, erhöht das den Risikowert deutlich
2. **Volume Anomaly (Numerical Clustered Model):** 12 Netzwerk-/System-Abfragen in ~4 Sekunden ist weit außerhalb normaler Benutzeraktivität

**Erwartete SIEM-Events:**
- `4688` / `Sysmon 1` × 12 — Schnelle Prozessabfolge von `net.exe`, `netstat.exe`, `arp.exe`, `ipconfig.exe`
- Exabeam: "First time user executed network reconnaissance commands"

---

## Kampagnen / Playbooks

Kampagnen sind geordnete Abfolgen von Techniken, die reale Angreifer-TTPs nachbilden. Sie werden in der Web-UI im Tab "Playbooks" ausgewählt.

### Übersicht

| Kampagne | Kategorie | Bedrohungsakteur | Schritte |
|---|---|---|---|
| `finance-fin7` | Branche | FIN7 / Carbanak | 10 |
| `healthcare-ransomware` | Branche | Conti / LockBit | 10 |
| `manufacturing-apt` | Branche | Sandworm / TEMP.Veles | 11 |
| `energy-longdwell` | Branche | Volt Typhoon / Dragonfly | 10 |
| `retail-pos` | Branche | FIN7 POS-Variante | 10 |
| `government-apt` | Branche | APT29 / APT28 | 14 |
| `ueba-exabeam-validation` | UEBA | Generisch | 10 |
| `lateral-movement-credential-theft` | Exabeam Validation | APT Lateral Movement | 9 |
| `account-manipulation-persistence` | Exabeam Validation | APT Post-Exploitation | 7 |
| `defense-evasion-lolbin` | Exabeam Validation | Sophisticated APT | 8 |
| `ransomware-full-chain` | Exabeam Validation | LockBit / Conti TTPs | 10 |
| `insider-threat` | Exabeam Validation | Malicious Insider | 9 |

---

### Branchen-Kampagnen

#### finance-fin7 — Finance / FIN7 Carbanak
Simuliert die TTPs der FIN7/Carbanak-Gruppe, die auf Finanzinstitute spezialisiert ist. Umfasst Spear-Phishing-Nachfolgeaktionen, PowerShell-Backdoors, Credential Dumping und laterale Bewegung zu Payment-Systemen.

Schritte: T1082 → T1087 → T1057 → T1059.001 → T1003.001 → T1021.001 → T1053.005 → T1547.001 → T1562.002 → T1070.001

#### healthcare-ransomware — Healthcare / Conti LockBit
Simuliert Ransomware-Angriffe auf Healthcare-Umgebungen (häufig da Patientendaten kritisch und oft schlecht geschützt). Fokus auf Discovery, Credential-Harvesting, Deaktivierung von Backups und Simulation der Verschlüsselungsphase.

Schritte: T1082 → T1083 → T1087 → T1003.001 → T1059.001 → T1547.001 → T1562.002 → T1490 → T1486 → T1070.001

#### manufacturing-apt — Manufacturing / Sandworm
Simuliert APT-Angriffe auf Fertigungsumgebungen (ICS/SCADA-Kontext). Langer Verweildauer (Long Dwell), Service-Installation für Persistence, gezielte System-Discovery.

Schritte: T1082 → T1016 → T1049 → T1057 → T1083 → T1087 → T1059.001 → T1543.003 → T1562.002 → T1070.001 → T1490

#### energy-longdwell — Energy / Volt Typhoon Dragonfly
Simuliert Long-Dwell-APT-Angriffe auf Energieunternehmen (Kritische Infrastruktur). Charakteristisch: Slow Recon, minimale Footprint-Hinterlasung, Living-off-the-Land.

Schritte: T1082 → T1049 → T1016 → T1057 → T1087 → T1083 → T1059.001 → T1547.001 → T1562.002 → T1070.001

#### retail-pos — Retail / FIN7 POS
Simuliert POS-System-Angriffe im Einzelhandel. Fokus auf Credential-Harvesting, Lateral Movement zu POS-Terminals und Persistence.

Schritte: T1082 → T1087 → T1057 → T1049 → T1059.001 → T1003.001 → T1547.001 → T1053.005 → T1562.002 → T1070.001

#### government-apt — Government / APT29 APT28
Simuliert nation-state APT-Angriffe auf Regierungsbehörden (14 Schritte — umfangreichste Branchen-Kampagne). Umfasst vollständige Kill-Chain von Discovery über Credential-Access bis zu Persistence und Defense-Evasion.

Schritte: T1082 → T1087 → T1069 → T1049 → T1016 → T1057 → T1083 → T1059.001 → T1003.001 → T1547.001 → T1053.005 → T1543.003 → T1562.002 → T1070.001

---

### Exabeam Validation Campaigns

Diese vier Kampagnen sind direkt auf die Exabeam Use Case Library ausgerichtet und testen gezielt die höchsten Regelabdeckungen.

#### ueba-exabeam-validation — Exabeam UEBA Validation Suite
Strukturierte Validierung aller wichtigen Exabeam UEBA-Use-Cases nach dem Onboarding:

```
T1082              → Baseline-Aktivität erzeugen
T1016              → Baseline-Aktivität erzeugen
UEBA-OFFHOURS      → Use Case: Abnormal Activity Time
UEBA-LATERAL-CHAIN → Use Case: First-Time Recon Behavior
UEBA-SPRAY-CHAIN   → Use Case: Brute Force / Credential Stuffing
T1053.005          → Anomalie: Neuer Scheduled Task
T1547.001          → Anomalie: Neuer Registry Run Key
T1003.001          → Use Case: Credential Dumping Precursor
T1562.002          → Use Case: Defense Evasion
T1070.001          → Use Case: Log Clearing
```

#### lateral-movement-credential-theft — Lateral Movement & Credential Theft Chain
Deckt Exabeam's Lateral Movement Use Case (118 Regeln für T1021) und Credential Access (49 Regeln für T1003) ab. Simuliert den vollständigen Pfad von initialer Enumeration über Kerberoasting bis zum DCSync-Recon.

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
Deckt Exabeam's Account Manipulation Use Case direkt ab (T1098 = 57 Regeln, T1136 = 35 Regeln). Validiert alle wichtigen Persistence-Mechanismen in einer Kampagne.

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
Deckt Exabeam's Evasion Use Case ab. Fokus auf Living-off-the-Land-Binaries (LOLBins) und Obfuskationstechniken. T1218 hat 116 Exabeam-Regeln — die zweithöchste Abdeckung nach T1078.

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
Simuliert eine vollständige Ransomware Kill-Chain von Discovery bis zur Simulation der Verschlüsselungsphase. Deckt alle dokumentierten Exabeam Ransomware-Correlation-Rules ab (bcdedit, vssadmin, Massenumbenennung).

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
Simuliert einen böswilligen Insider (z.B. ausscheidenden Mitarbeiter) bei Daten-Exfiltrations-Vorbereitung. Richtet sich an Exabeam's Malicious Insider Package. **Empfohlen: Mit einem dedizierten Benutzerprofil ausführen (User Rotation), um authentische UEBA-Verhaltensanomalien zu erzeugen.**

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

## Logging und Reporting

### Simulations-Log (`.log`)

Für jede Simulation wird eine Textdatei im Format `lognojutsu_YYYYMMDD_HHMMSS_[Kampagne].log` im selben Verzeichnis wie die `.exe` erstellt.

| Log-Typ | Beschreibung |
|---|---|
| `SIM_START` | Simulationsstart mit Konfigurationsdetails |
| `PHASE` | Phasenwechsel (Discovery / Attack / Cleanup) |
| `TECH_START` | Beginn einer Technik (ID, Name, Taktik) |
| `COMMAND` | Ausgeführter Befehl (Executor-Typ + Kommandozeile) |
| `OUTPUT` | Komplette stdout-Ausgabe der Technik |
| `ERROR` | stderr-Ausgabe (falls vorhanden) |
| `CLEANUP` | Ausgeführter Cleanup-Befehl und Ergebnis |
| `TECH_END` | Abschluss der Technik mit Dauer und Erfolgsstatus |
| `SIM_END` | Zusammenfassung (X/Y Techniken erfolgreich) |

**Beispiel-Log-Ausschnitt:**
```
[2026-03-21 22:00:00.000] [SIM_START   ] === LogNoJutsu Simulation Started ===
[2026-03-21 22:00:00.001] [INFO        ] Configuration: rotation=sequential profiles=2
[2026-03-21 22:00:00.002] [PHASE       ] ▶ PHASE: DISCOVERY
[2026-03-21 22:00:00.003] [TECH_START  ] [START] T1082 — System Information Discovery (as: CORP\jsmith)
[2026-03-21 22:00:00.004] [COMMAND     ]   Executor: powershell
[2026-03-21 22:00:01.200] [OUTPUT      ]   OS Name: Microsoft Windows 11 Pro
[2026-03-21 22:00:01.210] [TECH_END    ] [END] T1082 — SUCCESS ✓ (1.207s)
```

### JSON-Report (`.json`)

Zusätzlich wird `lognojutsu_report_YYYYMMDD_HHMMSS.json` erstellt:
- Gesamtstatistik: Gesamt / Erfolgreich / Fehlgeschlagen
- Pro Technik: ID, Name, Taktik, Start/Ende, Success, Stdout, Stderr, Cleanup-Status, ausführender Benutzer (`run_as_user`)

---

## Cleanup-Mechanismus

### Techniken mit Cleanup

| Technik | Angelegtes Artefakt | Cleanup |
|---|---|---|
| T1136.001 | Lokaler Account `lnj_test_acct` | `net user lnj_test_acct /delete` |
| T1098 | Account in Administrators-Gruppe | Gruppen-Mitgliedschaft + Account entfernt |
| T1547.001 | Registry-Key `HKCU\...\Run\LogNoJutsu_Persistence_Test` | Registry-Eintrag gelöscht |
| T1053.005 | Scheduled Task `LogNoJutsu_Task_Test` | Task deregistriert |
| T1543.003 | Windows Service `LogNoJutsuTestSvc` | Service gelöscht |
| T1548.002 | Registry-Key `HKCU\Software\Classes\mscfile\...` | Registry-Tree entfernt |
| T1197 | BITS-Job `LogNoJutsu_BITS_Test` | Job abgebrochen |
| T1546.003 | WMI Filter, Consumer, Binding | Alle drei WMI-Objekte entfernt |
| T1036.005 | Temporäre Binärdateien in `%TEMP%\LogNoJutsu_Masq\` | Verzeichnis entfernt |
| T1486 | Simuliertes Ransomware-Verzeichnis `%TEMP%\LogNoJutsu_RansomSim\` | Verzeichnis entfernt |
| T1562.002 | Logon-Auditierung deaktiviert | Auditierung wiederhergestellt |
| T1550.002 | Netzwerkverbindungen via `net use` | `net use * /delete` |

### Cleanup-Modi

**Modus 1 — Per-Technik-Cleanup (Standard, Checkbox aktiv):**
Nach jeder Technik wird sofort der zugehörige Cleanup ausgeführt. Artefakte existieren nur während der Technik-Ausführung.

**Modus 2 — End-of-Simulation-Cleanup (Checkbox inaktiv):**
Alle Artefakte bleiben während der gesamten Simulation bestehen und werden gesammelt am Ende entfernt. Sinnvoll, wenn das SIEM auch die Persistenz-Erkennung über Zeit testen soll.

**Modus 3 — Abbruch-Cleanup (Stop & Cleanup Button):**
Wird die Simulation manuell abgebrochen, wird automatisch ein vollständiger Cleanup aller bis dahin ausgeführten Techniken durchgeführt.

---

## Kommandozeilen-Optionen

```
lognojutsu.exe [Optionen]

Optionen:
  -host string
        Bind-Adresse für den HTTP-Server (Standard: "127.0.0.1")
        Für Netzwerkzugriff: 0.0.0.0

  -port int
        HTTP-Port (Standard: 8080)

  -password string
        Optionales Passwort für die Web-UI (HTTP Basic Auth)
        Leer lassen = keine Authentifizierung

Beispiele:
  lognojutsu.exe
        Startet mit Standard-Einstellungen (nur localhost, Port 8080)

  lognojutsu.exe -host 0.0.0.0 -port 9090
        Erreichbar im gesamten Netzwerk auf Port 9090

  lognojutsu.exe -host 0.0.0.0 -password "Simulation2026!"
        Netzwerkzugriff mit Passwortschutz
```

---

*LogNoJutsu ist ein Werkzeug ausschließlich für autorisierte SIEM-Validierung in kontrollierten Testumgebungen. Der Einsatz auf Systemen ohne ausdrückliche Genehmigung ist unzulässig.*
