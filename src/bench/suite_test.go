// tests suite discovery, validation, planning, and multi-manifest execution against fake harnesses.
package bench

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDiscoverManifestsFindsNestedBenchYAMLInSortedOrder(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "skills", "b-skill", "bench.yaml"), validManifest("prompt:\n        text: hello"))
	write(t, filepath.Join(root, "agents", "a-agent", "bench.yaml"), validManifest("prompt:\n        text: hello"))
	write(t, filepath.Join(root, "skills", "b-skill", "not-a-manifest.yaml"), "irrelevant")

	found, err := DiscoverManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 discovered manifests, got %d: %v", len(found), found)
	}
	if !strings.HasSuffix(found[0], filepath.Join("agents", "a-agent", "bench.yaml")) {
		t.Fatalf("expected agents/a-agent manifest first (sorted), got %s", found[0])
	}
	if !strings.HasSuffix(found[1], filepath.Join("skills", "b-skill", "bench.yaml")) {
		t.Fatalf("expected skills/b-skill manifest second (sorted), got %s", found[1])
	}
}

func TestDiscoverManifestsErrorsWhenRootIsEmpty(t *testing.T) {
	root := t.TempDir()
	found, err := DiscoverManifests(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Fatalf("expected no manifests under an empty root, got %v", found)
	}
	if _, err := ValidateSuite(root); err == nil {
		t.Fatal("expected ValidateSuite to error on a root with no manifests")
	}
	if _, err := RunSuite(context.Background(), root, SuiteOptions{}); err == nil {
		t.Fatal("expected RunSuite to error on a root with no manifests")
	}
}

func TestValidateSuiteReportsPerManifestErrorsWithoutStoppingOthers(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "good", "input.txt"), "visible")
	write(t, filepath.Join(root, "good", "bench.yaml"), validManifest("prompt:\n        text: hello"))
	write(t, filepath.Join(root, "bad", "bench.yaml"), "not: [valid yaml")

	result, err := ValidateSuite(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Fatal("expected suite validation to be invalid overall")
	}
	if len(result.Manifests) != 2 {
		t.Fatalf("expected 2 manifest results, got %d", len(result.Manifests))
	}
	var goodResult, badResult *SuiteManifestValidation
	for i := range result.Manifests {
		mv := &result.Manifests[i]
		if strings.Contains(mv.ManifestPath, "good") {
			goodResult = mv
		} else {
			badResult = mv
		}
	}
	if goodResult == nil || goodResult.Error != "" {
		t.Fatalf("expected the good manifest to validate cleanly: %+v", goodResult)
	}
	if goodResult.Name != "sample" || goodResult.Cases != 1 || goodResult.Variants != 1 {
		t.Fatalf("expected good manifest details populated: %+v", goodResult)
	}
	if badResult == nil || badResult.Error == "" {
		t.Fatalf("expected the bad manifest to record a validation error: %+v", badResult)
	}
}

func TestPlanSuiteWritesAPlanPerManifest(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "one", "input.txt"), "visible")
	write(t, filepath.Join(root, "one", "bench.yaml"), validManifest("prompt:\n        text: hello"))
	write(t, filepath.Join(root, "two", "input.txt"), "visible")
	write(t, filepath.Join(root, "two", "bench.yaml"), validManifest("prompt:\n        text: world"))

	result, err := PlanSuite(root)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Planned {
		t.Fatalf("expected suite planning to succeed: %+v", result)
	}
	if len(result.Manifests) != 2 {
		t.Fatalf("expected 2 planned manifests, got %d", len(result.Manifests))
	}
	for _, mp := range result.Manifests {
		if mp.Error != "" {
			t.Fatalf("manifest %s: unexpected planning error: %s", mp.ManifestPath, mp.Error)
		}
		if mp.PlanID == "" || mp.Runs == 0 {
			t.Fatalf("manifest %s: expected a populated plan: %+v", mp.ManifestPath, mp)
		}
		if _, err := os.Stat(mp.PlanPath); err != nil {
			t.Fatalf("manifest %s: expected plan file at %s: %v", mp.ManifestPath, mp.PlanPath, err)
		}
	}
}

func TestRunSuiteClassifiesPassingAndFailingManifestsIndependently(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "passes", "input.txt"), "visible")
	write(t, filepath.Join(root, "passes", "bench.yaml"), validManifest("prompt:\n        text: hello"))
	write(t, filepath.Join(root, "fails", "input.txt"), "visible")
	write(t, filepath.Join(root, "fails", "bench.yaml"), strings.ReplaceAll(validManifest("prompt:\n        text: hello"), "stdout: ok", "stdout: not-ok"))

	var events []string
	result, err := RunSuite(context.Background(), root, SuiteOptions{Observer: func(manifestPath string, event RunEvent) {
		events = append(events, filepath.Base(filepath.Dir(manifestPath))+":"+event.Type)
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed != 1 || result.Failed != 1 || result.Errored != 0 {
		t.Fatalf("expected 1 passed, 1 failed, 0 errored, got %+v", result)
	}
	if result.Eligible() {
		t.Fatal("expected the suite to be ineligible when any manifest fails")
	}
	if len(events) == 0 {
		t.Fatal("expected lifecycle events to be forwarded per manifest")
	}
	sawPasses, sawFails := false, false
	for _, e := range events {
		if strings.HasPrefix(e, "passes:") {
			sawPasses = true
		}
		if strings.HasPrefix(e, "fails:") {
			sawFails = true
		}
	}
	if !sawPasses || !sawFails {
		t.Fatalf("expected events prefixed with both manifest directories, got %v", events)
	}
}

func TestRunSuiteArtifactStagingBoundaryProducesAWinner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh/rm fixture is Unix-only")
	}
	root := t.TempDir()
	manifestDir := filepath.Join(root, "boundary")
	write(t, filepath.Join(manifestDir, "artifact.txt"), "artifact contents")
	write(t, filepath.Join(manifestDir, "bench.yaml"), `schemaVersion: 2
name: artifact-boundary
cases:
  - id: probe
    given:
      prompt:
        text: irrelevant
      files:
        - id: skill
          source: ./artifact.txt
          destination: artifact.txt
    expect:
      - id: sees-artifact
        type: text
        actual: stdout
        operator: contains
        value: present
variants:
  - id: with-artifact
    adapter: process
    process:
      executable: sh
      arguments: ["-c", "test -f {input:skill} && echo present || echo absent"]
      inputMode: stdin
  - id: without-artifact
    adapter: process
    process:
      executable: sh
      arguments: ["-c", "test -f {input:skill} && echo present || echo absent"]
      inputMode: stdin
    setup:
      - executable: chmod
        arguments: ["-R", "u+w", "{inputRoot}"]
      - executable: rm
        arguments: ["-f", "{input:skill}"]
    intendedDifferences:
      - Removes the staged artifact before the harness runs.
execution:
  repetitions: 1
  timeoutSeconds: 5
output:
  directory: ./results
`)

	result, err := RunSuite(context.Background(), root, SuiteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Manifests) != 1 || result.Manifests[0].Error != "" {
		t.Fatalf("expected one clean manifest result: %+v", result.Manifests)
	}
	experiment := result.Manifests[0].Experiment
	if experiment == nil {
		t.Fatal("expected an experiment to be recorded")
	}
	if experiment.Comparison.Outcome != "winner" || experiment.Comparison.Winner != "with-artifact" {
		t.Fatalf("expected with-artifact to win once the artifact is staged, got outcome=%s winner=%s", experiment.Comparison.Outcome, experiment.Comparison.Winner)
	}
	var withArtifactPassed, withoutArtifactPassed bool
	for _, r := range experiment.Results {
		switch r.VariantID {
		case "with-artifact":
			withArtifactPassed = r.RequiredPassed
		case "without-artifact":
			withoutArtifactPassed = r.RequiredPassed
		}
	}
	if !withArtifactPassed {
		t.Fatal("expected the with-artifact run to pass its required expectation (artifact present)")
	}
	if withoutArtifactPassed {
		t.Fatal("expected the without-artifact run to fail its required expectation (artifact removed by setup)")
	}
	// The suite still counts this manifest as passed: a "winner" outcome means
	// one variant is eligible, which is the expected shape for a comparison
	// manifest, not a suite-level failure.
	if result.Passed != 1 || result.Failed != 0 {
		t.Fatalf("expected the suite to classify a winner outcome as passed, got %+v", result)
	}
}
