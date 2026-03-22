// Package simlog provides structured simulation logging to file and memory.
// Every action during a simulation run is recorded with full detail.
package simlog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// EntryType categorizes log entries for filtering.
type EntryType string

const (
	TypeSimStart   EntryType = "SIM_START"
	TypeSimEnd     EntryType = "SIM_END"
	TypePhase      EntryType = "PHASE"
	TypeTechStart  EntryType = "TECH_START"
	TypeTechEnd    EntryType = "TECH_END"
	TypeCommand    EntryType = "COMMAND"
	TypeOutput     EntryType = "OUTPUT"
	TypeCleanup    EntryType = "CLEANUP"
	TypeError      EntryType = "ERROR"
	TypeInfo       EntryType = "INFO"
	TypePrepStep   EntryType = "PREP"
)

// Entry is a single log record.
type Entry struct {
	Timestamp   string    `json:"timestamp"`
	Type        EntryType `json:"type"`
	TechniqueID string    `json:"technique_id,omitempty"`
	Message     string    `json:"message"`
	Detail      string    `json:"detail,omitempty"`
}

// Logger writes simulation logs to file and keeps them in memory for the UI.
type Logger struct {
	mu       sync.RWMutex
	entries  []Entry
	file     *os.File
	filePath string
}

var current *Logger
var globalMu sync.Mutex

// Start creates a new log session and opens the log file.
func Start(campaignID string) {
	globalMu.Lock()
	defer globalMu.Unlock()

	ts := time.Now().Format("20060102_150405")
	name := "lognojutsu_" + ts
	if campaignID != "" {
		safe := strings.NewReplacer(" ", "_", "/", "-", "\\", "-").Replace(campaignID)
		name += "_" + safe
	}
	name += ".log"

	f, err := os.Create(name)
	if err != nil {
		fmt.Printf("[simlog] WARNING: could not create log file: %v\n", err)
		f = nil
	}

	current = &Logger{
		entries:  []Entry{},
		file:     f,
		filePath: name,
	}

	current.write(TypeSimStart, "", fmt.Sprintf("=== LogNoJutsu Simulation Started ==="), "")
	current.write(TypeSimStart, "", fmt.Sprintf("Campaign: %s | Time: %s", campaignID, time.Now().Format(time.RFC1123)), "")
	if f != nil {
		current.write(TypeInfo, "", fmt.Sprintf("Log file: %s", name), "")
	}
}

// Stop finalises the log session.
func Stop(summary string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	current.write(TypeSimEnd, "", "=== Simulation Ended ===", "")
	current.write(TypeSimEnd, "", summary, "")
	if current.file != nil {
		_ = current.file.Close()
	}
}

// Phase logs a phase transition.
func Phase(phase string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	separator := strings.Repeat("─", 60)
	current.write(TypePhase, "", separator, "")
	current.write(TypePhase, "", fmt.Sprintf("▶ PHASE: %s", strings.ToUpper(phase)), "")
	current.write(TypePhase, "", separator, "")
}

// TechStart logs the beginning of a technique execution.
func TechStart(id, name, tactic string, elevationRequired bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	current.write(TypeTechStart, id, fmt.Sprintf("[START] %s — %s", id, name), "")
	current.write(TypeInfo, id, fmt.Sprintf("  Tactic: %s | Elevation required: %v", tactic, elevationRequired), "")
}

// TechCommand logs the command that is about to be executed.
func TechCommand(id, execType, command string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	current.write(TypeCommand, id, fmt.Sprintf("  Executor: %s", execType), "")
	// Log command, but indent each line for readability
	for _, line := range strings.Split(strings.TrimSpace(command), "\n") {
		current.write(TypeCommand, id, "  CMD> "+strings.TrimSpace(line), "")
	}
}

// TechOutput logs stdout/stderr of a technique.
func TechOutput(id, stdout, stderr string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	if strings.TrimSpace(stdout) != "" {
		current.write(TypeOutput, id, "  --- Output ---", "")
		for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
			current.write(TypeOutput, id, "  "+line, "")
		}
	}
	if strings.TrimSpace(stderr) != "" {
		current.write(TypeError, id, "  --- Stderr ---", "")
		for _, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
			current.write(TypeError, id, "  "+line, "")
		}
	}
}

// TechEnd logs the completion of a technique.
func TechEnd(id string, success bool, duration time.Duration) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	status := "SUCCESS ✓"
	if !success {
		status = "FAILED ✗"
	}
	current.write(TypeTechEnd, id, fmt.Sprintf("[END] %s — %s (duration: %s)", id, status, duration.Round(time.Millisecond)), "")
}

// TechCleanup logs cleanup actions for a technique.
func TechCleanup(id, command string, success bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	current.write(TypeCleanup, id, fmt.Sprintf("  [CLEANUP] %s", id), "")
	for _, line := range strings.Split(strings.TrimSpace(command), "\n") {
		current.write(TypeCleanup, id, "  CLN> "+strings.TrimSpace(line), "")
	}
	status := "OK"
	if !success {
		status = "FAILED"
	}
	current.write(TypeCleanup, id, fmt.Sprintf("  [CLEANUP DONE] %s — %s", id, status), "")
}

// PrepStep logs a preparation step.
func PrepStep(step, message string, success bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		// Allow prep logging even outside a simulation
		fmt.Printf("[PREP] %s: %s (ok=%v)\n", step, message, success)
		return
	}
	status := "OK"
	if !success {
		status = "FAILED"
	}
	current.write(TypePrepStep, "", fmt.Sprintf("[PREP] %s — %s: %s", step, status, message), "")
}

// Info logs a general info message.
func Info(msg string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return
	}
	current.write(TypeInfo, "", msg, "")
}

// GetEntries returns a snapshot of all log entries (for the UI).
func GetEntries() []Entry {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return nil
	}
	current.mu.RLock()
	defer current.mu.RUnlock()
	cp := make([]Entry, len(current.entries))
	copy(cp, current.entries)
	return cp
}

// GetFilePath returns the current log file path.
func GetFilePath() string {
	globalMu.Lock()
	defer globalMu.Unlock()
	if current == nil {
		return ""
	}
	return current.filePath
}

// write is the internal write method — must be called with globalMu held.
func (l *Logger) write(t EntryType, techID, msg, detail string) {
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	entry := Entry{
		Timestamp:   ts,
		Type:        t,
		TechniqueID: techID,
		Message:     msg,
		Detail:      detail,
	}

	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()

	line := fmt.Sprintf("[%s] [%-12s] %s\n", ts, t, msg)
	fmt.Print(line)
	if l.file != nil {
		_, _ = l.file.WriteString(line)
	}
}
