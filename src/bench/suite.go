// discovers and drives multiple bench manifests under a directory tree as one suite.
package bench

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// DiscoverManifests walks root for files named "bench.yaml" and returns their
// absolute paths in a deterministic, sorted order.
func DiscoverManifests(root string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "bench.yaml" {
			abs, absErr := filepath.Abs(path)
			if absErr != nil {
				return absErr
			}
			found = append(found, abs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(found)
	return found, nil
}

// SuiteManifestValidation is one manifest's validation outcome within a suite.
type SuiteManifestValidation struct {
	ManifestPath string `json:"manifestPath"`
	Name         string `json:"name,omitempty"`
	Cases        int    `json:"cases,omitempty"`
	Variants     int    `json:"variants,omitempty"`
	Error        string `json:"error,omitempty"`
}

// SuiteValidation aggregates validation outcomes across every discovered manifest.
type SuiteValidation struct {
	Root      string                    `json:"root"`
	Manifests []SuiteManifestValidation `json:"manifests"`
	Valid     bool                      `json:"valid"`
}

// ValidateSuite loads and strictly validates every bench.yaml found under root
// without executing anything. It never returns an error for an individual
// manifest's invalidity — that is recorded per-manifest in the result so a
// caller can report every failure in one pass instead of stopping at the first.
func ValidateSuite(root string) (*SuiteValidation, error) {
	paths, err := DiscoverManifests(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no bench.yaml manifests found under %s", root)
	}
	result := &SuiteValidation{Root: root, Valid: true}
	for _, path := range paths {
		mv := SuiteManifestValidation{ManifestPath: path}
		m, err := LoadManifest(path)
		if err != nil {
			mv.Error = err.Error()
			result.Valid = false
		} else {
			mv.Name = m.Name
			mv.Cases = len(m.Cases)
			mv.Variants = len(m.Variants)
		}
		result.Manifests = append(result.Manifests, mv)
	}
	return result, nil
}

// SuiteManifestPlan is one manifest's planning outcome within a suite.
type SuiteManifestPlan struct {
	ManifestPath string `json:"manifestPath"`
	PlanPath     string `json:"planPath,omitempty"`
	PlanID       string `json:"planId,omitempty"`
	Runs         int    `json:"runs,omitempty"`
	Error        string `json:"error,omitempty"`
}

// SuitePlan aggregates planning outcomes across every discovered manifest.
type SuitePlan struct {
	Root      string              `json:"root"`
	Manifests []SuiteManifestPlan `json:"manifests"`
	Planned   bool                `json:"planned"`
}

// PlanSuite builds and saves a run plan for every bench.yaml found under root,
// writing each plan next to its manifest's own output directory (the same
// default location a single `bench plan` invocation would use).
func PlanSuite(root string) (*SuitePlan, error) {
	paths, err := DiscoverManifests(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no bench.yaml manifests found under %s", root)
	}
	result := &SuitePlan{Root: root, Planned: true}
	for _, path := range paths {
		mp := SuiteManifestPlan{ManifestPath: path}
		m, err := LoadManifest(path)
		if err != nil {
			mp.Error = err.Error()
			result.Planned = false
			result.Manifests = append(result.Manifests, mp)
			continue
		}
		plan, err := BuildPlan(path, m)
		if err != nil {
			mp.Error = err.Error()
			result.Planned = false
			result.Manifests = append(result.Manifests, mp)
			continue
		}
		planPath := filepath.Join(m.Output.Directory, safeManifestName(m.Name)+".plan.json")
		if err := WritePlan(planPath, plan); err != nil {
			mp.Error = err.Error()
			result.Planned = false
			result.Manifests = append(result.Manifests, mp)
			continue
		}
		mp.PlanPath = planPath
		mp.PlanID = plan.PlanID
		mp.Runs = len(plan.Runs)
		result.Manifests = append(result.Manifests, mp)
	}
	return result, nil
}

// SuiteManifestResult is one manifest's execution outcome within a suite run.
type SuiteManifestResult struct {
	ManifestPath string      `json:"manifestPath"`
	Name         string      `json:"name,omitempty"`
	Experiment   *Experiment `json:"experiment,omitempty"`
	Error        string      `json:"error,omitempty"`
}

// SuiteResult aggregates execution outcomes across every discovered manifest.
// There is no cross-manifest comparison primitive in the engine — each
// manifest's own Comparison.Outcome (computed independently by compare.go)
// is classified here as passed/failed/errored to produce a suite-level tally.
type SuiteResult struct {
	Root      string                `json:"root"`
	Manifests []SuiteManifestResult `json:"manifests"`
	Passed    int                   `json:"passed"`
	Failed    int                   `json:"failed"`
	Errored   int                   `json:"errored"`
}

// Eligible reports whether every discovered manifest ran and passed. It
// mirrors the exit-code classification `bench run` already applies to a
// single manifest: any outcome other than "evaluation-failed" or
// "inconclusive" counts as passed.
func (s SuiteResult) Eligible() bool {
	return len(s.Manifests) > 0 && s.Errored == 0 && s.Failed == 0
}

// SuiteOptions configures suite execution.
type SuiteOptions struct {
	// Observer receives every manifest's lifecycle events, prefixed with the
	// manifest path so a caller can distinguish concurrent-looking output
	// from what is actually sequential, per-manifest execution.
	Observer func(manifestPath string, event RunEvent)
}

// RunSuite discovers every bench.yaml under root and runs each to completion
// in sorted order, forwarding lifecycle events per-manifest to the optional
// observer. A manifest that fails to load, plan, or run is recorded as
// errored and does not stop the remaining manifests from running.
func RunSuite(ctx context.Context, root string, options SuiteOptions) (*SuiteResult, error) {
	paths, err := DiscoverManifests(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no bench.yaml manifests found under %s", root)
	}
	result := &SuiteResult{Root: root}
	for _, path := range paths {
		mr := SuiteManifestResult{ManifestPath: path}
		m, err := LoadManifest(path)
		if err != nil {
			mr.Error = err.Error()
			result.Errored++
			result.Manifests = append(result.Manifests, mr)
			continue
		}
		mr.Name = m.Name
		plan, err := BuildPlan(path, m)
		if err != nil {
			mr.Error = err.Error()
			result.Errored++
			result.Manifests = append(result.Manifests, mr)
			continue
		}
		var observer func(RunEvent)
		if options.Observer != nil {
			observer = func(event RunEvent) { options.Observer(path, event) }
		}
		experiment, err := RunPlanWithOptions(ctx, plan, RunOptions{Observer: observer})
		if err != nil {
			mr.Error = err.Error()
			result.Errored++
			result.Manifests = append(result.Manifests, mr)
			continue
		}
		mr.Experiment = experiment
		if experiment.Comparison.Outcome == "evaluation-failed" || experiment.Comparison.Outcome == "inconclusive" {
			result.Failed++
		} else {
			result.Passed++
		}
		result.Manifests = append(result.Manifests, mr)
	}
	return result, nil
}

// safeManifestName mirrors benchcli's own filename-safing so suite-planned
// paths match what a single `bench plan` invocation would produce.
func safeManifestName(name string) string {
	var b []byte
	for _, r := range name {
		lower := r
		if lower >= 'A' && lower <= 'Z' {
			lower += 'a' - 'A'
		}
		switch {
		case lower >= 'a' && lower <= 'z', lower >= '0' && lower <= '9', lower == '-', lower == '_':
			b = append(b, byte(lower))
		case len(b) > 0 && b[len(b)-1] != '-':
			b = append(b, '-')
		}
	}
	for len(b) > 0 && b[len(b)-1] == '-' {
		b = b[:len(b)-1]
	}
	if len(b) == 0 {
		return "bench"
	}
	return string(b)
}
