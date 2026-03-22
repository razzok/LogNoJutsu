package reporter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"lognojutsu/internal/playbooks"
)

// Report is the full simulation report written to disk.
type Report struct {
	GeneratedAt string                      `json:"generated_at"`
	LogFile     string                      `json:"log_file"`
	TotalRun    int                         `json:"total_run"`
	Succeeded   int                         `json:"succeeded"`
	Failed      int                         `json:"failed"`
	Results     []playbooks.ExecutionResult `json:"results"`
}

// SaveResults writes a JSON report file next to the log file.
func SaveResults(results []playbooks.ExecutionResult, logFile string) {
	if len(results) == 0 {
		return
	}

	succeeded := 0
	for _, r := range results {
		if r.Success {
			succeeded++
		}
	}

	report := Report{
		GeneratedAt: time.Now().Format(time.RFC3339),
		LogFile:     logFile,
		TotalRun:    len(results),
		Succeeded:   succeeded,
		Failed:      len(results) - succeeded,
		Results:     results,
	}

	filename := fmt.Sprintf("lognojutsu_report_%s.json", time.Now().Format("20060102_150405"))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("[Reporter] Failed to marshal results: %v", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("[Reporter] Failed to write report: %v", err)
		return
	}

	log.Printf("[Reporter] JSON report saved: %s (%d/%d techniques succeeded)", filename, succeeded, len(results))
}
