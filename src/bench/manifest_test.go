// tests manifest parsing, validation diagnostics, and input safety rules.
package bench

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestStrictAndManifestRelative(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "prompt.txt"), "do it")
	write(t, filepath.Join(root, "input.txt"), "input")
	manifest := validManifest("prompt:\n        file: ./prompt.txt")
	manifest += "\nmisspelled: true\n"
	path := filepath.Join(root, "bench.yaml")
	write(t, path, manifest)
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "misspelled") {
		t.Fatalf("expected unknown field diagnostic, got %v", err)
	}
	write(t, path, validManifest("prompt:\n        file: ./prompt.txt"))
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Cases[0].Given.Prompt.File != filepath.Join(root, "prompt.txt") {
		t.Fatalf("path was not manifest-relative: %s", m.Cases[0].Given.Prompt.File)
	}
}

func TestUnknownFieldReportsNestedPath(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "input")
	body := strings.Replace(validManifest("prompt:\n        text: hello"), "stdout: ok", "stdout: ok\n      misspelled: true", 1)
	path := filepath.Join(root, "bench.yaml")
	write(t, path, body)
	_, err := LoadManifest(path)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(validation.Diagnostics) != 1 || validation.Diagnostics[0].Path != "variants[0].mock.misspelled" {
		t.Fatalf("unexpected diagnostics: %+v", validation.Diagnostics)
	}
}

func TestLoadManifestRejectsEscapesAndWeights(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bench.yaml")
	body := validManifest("prompt:\n        text: hello")
	body = strings.Replace(body, "source: ./input.txt", "source: ./input.txt\n          destination: ../escape", 1)
	body += `graders:
  - id: quality
    type: command
    weight: 0.9
    command:
      executable: true
`
	write(t, filepath.Join(root, "input.txt"), "x")
	write(t, path, body)
	_, err := LoadManifest(path)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	message := err.Error()
	for _, want := range []string{"cases[0].given.files[0].destination", "graders"} {
		if !strings.Contains(message, want) {
			t.Errorf("missing %s in %s", want, message)
		}
	}
}

func TestManifestV2RejectsLegacyTaskPlaceholders(t *testing.T) {
	for _, test := range []struct{ legacy, replacement string }{{"{task}", "{brief}"}, {"{taskFile}", "{briefFile}"}} {
		m := Manifest{
			SchemaVersion: ManifestSchemaVersion,
			Name:          "legacy-placeholder",
			Cases:         []Case{{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}}}},
			Variants: []Variant{{ID: "process", Adapter: "process", Process: &ProcessConfig{
				Executable: "echo", Arguments: []string{test.legacy}, InputMode: "argument",
			}}},
			Execution: Execution{Repetitions: 1, TimeoutSeconds: 5},
			Output:    Output{Directory: t.TempDir()},
		}
		want := "legacy placeholder " + test.legacy + "; schema v2 uses " + test.replacement
		var found bool
		for _, diagnostic := range m.Validate() {
			if diagnostic.Path == "variants[0].process.arguments[0]" && diagnostic.Message == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("legacy placeholder %s was not rejected: %+v", test.legacy, m.Validate())
		}
	}
}

func TestManifestRejectsVariantSetupAndInvalidTreatmentContract(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "input.txt"), "input")
	path := filepath.Join(root, "bench.yaml")
	body := strings.Replace(validManifest("prompt:\n        text: hello"), "    mock:\n      stdout: ok", "    setup:\n      - executable: true\n    mock:\n      stdout: ok", 1)
	write(t, path, body)
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "variants[0].setup: unknown field") {
		t.Fatalf("expected variant setup rejection, got %v", err)
	}

	treatment := filepath.Join(root, "treatment.txt")
	write(t, treatment, "guidance")
	body = strings.Replace(validManifest("prompt:\n        text: hello"), "    mock:\n      stdout: ok", "    treatment:\n      - id: guide\n        source: ./treatment.txt\n        destination: project/guide\n    intendedDifferences: [guide]\n    mock:\n      stdout: ok", 1)
	write(t, path, body)
	_, err = LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "variants[0].treatment[0].destination: unknown field") {
		t.Fatalf("expected treatment destination rejection, got %v", err)
	}

	m := Manifest{
		SchemaVersion: ManifestSchemaVersion, Name: "treatment-validation",
		Cases:     []Case{{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", Value: "ok"}}}},
		Variants:  []Variant{{ID: "with", Adapter: "process", Treatment: []TreatmentAsset{{ID: "guide", Source: treatment}}, Process: &ProcessConfig{Executable: "echo", Arguments: []string{"{input:missing}", "{treatment:missing}"}, InputMode: "argument"}}},
		Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
	}
	m.Variants[0].IntendedDifferences = []string{""}
	message := (&ValidationError{Diagnostics: m.Validate()}).Error()
	for _, want := range []string{"intendedDifferences[0]: must not be blank", "intendedDifferences: is required", "references undeclared input missing", "references unavailable treatment missing"} {
		if !strings.Contains(message, want) {
			t.Errorf("missing %q in %s", want, message)
		}
	}
}

func TestManifestRejectsReservedFixtureDestination(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture")
	write(t, filepath.Join(fixture, "file.txt"), "fixture")
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion, Name: "fixture-destination",
		Cases:    []Case{{ID: "case", Fixture: &SourceRef{Source: fixture, Destination: inputDirectoryName + "/nested"}, Given: Given{Prompt: Prompt{Text: "hello"}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", Value: "ok"}}}},
		Variants: []Variant{{ID: "mock", Adapter: "mock", Mock: &MockConfig{Stdout: "ok"}}}, Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
	}
	if message := (&ValidationError{Diagnostics: m.Validate()}).Error(); !strings.Contains(message, "must not overlap the reserved Bench sidecar") {
		t.Fatalf("expected reserved destination rejection, got %s", message)
	}
}

func TestManifestRejectsBenchManagedInputDestinations(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	write(t, input, "input")
	for _, destination := range []string{".", "brief.md/nested", "treatment/guide"} {
		m := Manifest{
			SchemaVersion: ManifestSchemaVersion, Name: "reserved-input-destination",
			Cases:    []Case{{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}, Files: []InputAsset{{ID: "input", Source: input, Destination: destination}}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", Value: "ok"}}}},
			Variants: []Variant{{ID: "mock", Adapter: "mock", Mock: &MockConfig{Stdout: "ok"}}}, Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
		}
		if message := (&ValidationError{Diagnostics: m.Validate()}).Error(); !strings.Contains(message, "must not overlap Bench-managed sidecar paths") {
			t.Fatalf("destination %q was accepted: %s", destination, message)
		}
	}
}

func TestManifestDetectsEffectiveInputDestinationCollision(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "docs")
	write(t, filepath.Join(directory, "readme.md"), "docs")
	file := filepath.Join(root, "input.txt")
	write(t, file, "input")
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion, Name: "input-destination-collision",
		Cases:    []Case{{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}, Directories: []InputAsset{{ID: "docs", Source: directory}}, Files: []InputAsset{{ID: "input", Source: file, Destination: "directories/docs"}}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", Value: "ok"}}}},
		Variants: []Variant{{ID: "mock", Adapter: "mock", Mock: &MockConfig{Stdout: "ok"}}}, Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
	}
	if message := (&ValidationError{Diagnostics: m.Validate()}).Error(); !strings.Contains(message, "must be unique within the case") {
		t.Fatalf("expected effective destination collision, got %s", message)
	}
}

func TestManifestRejectsTreatmentOverlapAndHiddenOracleLeakage(t *testing.T) {
	root := t.TempDir()
	container := filepath.Join(root, "container")
	fixture := filepath.Join(container, "fixture")
	write(t, filepath.Join(fixture, "project.md"), "project")
	oracle := filepath.Join(container, "golden.txt")
	write(t, oracle, "ok")
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion, Name: "treatment-overlap",
		Cases:    []Case{{ID: "case", Fixture: &SourceRef{Source: fixture}, Given: Given{Prompt: Prompt{Text: "hello"}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", ValueFile: oracle}}}},
		Variants: []Variant{{ID: "with", Adapter: "mock", Treatment: []TreatmentAsset{{ID: "guide", Source: container}}, Mock: &MockConfig{Stdout: "ok"}, IntendedDifferences: []string{"Exposes guidance."}}}, Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
	}
	message := (&ValidationError{Diagnostics: m.Validate()}).Error()
	for _, want := range []string{"must be outside cases[0].fixture", "must not contain hidden oracle cases[0].expect[0].valueFile"} {
		if !strings.Contains(message, want) {
			t.Errorf("missing %q in %s", want, message)
		}
	}
}

func TestLoadManifestRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "real.txt"), "x")
	if err := os.Symlink("real.txt", filepath.Join(root, "input.txt")); err != nil {
		t.Skip(err)
	}
	path := filepath.Join(root, "bench.yaml")
	write(t, path, validManifest("prompt:\n        text: hello"))
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestDirectoryAssetsRejectNonRegularDescendants(t *testing.T) {
	mkfifo, err := exec.LookPath("mkfifo")
	if err != nil {
		t.Skip("mkfifo is unavailable")
	}
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture")
	if err := os.MkdirAll(fixture, 0755); err != nil {
		t.Fatal(err)
	}
	pipe := filepath.Join(fixture, "blocked.pipe")
	if output, err := exec.Command(mkfifo, pipe).CombinedOutput(); err != nil {
		t.Skipf("cannot create FIFO: %v: %s", err, output)
	}
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion, Name: "special-file",
		Cases:    []Case{{ID: "case", Fixture: &SourceRef{Source: fixture}, Given: Given{Prompt: Prompt{Text: "hello"}}, Expect: []Expectation{{ID: "ok", Type: "text", Actual: "stdout", Value: "ok"}}}},
		Variants: []Variant{{ID: "mock", Adapter: "mock", Mock: &MockConfig{Stdout: "ok"}}}, Execution: Execution{Repetitions: 1, TimeoutSeconds: 5}, Output: Output{Directory: filepath.Join(root, "results")},
	}
	if message := (&ValidationError{Diagnostics: m.Validate()}).Error(); !strings.Contains(message, "contains non-regular file") {
		t.Fatalf("expected non-regular descendant rejection, got %s", message)
	}
	if _, err := digestPath(fixture); err == nil || !strings.Contains(err.Error(), "non-regular files are not allowed") {
		t.Fatalf("expected hashing defense, got %v", err)
	}
	if err := copyDirectory(fixture, filepath.Join(root, "copy"), nil); err == nil || !strings.Contains(err.Error(), "non-regular files are not allowed") {
		t.Fatalf("expected copying defense, got %v", err)
	}
}

func TestOutputMustBeOutsideFixture(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture")
	write(t, filepath.Join(fixture, "input.txt"), "x")
	body := strings.Replace(validManifest("prompt:\n        text: hello"), "source: ./input.txt", "source: ./fixture/input.txt", 1)
	body = strings.Replace(body, "given:\n", "fixture:\n      source: ./fixture\n    given:\n", 1)
	body = strings.Replace(body, "directory: ./results", "directory: ./fixture/results", 1)
	path := filepath.Join(root, "bench.yaml")
	write(t, path, body)
	_, err := LoadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "outside fixture") {
		t.Fatalf("expected output isolation error, got %v", err)
	}
}

func validManifest(prompt string) string {
	return `schemaVersion: 2
name: sample
cases:
  - id: case-1
    given:
      ` + prompt + `
      files:
        - id: input
          source: ./input.txt
    expect:
      - id: says-ok
        type: text
        actual: stdout
        operator: exact
        value: ok
variants:
  - id: mock
    adapter: mock
    mock:
      stdout: ok
execution:
  repetitions: 1
  timeoutSeconds: 5
output:
  directory: ./results
`
}
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
