//go:build windows

package native

import (
	"fmt"
	"strings"

	"github.com/yusufpapurcu/wmi"
)

// win32Process maps WMI Win32_Process fields used for T1057 process discovery.
// CommandLine can be empty for system/kernel processes — this is expected (Pitfall 3).
type win32Process struct {
	Name            string
	ProcessId       uint32
	ParentProcessId uint32
	CommandLine     string
}

// runT1057 performs T1057 Process Discovery by querying Win32_Process via WMI.
// It returns the process list with PID, PPID, name, and command line.
// Uses pure Go via go-ole/wmi — no CGO required (ARCH-03).
func runT1057() (NativeResult, error) {
	var procs []win32Process
	// No WHERE filter — include all processes; empty CommandLine is valid for system procs (Pitfall 3)
	// Pass "Win32_Process" explicitly because the struct name (win32Process) differs from the WMI class name.
	q := wmi.CreateQuery(&procs, "", "Win32_Process")
	if err := wmi.Query(q, &procs); err != nil {
		msg := fmt.Sprintf("Process discovery: WMI Win32_Process query failed: %s", err.Error())
		return NativeResult{Success: false, ErrorOutput: msg}, fmt.Errorf("T1057 wmi.Query: %w", err)
	}

	var sb strings.Builder
	total := len(procs)
	fmt.Fprintf(&sb, "Process Discovery (Win32_Process) — %d processes found\n", total)

	// Limit output to first 50 processes to avoid enormous output
	limit := total
	if limit > 50 {
		limit = 50
	}
	for i := 0; i < limit; i++ {
		p := procs[i]
		fmt.Fprintf(&sb, "PID=%-6d PPID=%-6d Name=%s CommandLine=%s\n",
			p.ProcessId, p.ParentProcessId, p.Name, p.CommandLine)
	}
	if total > 50 {
		fmt.Fprintf(&sb, "... (%d more processes omitted)\n", total-50)
	}

	return NativeResult{Output: sb.String(), Success: true}, nil
}

func init() {
	Register("T1057", runT1057, nil)
}
