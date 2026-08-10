// tests deterministic run-matrix planning and reference drift detection.
package bench

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildPlanStableSeededMatrixAndDrift(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "before")
	seed := int64(42)
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	m.Execution.Seed = &seed
	m.Execution.Repetitions = 3
	m.Execution.RandomizeOrder = true
	p1, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := BuildPlan(manifestPath, m)
	if err != nil {
		t.Fatal(err)
	}
	if p1.PlanID != p2.PlanID || !reflect.DeepEqual(p1.Runs, p2.Runs) {
		t.Fatal("seeded plans differ")
	}
	write(t, filepath.Join(root, "input.txt"), "after")
	err = VerifyPlan(p1)
	var drift *DriftError
	if errors.As(err, &drift) {
		t.Fatal("VerifyPlan returns ordinary errors; RunPlan wraps drift")
	}
	if err == nil {
		t.Fatal("expected input drift")
	}
}

func TestReadPlanRejectsTampering(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "x")
	manifestPath := filepath.Join(root, "bench.yaml")
	write(t, manifestPath, validManifest("prompt:\n        text: hello"))
	m, _ := LoadManifest(manifestPath)
	p, _ := BuildPlan(manifestPath, m)
	planPath := filepath.Join(root, "plan.json")
	if err := WritePlan(planPath, p); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(planPath)
	data = []byte(string(data[:len(data)-2]) + `,"unknown":true}\n`)
	if err := os.WriteFile(planPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPlan(planPath); err == nil {
		t.Fatal("expected strict plan parsing failure")
	}
}
