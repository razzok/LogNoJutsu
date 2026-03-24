package verifier

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"lognojutsu/internal/playbooks"
)

// QueryFn abstracts event log querying for testability.
// Takes channel, eventID, since time; returns count of matching events.
type QueryFn func(channel string, eventID int, since time.Time) (int, error)

// DefaultQueryFn queries the Windows Event Log via PowerShell Get-WinEvent.
func DefaultQueryFn(channel string, eventID int, since time.Time) (int, error) {
	sinceStr := since.Format(time.RFC3339)
	script := fmt.Sprintf(
		`(Get-WinEvent -FilterHashtable @{LogName='%s'; Id=%d; StartTime='%s'} -ErrorAction SilentlyContinue | Measure-Object).Count`,
		channel, eventID, sinceStr,
	)
	cmd := exec.Command("powershell.exe",
		"-NonInteractive", "-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", script,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return 0, nil // channel may not exist — treat as zero
	}
	count, _ := strconv.Atoi(strings.TrimSpace(out.String()))
	return count, nil
}

// Verify checks whether expected events appeared in the Windows Event Log.
// If executionSuccess is false, returns VerifNotExecuted immediately.
// If specs is empty, returns VerifNotRun.
// Otherwise queries each EventSpec using queryFn and returns pass/fail.
func Verify(specs []playbooks.EventSpec, since time.Time, executionSuccess bool, queryFn QueryFn) (playbooks.VerificationStatus, []playbooks.VerifiedEvent) {
	if !executionSuccess {
		return playbooks.VerifNotExecuted, nil
	}
	if len(specs) == 0 {
		return playbooks.VerifNotRun, nil
	}

	verified := make([]playbooks.VerifiedEvent, 0, len(specs))
	allFound := true
	for _, spec := range specs {
		count, _ := queryFn(spec.Channel, spec.EventID, since)
		found := count > 0
		if !found {
			allFound = false
		}
		verified = append(verified, playbooks.VerifiedEvent{
			EventID:     spec.EventID,
			Channel:     spec.Channel,
			Description: spec.Description,
			Found:       found,
			Count:       count,
		})
	}

	status := playbooks.VerifPass
	if !allFound {
		status = playbooks.VerifFail
	}
	return status, verified
}
