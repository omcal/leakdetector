package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"leakcheck/internal/parser"
	"path/filepath"
	"sort"
)

// Reporter formats and outputs leak detection results
type Reporter struct {
	output io.Writer
	json   bool
}

// NewReporter creates a new reporter
func NewReporter(output io.Writer, jsonOutput bool) *Reporter {
	return &Reporter{
		output: output,
		json:   jsonOutput,
	}
}

// Report outputs the leak findings
func (r *Reporter) Report(leaks []parser.Leak) error {
	if r.json {
		return r.reportJSON(leaks)
	}
	return r.reportConsole(leaks)
}

func (r *Reporter) reportConsole(leaks []parser.Leak) error {
	if len(leaks) == 0 {
		fmt.Fprintln(r.output, "[OK] No potential memory leaks detected.")
		return nil
	}

	// Sort by file, then line
	sort.Slice(leaks, func(i, j int) bool {
		if leaks[i].File != leaks[j].File {
			return leaks[i].File < leaks[j].File
		}
		return leaks[i].Line < leaks[j].Line
	})

	// Group by file
	currentFile := ""
	for _, leak := range leaks {
		if leak.File != currentFile {
			currentFile = leak.File
			relPath := filepath.Base(currentFile)
			fmt.Fprintf(r.output, "\n%s:\n", relPath)
		}

		icon := "[ERROR]"
		if leak.Severity == "warning" {
			icon = "[WARN] "
		}

		fmt.Fprintf(r.output, "  %s Line %d [%s::%s]: %s\n",
			icon, leak.Line, leak.ClassName, leak.VarName, leak.Reason)

		if leak.Recommendation != "" {
			fmt.Fprintf(r.output, "         -> Fix: %s\n", leak.Recommendation)
		}
	}

	// Summary
	errors := 0
	warnings := 0
	for _, leak := range leaks {
		if leak.Severity == "error" {
			errors++
		} else {
			warnings++
		}
	}

	fmt.Fprintf(r.output, "\nSummary: %d error(s), %d warning(s)\n", errors, warnings)
	return nil
}

func (r *Reporter) reportJSON(leaks []parser.Leak) error {
	output := struct {
		Leaks   []parser.Leak `json:"leaks"`
		Summary Summary       `json:"summary"`
	}{
		Leaks: leaks,
		Summary: Summary{
			TotalIssues: len(leaks),
			Errors:      countBySeverity(leaks, "error"),
			Warnings:    countBySeverity(leaks, "warning"),
		},
	}

	if output.Leaks == nil {
		output.Leaks = []parser.Leak{}
	}

	encoder := json.NewEncoder(r.output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// Summary holds aggregate information about the analysis
type Summary struct {
	TotalIssues int `json:"total_issues"`
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
}

func countBySeverity(leaks []parser.Leak, severity string) int {
	count := 0
	for _, leak := range leaks {
		if leak.Severity == severity {
			count++
		}
	}
	return count
}
