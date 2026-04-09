package reporter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
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
	WhatIf      bool                        `json:"whatif"`
	Results     []playbooks.ExecutionResult `json:"results"`
}

// SaveResults writes JSON + HTML report files next to the log file.
// Returns the path of the generated HTML file (empty if none generated).
func SaveResults(results []playbooks.ExecutionResult, logFile string, whatIf bool) string {
	if len(results) == 0 {
		return ""
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
		WhatIf:      whatIf,
		Results:     results,
	}

	stamp := time.Now().Format("20060102_150405")

	// JSON report
	jsonFile := fmt.Sprintf("lognojutsu_report_%s.json", stamp)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("[Reporter] Failed to marshal results: %v", err)
	} else if err := os.WriteFile(jsonFile, data, 0644); err != nil {
		log.Printf("[Reporter] Failed to write JSON report: %v", err)
	} else {
		log.Printf("[Reporter] JSON report saved: %s (%d/%d succeeded)", jsonFile, succeeded, len(results))
	}

	// HTML report
	htmlFile := fmt.Sprintf("lognojutsu_report_%s.html", stamp)
	if err := saveHTML(report, htmlFile); err != nil {
		log.Printf("[Reporter] Failed to write HTML report: %v", err)
		return ""
	}
	log.Printf("[Reporter] HTML report saved: %s", htmlFile)
	return htmlFile
}

// ── HTML generation ───────────────────────────────────────────────────────────

type tacticStat struct {
	Tactic     string
	Total      int
	Succeeded  int
	Failed     int
	PctSuccess int
}

type htmlData struct {
	GeneratedAt    string
	LogFile        string
	TotalRun       int
	Succeeded      int
	Failed         int
	SuccessRate    int
	WhatIf         bool
	TacticStats    []tacticStat
	Results        []playbooks.ExecutionResult
	VerifPassed    int
	VerifFailed       int
	VerifAMSIBlocked  int
	VerifElevRequired int
	HasCrowdStrike    bool
	HasSentinel    bool
	HasTier        bool
}

func fmtTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("02.01. 15:04:05")
}

func saveHTML(r Report, filename string) error {
	// Build tactic stats
	tacticMap := make(map[string]*tacticStat)
	for _, res := range r.Results {
		tactic := res.TacticID
		if tactic == "" {
			tactic = "unknown"
		}
		if _, ok := tacticMap[tactic]; !ok {
			tacticMap[tactic] = &tacticStat{Tactic: tactic}
		}
		tacticMap[tactic].Total++
		if res.Success {
			tacticMap[tactic].Succeeded++
		} else {
			tacticMap[tactic].Failed++
		}
	}
	var tactics []tacticStat
	for _, ts := range tacticMap {
		if ts.Total > 0 {
			ts.PctSuccess = ts.Succeeded * 100 / ts.Total
		}
		tactics = append(tactics, *ts)
	}
	// sort by tactic name
	for i := 0; i < len(tactics); i++ {
		for j := i + 1; j < len(tactics); j++ {
			if tactics[i].Tactic > tactics[j].Tactic {
				tactics[i], tactics[j] = tactics[j], tactics[i]
			}
		}
	}

	successRate := 0
	if r.TotalRun > 0 {
		successRate = r.Succeeded * 100 / r.TotalRun
	}

	verifPassed := 0
	verifFailed := 0
	verifAMSIBlocked := 0
	verifElevRequired := 0
	for _, res := range r.Results {
		if res.VerificationStatus == playbooks.VerifPass {
			verifPassed++
		} else if res.VerificationStatus == playbooks.VerifFail {
			verifFailed++
		} else if res.VerificationStatus == playbooks.VerifAMSIBlocked {
			verifAMSIBlocked++
		} else if res.VerificationStatus == playbooks.VerifElevationRequired {
			verifElevRequired++
		}
	}

	hasCrowdStrike := false
	for _, res := range r.Results {
		if len(res.SIEMCoverage["crowdstrike"]) > 0 {
			hasCrowdStrike = true
			break
		}
	}

	hasSentinel := false
	for _, res := range r.Results {
		if len(res.SIEMCoverage["sentinel"]) > 0 {
			hasSentinel = true
			break
		}
	}

	hasTier := false
	for _, res := range r.Results {
		if res.Tier > 0 {
			hasTier = true
			break
		}
	}

	data := htmlData{
		GeneratedAt:    fmtTime(r.GeneratedAt),
		LogFile:        r.LogFile,
		TotalRun:       r.TotalRun,
		Succeeded:      r.Succeeded,
		Failed:         r.Failed,
		SuccessRate:    successRate,
		WhatIf:         r.WhatIf,
		TacticStats:    tactics,
		Results:        r.Results,
		VerifPassed:      verifPassed,
		VerifFailed:      verifFailed,
		VerifAMSIBlocked: verifAMSIBlocked,
		VerifElevRequired: verifElevRequired,
		HasCrowdStrike:   hasCrowdStrike,
		HasSentinel:    hasSentinel,
		HasTier:        hasTier,
	}

	funcMap := template.FuncMap{
		"fmtTime": fmtTime,
		"verifStr": func(v playbooks.VerificationStatus) string { return string(v) },
		"siemCoverage": func(m map[string][]string, key string) []string {
			if m == nil {
				return nil
			}
			return m[key]
		},
		"truncate": func(s string, n int) string {
			s = strings.ReplaceAll(s, "\r\n", "\n")
			if len(s) > n {
				return s[:n] + "…"
			}
			return s
		},
		"tacticColor": func(tactic string) string {
			colors := map[string]string{
				"discovery":          "#58a6ff",
				"lateral-movement":   "#d29922",
				"exfiltration":       "#bc8cff",
				"privilege-escalation": "#f85149",
				"defense-evasion":    "#f85149",
				"credential-access":  "#ff7b72",
				"persistence":        "#39d353",
				"execution":          "#58a6ff",
				"impact":             "#f85149",
				"collection":         "#bc8cff",
			"command-and-control": "#f85149",
			"ueba-scenario":       "#bc8cff",
			}
			if c, ok := colors[strings.ToLower(tactic)]; ok {
				return c
			}
			return "#8b949e"
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<title>LogNoJutsu — Simulation Report</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0d1117;color:#e6edf3;font-family:'Segoe UI',system-ui,sans-serif;font-size:14px}
.hdr{background:#161b22;border-bottom:1px solid #30363d;padding:18px 32px;display:flex;align-items:center;gap:16px}
.hdr h1{font-size:20px;color:#58a6ff;letter-spacing:.5px}
.hdr .meta{color:#8b949e;font-size:12px;margin-top:4px}
.whatif-badge{background:rgba(210,153,34,.15);border:1px solid #d29922;color:#d29922;border-radius:12px;padding:2px 12px;font-size:12px;margin-left:12px}
.wrap{max-width:1200px;margin:0 auto;padding:24px 32px}
.stat-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:28px}
.stat-box{background:#161b22;border:1px solid #30363d;border-radius:8px;padding:16px;text-align:center}
.stat-box .val{font-size:32px;font-weight:700}
.stat-box .lbl{font-size:11px;color:#8b949e;margin-top:4px}
.c-total{color:#58a6ff}.c-ok{color:#3fb950}.c-fail{color:#f85149}.c-pct{color:#d29922}
h2{font-size:15px;margin:28px 0 14px;border-bottom:1px solid #30363d;padding-bottom:8px;color:#e6edf3}
.tactic-grid{display:flex;flex-wrap:wrap;gap:10px;margin-bottom:4px}
.tactic-cell{background:#161b22;border:1px solid #30363d;border-radius:6px;padding:10px 16px;min-width:150px}
.tactic-cell .t-name{font-size:11px;color:#8b949e;margin-bottom:4px;text-transform:uppercase;letter-spacing:.5px}
.tactic-cell .t-nums{font-size:16px;font-weight:700}
.t-bar{height:4px;background:#21262d;border-radius:2px;margin-top:8px}
.t-bar-fill{height:4px;border-radius:2px;background:#3fb950}
.t-bar-fail{background:#f85149}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;color:#8b949e;padding:8px 12px;border-bottom:1px solid #30363d;font-weight:500;font-size:12px}
td{padding:9px 12px;border-bottom:1px solid #30363d;vertical-align:top}
tr:last-child td{border-bottom:none}
tr:hover td{background:#161b22}
.ok{color:#3fb950;font-weight:600}.fail{color:#f85149;font-weight:600}
.tag{display:inline-block;background:#21262d;border:1px solid #30363d;border-radius:4px;padding:1px 7px;font-size:11px;color:#8b949e;margin-right:3px}
.mono{font-family:Consolas,'Courier New',monospace}
.output{font-family:Consolas,monospace;font-size:11px;white-space:pre-wrap;background:#0a0e13;padding:8px;border-radius:4px;color:#8b949e;max-height:120px;overflow-y:auto;margin-top:6px;border:1px solid #21262d}
.verif-pass{color:#3fb950;font-weight:600}
.verif-fail{color:#f85149;font-weight:600}
.verif-skip{color:#8b949e}
.verif-amsi{color:#d29922;font-weight:600}
.verif-elev{color:#8b949e;font-weight:600}
.verif-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.verif-list li{margin:1px 0}
.footer{text-align:center;color:#8b949e;font-size:12px;padding:24px;border-top:1px solid #30363d;margin-top:32px}
{{if .HasCrowdStrike}}.cs-badge{background:#e01b22;color:#fff;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.cs-na{color:#8b949e;font-size:11px}
.cs-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.cs-list li{margin:1px 0;color:#e6edf3}{{end}}
{{if .HasSentinel}}.ms-badge{background:#0078D4;color:#fff;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}
.ms-na{color:#8b949e;font-size:11px}
.ms-list{font-size:11px;margin-top:4px;padding-left:0;list-style:none}
.ms-list li{margin:1px 0;color:#e6edf3}{{end}}
{{if .HasTier}}.tier1-badge{background:rgba(63,185,80,0.15);color:#3fb950;border:1px solid #3fb950;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}.tier2-badge{background:rgba(210,153,34,0.15);color:#d29922;border:1px solid #d29922;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}.tier3-badge{background:rgba(139,148,158,0.15);color:#8b949e;border:1px solid #8b949e;border-radius:4px;padding:1px 7px;font-size:11px;font-weight:600;display:inline-block}{{end}}
@media print{.hdr{background:#fff;color:#000}.body{background:#fff}}
</style>
</head>
<body>
<div class="hdr">
  <div>
    <h1>⚔️ LogNoJutsu — Simulation Report{{if .WhatIf}} <span class="whatif-badge">⚠ WhatIf-Modus — Keine echte Ausführung</span>{{end}}</h1>
    <div class="meta">Generiert: {{.GeneratedAt}}{{if .LogFile}} · Log: {{.LogFile}}{{end}}</div>
  </div>
</div>
<div class="wrap">

  <div class="stat-grid">
    <div class="stat-box"><div class="val c-total">{{.TotalRun}}</div><div class="lbl">Gesamt</div></div>
    <div class="stat-box"><div class="val c-ok">{{.Succeeded}}</div><div class="lbl">Erfolgreich</div></div>
    <div class="stat-box"><div class="val c-fail">{{.Failed}}</div><div class="lbl">Fehlgeschlagen</div></div>
    <div class="stat-box"><div class="val c-pct">{{.SuccessRate}}%</div><div class="lbl">Erfolgsquote</div></div>
    {{if gt .VerifPassed 0}}<div class="stat-box"><div class="val" style="color:#3fb950">{{.VerifPassed}}</div><div class="lbl">Verified Pass</div></div>{{end}}
    {{if gt .VerifFailed 0}}<div class="stat-box"><div class="val" style="color:#f85149">{{.VerifFailed}}</div><div class="lbl">Verified Fail</div></div>{{end}}
    {{if gt .VerifAMSIBlocked 0}}<div class="stat-box"><div class="val" style="color:#d29922">{{.VerifAMSIBlocked}}</div><div class="lbl">AMSI Blocked</div></div>{{end}}
    {{if gt .VerifElevRequired 0}}<div class="stat-box"><div class="val" style="color:#8b949e">{{.VerifElevRequired}}</div><div class="lbl">Elevation Skipped</div></div>{{end}}
  </div>

  <h2>📊 MITRE ATT&amp;CK Taktiken</h2>
  <div class="tactic-grid">
    {{range .TacticStats}}
    <div class="tactic-cell">
      <div class="t-name">{{.Tactic}}</div>
      <div class="t-nums" style="color:{{tacticColor .Tactic}}">{{.Total}} <span style="font-size:12px;font-weight:400;color:#8b949e;">Techniken</span></div>
      <div style="font-size:11px;color:#8b949e;margin-top:2px;">✓ {{.Succeeded}} / ✗ {{.Failed}}</div>
      <div class="t-bar"><div class="t-bar-fill{{if gt .Failed 0}} t-bar-fail{{end}}" style="width:{{.PctSuccess}}%"></div></div>
    </div>
    {{end}}
  </div>

  <h2>📋 Ausgeführte Techniken</h2>
  <table>
    <thead>
      <tr>
        <th>Zeit</th>
        <th>Technik-ID</th>
        <th>Name</th>
        <th>Taktik</th>
        <th>Status</th>
        <th>Verifikation</th>
        {{if .HasCrowdStrike}}<th>CrowdStrike</th>{{end}}
        {{if .HasSentinel}}<th>Microsoft Sentinel</th>{{end}}
        {{if .HasTier}}<th>Tier</th>{{end}}
        <th>Benutzer</th>
      </tr>
    </thead>
    <tbody>
    {{range .Results}}
    <tr>
      <td class="mono" style="color:#8b949e;white-space:nowrap;font-size:12px;">{{fmtTime .StartTime}}</td>
      <td class="mono" style="color:#58a6ff;">{{.TechniqueID}}</td>
      <td>
        <div>{{.TechniqueName}}</div>
        {{if .Output}}<div class="output">{{truncate .Output 600}}</div>{{end}}
        {{if .ErrorOutput}}<div class="output" style="border-color:#f85149;">{{truncate .ErrorOutput 300}}</div>{{end}}
      </td>
      <td><span class="tag">{{.TacticID}}</span></td>
      <td class="{{if .Success}}ok{{else}}fail{{end}}">{{if .Success}}✓ OK{{else}}✗ Fehler{{end}}</td>
      <td>
        {{if eq (verifStr .VerificationStatus) "pass"}}
          <span class="verif-pass">&#10003; Pass</span>
        {{else if eq (verifStr .VerificationStatus) "fail"}}
          <span class="verif-fail">&#10007; Fail</span>
        {{else if eq (verifStr .VerificationStatus) "not_executed"}}
          <span class="verif-skip">&mdash; Nicht ausgeführt</span>
        {{else if eq (verifStr .VerificationStatus) "amsi_blocked"}}
          <span class="verif-amsi">&#9888; AMSI Blocked</span>
        {{else if eq (verifStr .VerificationStatus) "elevation_required"}}
          <span class="verif-elev">&#8593; Elevation Required</span>
        {{else}}
          <span class="verif-skip">&mdash;</span>
        {{end}}
        {{if .VerifiedEvents}}
        <ul class="verif-list">
          {{range .VerifiedEvents}}
          <li>{{if .Found}}&#10003;{{else}}&#10007;{{end}} EID {{.EventID}} <span style="color:#8b949e">{{.Channel}}</span></li>
          {{end}}
        </ul>
        {{end}}
      </td>
      {{if $.HasCrowdStrike}}
      <td>
        {{$cs := siemCoverage .SIEMCoverage "crowdstrike"}}
        {{if $cs}}
          <span class="cs-badge">CS</span>
          <ul class="cs-list">
            {{range $cs}}<li>{{.}}</li>{{end}}
          </ul>
        {{else}}
          <span class="cs-na">N/A</span>
        {{end}}
      </td>
      {{end}}
      {{if $.HasSentinel}}
      <td>
        {{$ms := siemCoverage .SIEMCoverage "sentinel"}}
        {{if $ms}}
          <span class="ms-badge">MS</span>
          <ul class="ms-list">
            {{range $ms}}<li>{{.}}</li>{{end}}
          </ul>
        {{else}}
          <span class="ms-na">N/A</span>
        {{end}}
      </td>
      {{end}}
      {{if $.HasTier}}<td>{{if eq .Tier 1}}<span class="tier1-badge">T1</span>{{else if eq .Tier 2}}<span class="tier2-badge">T2</span>{{else if eq .Tier 3}}<span class="tier3-badge">T3</span>{{else}}&mdash;{{end}}</td>{{end}}
      <td style="color:#bc8cff;font-size:12px;">{{.RunAsUser}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>

</div>
<div class="footer">LogNoJutsu SIEM Validation Tool · Nur für autorisierte Tests in kontrollierten Umgebungen</div>
</body>
</html>`
