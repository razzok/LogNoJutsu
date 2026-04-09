package executor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lognojutsu/internal/native"
	"lognojutsu/internal/playbooks"
	"lognojutsu/internal/simlog"
	"lognojutsu/internal/userstore"
)

// Run executes a technique as the current user.
func Run(t *playbooks.Technique) playbooks.ExecutionResult {
	return runInternal(t, nil, "")
}

// RunAs executes a technique in the security context of a different user.
// The profile provides username/domain; password is the already-decrypted plaintext.
func RunAs(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
	return runInternal(t, profile, password)
}

// RunWithCleanup executes a technique (optionally as another user) and runs cleanup.
// Cleanup is registered via defer — it fires even if runInternal panics.
func RunWithCleanup(t *playbooks.Technique, profile *userstore.UserProfile, password string) (result playbooks.ExecutionResult) {
	// Register cleanup as deferred action — fires even on panic
	if strings.ToLower(t.Executor.Type) == "go" {
		// Native Go technique: use registered CleanupFunc if present
		if cleanFn := native.LookupCleanup(t.ID); cleanFn != nil {
			defer func() {
				simlog.TechCleanup(t.ID, "(go native cleanup)", false)
				cleanErr := cleanFn()
				result.CleanupRun = true
				simlog.TechCleanup(t.ID, "(go cleanup completed)", cleanErr == nil)
			}()
		}
	} else if strings.TrimSpace(t.Cleanup) != "" {
		defer func() {
			simlog.TechCleanup(t.ID, t.Cleanup, false)
			// Cleanup always runs as the launching user (we own the artifacts)
			_, _, cleanErr := runCommand(t.Executor.Type, t.Cleanup)
			result.CleanupRun = true
			simlog.TechCleanup(t.ID, "(cleanup completed)", cleanErr == nil)
		}()
	}
	result = runInternal(t, profile, password)
	return result
}

// RunCleanupOnly runs only the cleanup command for a technique (after abort).
func RunCleanupOnly(t *playbooks.Technique) {
	if strings.ToLower(t.Executor.Type) == "go" {
		if cleanFn := native.LookupCleanup(t.ID); cleanFn != nil {
			simlog.TechCleanup(t.ID, "(go native cleanup)", false)
			cleanErr := cleanFn()
			simlog.TechCleanup(t.ID, "(go cleanup-only run)", cleanErr == nil)
		}
		return
	}
	if strings.TrimSpace(t.Cleanup) == "" {
		return
	}
	simlog.TechCleanup(t.ID, t.Cleanup, false)
	_, _, err := runCommand(t.Executor.Type, t.Cleanup)
	simlog.TechCleanup(t.ID, "(cleanup-only run)", err == nil)
}

// ── internal ──────────────────────────────────────────────────────────────────

func runInternal(t *playbooks.Technique, profile *userstore.UserProfile, password string) playbooks.ExecutionResult {
	start := time.Now()
	result := playbooks.ExecutionResult{
		TechniqueID:   t.ID,
		TechniqueName: t.Name,
		TacticID:      t.Tactic,
		StartTime:     start.Format(time.RFC3339),
		Tier:          t.Tier,
	}

	userLabel := "current user"
	if profile != nil && profile.UserType != userstore.UserTypeCurrent {
		userLabel = profile.QualifiedName()
	}

	simlog.TechStart(t.ID, t.Name, t.Tactic, t.ElevationRequired)
	if profile != nil && profile.UserType != userstore.UserTypeCurrent {
		simlog.Info(fmt.Sprintf("  Running as: %s (%s)", profile.QualifiedName(), profile.UserType))
	}
	simlog.TechCommand(t.ID, t.Executor.Type, t.Executor.Command)

	// Native Go technique dispatch — no child process spawned
	if strings.ToLower(t.Executor.Type) == "go" {
		if profile != nil && profile.UserType != userstore.UserTypeCurrent {
			simlog.Info(fmt.Sprintf("[%s] type:go does not support RunAs — executing as current user", t.ID))
		}
		fn := native.Lookup(t.ID)
		if fn == nil {
			result.ErrorOutput = fmt.Sprintf("no native Go function registered for %s", t.ID)
			result.Success = false
		} else {
			nr, nErr := fn()
			result.Output = nr.Output
			result.ErrorOutput = nr.ErrorOutput
			result.Success = nr.Success
			if nErr != nil {
				result.ErrorOutput = nErr.Error() + "\n" + result.ErrorOutput
				result.Success = false
			}
		}
		result.EndTime = time.Now().Format(time.RFC3339)
		result.RunAsUser = userLabel
		simlog.TechOutput(t.ID, result.Output, result.ErrorOutput)
		simlog.TechEnd(t.ID, result.Success, time.Since(start))
		return result
	}

	var out, errOut string
	var err error

	if profile == nil || profile.UserType == userstore.UserTypeCurrent || password == "" {
		out, errOut, err = runCommand(t.Executor.Type, t.Executor.Command)
	} else {
		out, errOut, err = runCommandAs(t.Executor.Type, t.Executor.Command, profile, password)
	}

	// AMSI detection — PowerShell only (per D-02)
	if strings.ToLower(t.Executor.Type) == "powershell" || strings.ToLower(t.Executor.Type) == "psh" {
		if isAMSIBlocked(errOut, err) {
			result.Output = out
			result.ErrorOutput = errOut
			result.Success = false
			result.EndTime = time.Now().Format(time.RFC3339)
			result.RunAsUser = userLabel
			result.VerificationStatus = playbooks.VerifAMSIBlocked
			simlog.Info(fmt.Sprintf("[AMSI] %s — blocked by antimalware scan", t.ID))
			simlog.TechOutput(t.ID, out, errOut)
			simlog.TechEnd(t.ID, false, time.Since(start))
			return result
		}
	}

	result.Output = out
	result.ErrorOutput = errOut
	result.Success = (err == nil)
	result.EndTime = time.Now().Format(time.RFC3339)
	result.RunAsUser = userLabel

	simlog.TechOutput(t.ID, out, errOut)
	simlog.TechEnd(t.ID, result.Success, time.Since(start))

	return result
}

// isAMSIBlocked returns true if the error output or exit code indicates that
// Windows Antimalware Scan Interface (AMSI) blocked the PowerShell script.
func isAMSIBlocked(errOut string, err error) bool {
	amsiPatterns := []string{
		"ScriptContainedMaliciousContent",
		"This script contains malicious content",
		"has been blocked by your antivirus software",
	}
	for _, p := range amsiPatterns {
		if strings.Contains(errOut, p) {
			return true
		}
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == -196608 {
			return true
		}
	}
	return false
}

// runCommandAs executes a command in the context of another user via
// Start-Process -Credential. Output is captured through a temporary file.
// This generates Windows Event 4648 (explicit credential use) — a key UEBA signal.
func runCommandAs(execType, command string, profile *userstore.UserProfile, password string) (stdout, stderr string, err error) {
	if strings.TrimSpace(command) == "" {
		return "", "", nil
	}

	// Write the actual command to a temp script file so we can pass it cleanly
	tmpDir := os.TempDir()
	scriptFile := filepath.Join(tmpDir, fmt.Sprintf("lnj_%d.ps1", time.Now().UnixNano()))
	outFile := filepath.Join(tmpDir, fmt.Sprintf("lnj_out_%d.txt", time.Now().UnixNano()))
	errFile := filepath.Join(tmpDir, fmt.Sprintf("lnj_err_%d.txt", time.Now().UnixNano()))

	defer os.Remove(scriptFile)
	defer os.Remove(outFile)
	defer os.Remove(errFile)

	// Determine the actual command based on executor type
	var innerCmd string
	switch strings.ToLower(execType) {
	case "cmd", "command_prompt":
		// Wrap cmd command in a PowerShell-invokable form
		innerCmd = fmt.Sprintf("cmd.exe /C %s", command)
	default:
		innerCmd = command
	}

	// Write inner script
	if err := os.WriteFile(scriptFile, []byte(innerCmd), 0600); err != nil {
		return "", "", fmt.Errorf("writing temp script: %w", err)
	}

	// Build the encoded inner script path for safe passing
	encodedScriptPath := base64.StdEncoding.EncodeToString([]byte(scriptFile))

	qualifiedUser := profile.QualifiedName()
	escapedPassword := strings.ReplaceAll(password, "`", "``")
	escapedPassword = strings.ReplaceAll(escapedPassword, `"`, "`\"")
	escapedPassword = strings.ReplaceAll(escapedPassword, "$", "`$")

	// Outer launcher script: creates PSCredential and starts process as target user
	// Event 4648 is generated here — "A logon was attempted using explicit credentials"
	// Note: backtick is PowerShell's escape char; we build the string via concatenation
	// because Go raw string literals cannot contain backtick characters.
	bq := "`"
	launcher := fmt.Sprintf(
		"$scriptPath = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String(\"%s\"))\n"+
			"$secPass = ConvertTo-SecureString \"%s\" -AsPlainText -Force\n"+
			"$cred = New-Object System.Management.Automation.PSCredential(\"%s\", $secPass)\n"+
			"$psi = New-Object System.Diagnostics.ProcessStartInfo\n"+
			"$psi.FileName = \"powershell.exe\"\n"+
			"$psi.Arguments = \"-NonInteractive -NoProfile -ExecutionPolicy Bypass -File %s\"+$scriptPath+\"%s\"\n"+
			"$psi.UserName = $cred.UserName\n"+
			"$psi.Password = $cred.Password\n"+
			"$psi.UseShellExecute = $false\n"+
			"$psi.RedirectStandardOutput = $true\n"+
			"$psi.RedirectStandardError = $true\n"+
			"$psi.CreateNoWindow = $true\n"+
			"$proc = [System.Diagnostics.Process]::Start($psi)\n"+
			"$outText = $proc.StandardOutput.ReadToEnd()\n"+
			"$errText = $proc.StandardError.ReadToEnd()\n"+
			"$proc.WaitForExit()\n"+
			"[System.IO.File]::WriteAllText(\"%s\", $outText)\n"+
			"[System.IO.File]::WriteAllText(\"%s\", $errText)\n",
		encodedScriptPath,
		escapedPassword,
		qualifiedUser,
		bq+"\"", bq+"\"",
		strings.ReplaceAll(outFile, `\`, `\\`),
		strings.ReplaceAll(errFile, `\`, `\\`),
	)

	cmd := exec.Command("powershell.exe",
		"-NonInteractive", "-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", launcher,
	)
	var launchOut bytes.Buffer
	cmd.Stdout = &launchOut
	cmd.Stderr = &launchOut

	runErr := cmd.Run()

	// Read captured output from temp files
	outBytes, _ := os.ReadFile(outFile)
	errBytes, _ := os.ReadFile(errFile)

	outStr := string(outBytes)
	errStr := string(errBytes)

	if runErr != nil && outStr == "" {
		errStr = launchOut.String() + "\n" + errStr
	}

	return outStr, errStr, runErr
}

// runCommand executes a powershell or cmd command as the current user.
func runCommand(execType, command string) (stdout, stderr string, err error) {
	if strings.TrimSpace(command) == "" {
		return "", "", nil
	}
	var cmd *exec.Cmd
	switch strings.ToLower(execType) {
	case "powershell", "psh":
		cmd = exec.Command("powershell.exe",
			"-NonInteractive", "-NoProfile",
			"-ExecutionPolicy", "Bypass",
			"-Command", command,
		)
	case "cmd", "command_prompt":
		cmd = exec.Command("cmd.exe", "/C", command)
	default:
		return "", "", fmt.Errorf("unsupported executor type: %s", execType)
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
