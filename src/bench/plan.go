// builds, persists, and verifies immutable hashed benchmark run plans.
package bench

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const PlanSchemaVersion = 2

type Plan struct {
	SchemaVersion int           `json:"schemaVersion"`
	PlanID        string        `json:"planId"`
	ManifestPath  string        `json:"manifestPath"`
	ManifestHash  string        `json:"manifestHash"`
	Name          string        `json:"name"`
	Seed          int64         `json:"seed"`
	Manifest      Manifest      `json:"manifest"`
	Assets        []AssetDigest `json:"assets"`
	Runs          []PlannedRun  `json:"runs"`
	Warnings      []string      `json:"warnings,omitempty"`
}
type PlannedRun struct {
	RunID      string `json:"runId"`
	CaseID     string `json:"caseId"`
	VariantID  string `json:"variantId"`
	Repetition int    `json:"repetition"`
}

func BuildPlan(manifestPath string, m *Manifest) (*Plan, error) {
	abs, err := filepath.Abs(manifestPath)
	if err != nil {
		return nil, err
	}
	mh, err := digestFile(abs)
	if err != nil {
		return nil, err
	}
	seed := int64(0)
	if m.Execution.Seed != nil {
		seed = *m.Execution.Seed
	} else {
		if err := binary.Read(rand.Reader, binary.LittleEndian, &seed); err != nil {
			return nil, err
		}
	}
	p := &Plan{SchemaVersion: PlanSchemaVersion, ManifestPath: abs, ManifestHash: mh, Name: m.Name, Seed: seed, Manifest: *m}
	add := func(kind, path string) error {
		if path == "" {
			return nil
		}
		var excludes []string
		info, statErr := os.Stat(path)
		if statErr != nil {
			return statErr
		}
		if info.IsDir() {
			rel, relErr := filepath.Rel(path, p.Manifest.Output.Directory)
			if relErr == nil && rel == "." {
				return fmt.Errorf("output directory must not equal referenced directory %s", path)
			} else if relErr == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				excludes = append(excludes, p.Manifest.Output.Directory)
			}
		}
		d, err := digestPathExcluding(path, excludes)
		if err != nil {
			return fmt.Errorf("hash %s %s: %w", kind, path, err)
		}
		p.Assets = append(p.Assets, AssetDigest{Kind: kind, Path: path, SHA256: d, Excludes: excludes})
		return nil
	}
	for _, c := range m.Cases {
		if c.Given.Prompt.File != "" {
			if err := add("prompt", c.Given.Prompt.File); err != nil {
				return nil, err
			}
		}
		if c.Fixture != nil {
			if err := add("fixture", c.Fixture.Source); err != nil {
				return nil, err
			}
		}
		for _, group := range [][]InputAsset{c.Given.Specs, c.Given.Files, c.Given.Directories, c.Given.Images} {
			for _, a := range group {
				if err := add("input", a.Source); err != nil {
					return nil, err
				}
			}
		}
		for _, e := range c.Expect {
			if err := add("oracle", e.ValueFile); err != nil {
				return nil, err
			}
			if err := add("oracle", e.Golden); err != nil {
				return nil, err
			}
		}
	}
	for _, g := range m.Graders {
		for _, a := range g.Assets {
			if err := add("grader", a); err != nil {
				return nil, err
			}
		}
	}
	resolveExecutable := func(label string, executable *string) error {
		resolved, err := exec.LookPath(*executable)
		if err != nil {
			return fmt.Errorf("resolve %s executable: %w", label, err)
		}
		absExe, err := filepath.Abs(resolved)
		if err != nil {
			return fmt.Errorf("resolve %s executable path: %w", label, err)
		}
		realExe, err := filepath.EvalSymlinks(absExe)
		if err != nil {
			return fmt.Errorf("resolve %s executable identity: %w", label, err)
		}
		*executable = realExe
		return add("executable", realExe)
	}
	for i := range p.Manifest.Variants {
		v := &p.Manifest.Variants[i]
		if v.Adapter == "process" && v.Process != nil {
			if err := resolveExecutable("variant "+v.ID, &v.Process.Executable); err != nil {
				return nil, err
			}
		}
		for j := range v.Setup {
			if err := resolveExecutable(fmt.Sprintf("variant %s setup[%d]", v.ID, j), &v.Setup[j].Executable); err != nil {
				return nil, err
			}
		}
	}
	for i := range p.Manifest.Cases {
		for j := range p.Manifest.Cases[i].Expect {
			expectation := &p.Manifest.Cases[i].Expect[j]
			if expectation.Command != nil {
				if err := resolveExecutable("case "+p.Manifest.Cases[i].ID+" expectation "+expectation.ID, &expectation.Command.Executable); err != nil {
					return nil, err
				}
			}
		}
	}
	for i := range p.Manifest.Graders {
		if err := resolveExecutable("grader "+p.Manifest.Graders[i].ID, &p.Manifest.Graders[i].Command.Executable); err != nil {
			return nil, err
		}
	}
	sort.Slice(p.Assets, func(i, j int) bool {
		if p.Assets[i].Path == p.Assets[j].Path {
			return p.Assets[i].Kind < p.Assets[j].Kind
		}
		return p.Assets[i].Path < p.Assets[j].Path
	})
	for _, c := range m.Cases {
		for _, v := range m.Variants {
			for rep := 1; rep <= m.Execution.Repetitions; rep++ {
				p.Runs = append(p.Runs, PlannedRun{RunID: fmt.Sprintf("%s-%s-%03d", c.ID, v.ID, rep), CaseID: c.ID, VariantID: v.ID, Repetition: rep})
			}
		}
	}
	if m.Execution.RandomizeOrder {
		mathrand.New(mathrand.NewSource(seed)).Shuffle(len(p.Runs), func(i, j int) { p.Runs[i], p.Runs[j] = p.Runs[j], p.Runs[i] })
	}
	if len(m.Variants) > 1 {
		base := p.Manifest.Variants[0]
		for _, v := range p.Manifest.Variants[1:] {
			if v.Adapter != base.Adapter {
				p.Warnings = append(p.Warnings, "variants use different adapters")
			}
			if len(v.IntendedDifferences) == 0 {
				p.Warnings = append(p.Warnings, "variant "+v.ID+" has no declared intendedDifferences relative to "+base.ID)
			}
		}
	}
	p.PlanID, err = calculatePlanID(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func VerifyPlan(p *Plan) error {
	calculated, err := calculatePlanID(p)
	if err != nil {
		return err
	}
	if calculated != p.PlanID {
		return fmt.Errorf("plan drift: plan contents changed")
	}
	for _, a := range p.Assets {
		d, err := digestPathExcluding(a.Path, a.Excludes)
		if err != nil {
			return fmt.Errorf("plan drift: %s: %w", a.Path, err)
		}
		if d != a.SHA256 {
			return fmt.Errorf("plan drift: %s changed", a.Path)
		}
	}
	d, err := digestFile(p.ManifestPath)
	if err != nil {
		return fmt.Errorf("plan drift: manifest: %w", err)
	}
	if d != p.ManifestHash {
		return fmt.Errorf("plan drift: manifest changed")
	}
	return nil
}
func WritePlan(path string, p *Plan) error { return writeJSONAtomic(path, p) }
func ReadPlan(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Plan
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&p); err != nil {
		return nil, err
	}
	if p.SchemaVersion != PlanSchemaVersion {
		return nil, fmt.Errorf("unsupported plan schema version %d", p.SchemaVersion)
	}
	return &p, nil
}

func calculatePlanID(p *Plan) (string, error) {
	canonical := *p
	canonical.PlanID = ""
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256Bytes(data)), nil
}
func writeJSONAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.CreateTemp(filepath.Dir(path), ".bench-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func sha256Bytes(data []byte) []byte { h := sha256.New(); h.Write(data); return h.Sum(nil) }
