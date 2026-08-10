// renders persistent Markdown and JSON benchmark reports from experiment evidence.
package bench

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteReport(path string, experiment *Experiment) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Benchmark report: %s\n\n", experiment.Name)
	fmt.Fprintf(&b, "- Experiment: `%s`\n- Plan: `%s`\n- Created: %s\n- Outcome: **%s**\n", experiment.ID, experiment.PlanID, experiment.CreatedAt.Format("2006-01-02T15:04:05Z"), experiment.Comparison.Outcome)
	if experiment.Comparison.Winner != "" {
		fmt.Fprintf(&b, "- Winner: **%s**\n", experiment.Comparison.Winner)
	}
	fmt.Fprintf(&b, "- Reason: %s\n\n## Variants\n\n| Variant | Runs | Required pass rate | Optional mean | Median duration (ms) | Eligible |\n|---|---:|---:|---:|---:|---|\n", experiment.Comparison.Reason)
	for _, s := range experiment.Statistics {
		optional := "unknown"
		if s.OptionalScore.Mean != nil {
			optional = fmt.Sprintf("%.3f", *s.OptionalScore.Mean)
		}
		duration := "unknown"
		if s.DurationMS.Median != nil {
			duration = fmt.Sprintf("%.0f", *s.DurationMS.Median)
		}
		fmt.Fprintf(&b, "| %s | %d | %.1f%% | %s | %s | %t |\n", s.VariantID, s.Count, s.RequiredPassRate*100, optional, duration, s.Eligible)
	}
	if len(experiment.Warnings) > 0 {
		b.WriteString("\n## Comparability warnings\n\n")
		for _, warning := range experiment.Warnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	missingTokens, missingCosts := 0, 0
	for _, s := range experiment.Statistics {
		missingTokens += s.TotalTokens.Missing
		missingCosts += s.EstimatedCost.Missing
	}
	if missingTokens > 0 || missingCosts > 0 {
		fmt.Fprintf(&b, "\n## Missing metrics\n\n- Total-token measurements unavailable for %d run(s).\n- Cost measurements unavailable for %d run(s).\n", missingTokens, missingCosts)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".report-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
