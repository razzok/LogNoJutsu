package preparation

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result holds the outcome of a preparation step.
type Result struct {
	Step    string `json:"step"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RunAll executes all preparation steps and returns results.
func RunAll() []Result {
	steps := []struct {
		name string
		fn   func() Result
	}{
		{"PowerShell ScriptBlock Logging", EnablePowerShellLogging},
		{"Windows Audit Policy", ConfigureAuditPolicy},
		{"Sysmon Installation", InstallSysmon},
	}

	var results []Result
	for _, s := range steps {
		log.Printf("[Preparation] Running: %s", s.name)
		r := s.fn()
		results = append(results, r)
		log.Printf("[Preparation] %s: success=%v | %s", s.name, r.Success, r.Message)
	}
	return results
}

// EnablePowerShellLogging enables PowerShell ScriptBlock and Module logging via registry.
func EnablePowerShellLogging() Result {
	commands := []string{
		// ScriptBlock Logging (Event 4104)
		`New-Item -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging" -Force | Out-Null`,
		`Set-ItemProperty -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging" -Name "EnableScriptBlockLogging" -Value 1 -Type DWord -Force`,
		// Module Logging (Event 4103)
		`New-Item -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ModuleLogging" -Force | Out-Null`,
		`Set-ItemProperty -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ModuleLogging" -Name "EnableModuleLogging" -Value 1 -Type DWord -Force`,
		// Transcription logging
		`New-Item -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\Transcription" -Force | Out-Null`,
		`Set-ItemProperty -Path "HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\Transcription" -Name "EnableTranscripting" -Value 1 -Type DWord -Force`,
	}

	script := strings.Join(commands, "\n")
	out, err := runPS(script)
	if err != nil {
		return Result{Step: "PowerShell ScriptBlock Logging", Success: false,
			Message: fmt.Sprintf("Failed: %v | %s", err, out)}
	}
	return Result{Step: "PowerShell ScriptBlock Logging", Success: true,
		Message: "ScriptBlock logging (4104), Module logging (4103), and Transcription enabled"}
}

// auditPolicies lists Windows audit subcategories to enable, identified by stable GUIDs.
// GUIDs are locale-independent — English subcategory names fail on non-English Windows.
// Reference: https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/
var auditPolicies = []struct {
	guid        string
	description string
}{
	{"{0CCE9215-69AE-11D9-BED3-505054503030}", "Logon/Logoff events (4624, 4625, 4634)"},
	{"{0CCE9216-69AE-11D9-BED3-505054503030}", "Logoff events"},
	{"{0CCE9217-69AE-11D9-BED3-505054503030}", "Account lockout events (4740)"},
	{"{0CCE922B-69AE-11D9-BED3-505054503030}", "Process creation events (4688)"},
	{"{0CCE922F-69AE-11D9-BED3-505054503030}", "Audit policy changes (4719)"},           // VERIFY: run auditpol /list /subcategory:* /v — GUID disputed between research sources
	{"{0CCE9237-69AE-11D9-BED3-505054503030}", "Group changes (4728, 4732)"},
	{"{0CCE9235-69AE-11D9-BED3-505054503030}", "Account creation/changes (4720)"},
	{"{0CCE9228-69AE-11D9-BED3-505054503030}", "Privilege use (4673, 4674)"},
	{"{0CCE921B-69AE-11D9-BED3-505054503030}", "Special privilege logons (4672)"},
	{"{0CCE9227-69AE-11D9-BED3-505054503030}", "Object access / Scheduled task events (4698)"}, // VERIFY: run auditpol /list /subcategory:* /v — GUID disputed between research sources. Deduplicated: covers both "Other Object Access Events" and "Scheduled Task"
	{"{0CCE9226-69AE-11D9-BED3-505054503030}", "Network connection events"},
}

// ConfigureAuditPolicy sets Windows audit policy for SIEM-relevant events.
func ConfigureAuditPolicy() Result {
	var failures []string
	for _, p := range auditPolicies {
		cmd := exec.Command("auditpol.exe", "/set",
			"/subcategory:"+p.guid,
			"/success:enable",
			"/failure:enable")
		if _, err := cmd.CombinedOutput(); err != nil {
			failures = append(failures, fmt.Sprintf("%s: failed (%v)", p.description, err))
		}
	}

	// Enable command line logging in process creation (Event 4688 with command line)
	_, _ = runPS(`Set-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System\Audit" -Name "ProcessCreationIncludeCmdLine_Enabled" -Value 1 -Type DWord -Force -ErrorAction SilentlyContinue`)

	if len(failures) > 0 {
		return Result{Step: "Windows Audit Policy", Success: false,
			Message: fmt.Sprintf("Partial failure: %s", strings.Join(failures, "; "))}
	}
	return Result{Step: "Windows Audit Policy", Success: true,
		Message: fmt.Sprintf("Configured %d audit subcategories + command line logging in 4688", len(auditPolicies))}
}

// InstallSysmon downloads and installs Sysmon with a recommended config.
func InstallSysmon() Result {
	// Check if already installed
	checkCmd := exec.Command("sc.exe", "query", "Sysmon64")
	if err := checkCmd.Run(); err == nil {
		return Result{Step: "Sysmon Installation", Success: true,
			Message: "Sysmon64 already installed and running"}
	}
	checkCmd = exec.Command("sc.exe", "query", "Sysmon")
	if err := checkCmd.Run(); err == nil {
		return Result{Step: "Sysmon Installation", Success: true,
			Message: "Sysmon already installed and running"}
	}

	// Write Sysmon config
	sysmonConfig := sysmonConfigXML()
	configPath := filepath.Join(os.TempDir(), "lognojutsu_sysmon.xml")
	if err := os.WriteFile(configPath, []byte(sysmonConfig), 0644); err != nil {
		return Result{Step: "Sysmon Installation", Success: false,
			Message: fmt.Sprintf("Failed to write Sysmon config: %v", err)}
	}

	// Attempt download via PowerShell
	sysmonPath := filepath.Join(os.TempDir(), "Sysmon64.exe")
	downloadScript := fmt.Sprintf(`
$url = "https://download.sysinternals.com/files/Sysmon.zip"
$zipPath = "%s"
$extractPath = "%s"
try {
    Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing -TimeoutSec 30
    Expand-Archive -Path $zipPath -DestinationPath $extractPath -Force
    Write-Host "Downloaded"
} catch {
    Write-Host "DOWNLOAD_FAILED: $_"
}`,
		filepath.Join(os.TempDir(), "Sysmon.zip"),
		os.TempDir(),
	)

	out, _ := runPS(downloadScript)
	if strings.Contains(out, "DOWNLOAD_FAILED") || !fileExists(sysmonPath) {
		return Result{Step: "Sysmon Installation", Success: false,
			Message: "Could not download Sysmon automatically. Please download Sysmon64.exe from https://docs.microsoft.com/sysinternals/downloads/sysmon and place it next to lognojutsu.exe, then click Install Sysmon again."}
	}

	// Install Sysmon
	installCmd := exec.Command(sysmonPath, "-accepteula", "-i", configPath)
	installOut, err := installCmd.CombinedOutput()
	if err != nil {
		return Result{Step: "Sysmon Installation", Success: false,
			Message: fmt.Sprintf("Installation failed: %v | %s", err, string(installOut))}
	}

	return Result{Step: "Sysmon Installation", Success: true,
		Message: "Sysmon64 installed with LogNoJutsu recommended configuration"}
}

func runPS(script string) (string, error) {
	cmd := exec.Command("powershell.exe",
		"-NonInteractive", "-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", script)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// sysmonConfigXML returns a comprehensive Sysmon config for SIEM validation.
func sysmonConfigXML() string {
	return `<Sysmon schemaversion="4.90">
  <HashAlgorithms>md5,sha256,IMPHASH</HashAlgorithms>
  <CheckRevocation/>
  <EventFiltering>

    <!-- Event ID 1: Process creation -->
    <RuleGroup name="" groupRelation="or">
      <ProcessCreate onmatch="exclude">
        <Image condition="is">C:\Windows\System32\wermgr.exe</Image>
      </ProcessCreate>
    </RuleGroup>

    <!-- Event ID 3: Network connection -->
    <RuleGroup name="" groupRelation="or">
      <NetworkConnect onmatch="exclude">
        <Image condition="is">C:\Windows\System32\svchost.exe</Image>
        <DestinationPort condition="is">443</DestinationPort>
      </NetworkConnect>
    </RuleGroup>

    <!-- Event ID 7: Image loaded -->
    <RuleGroup name="" groupRelation="or">
      <ImageLoad onmatch="include">
        <ImageLoaded condition="contains">clrjit.dll</ImageLoaded>
        <ImageLoaded condition="contains">clr.dll</ImageLoaded>
      </ImageLoad>
    </RuleGroup>

    <!-- Event ID 8: CreateRemoteThread -->
    <RuleGroup name="" groupRelation="or">
      <CreateRemoteThread onmatch="exclude">
        <SourceImage condition="is">C:\Windows\System32\csrss.exe</SourceImage>
      </CreateRemoteThread>
    </RuleGroup>

    <!-- Event ID 10: ProcessAccess (LSASS access detection) -->
    <RuleGroup name="" groupRelation="or">
      <ProcessAccess onmatch="include">
        <TargetImage condition="is">C:\Windows\system32\lsass.exe</TargetImage>
      </ProcessAccess>
    </RuleGroup>

    <!-- Event ID 11: FileCreate -->
    <RuleGroup name="" groupRelation="or">
      <FileCreate onmatch="include">
        <TargetFilename condition="contains">\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup</TargetFilename>
        <TargetFilename condition="end with">.ps1</TargetFilename>
        <TargetFilename condition="end with">.bat</TargetFilename>
        <TargetFilename condition="end with">.vbs</TargetFilename>
      </FileCreate>
    </RuleGroup>

    <!-- Event ID 12/13: Registry events -->
    <RuleGroup name="" groupRelation="or">
      <RegistryEvent onmatch="include">
        <TargetObject condition="contains">CurrentVersion\Run</TargetObject>
        <TargetObject condition="contains">CurrentVersion\RunOnce</TargetObject>
        <TargetObject condition="contains">Winlogon</TargetObject>
        <TargetObject condition="contains">Image File Execution Options</TargetObject>
      </RegistryEvent>
    </RuleGroup>

    <!-- Event ID 22: DNS query -->
    <RuleGroup name="" groupRelation="or">
      <DnsQuery onmatch="exclude">
        <QueryName condition="end with">.microsoft.com</QueryName>
        <QueryName condition="end with">.windows.com</QueryName>
      </DnsQuery>
    </RuleGroup>

  </EventFiltering>
</Sysmon>`
}
