// tests manifest parsing, validation diagnostics, and input safety rules.
package bench

import (
	"errors"
	"os"
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
	m := Manifest{
		SchemaVersion: ManifestSchemaVersion,
		Name:          "legacy-placeholder",
		Cases:         []Case{{ID: "case", Given: Given{Prompt: Prompt{Text: "hello"}}}},
		Variants: []Variant{{ID: "process", Adapter: "process", Process: &ProcessConfig{
			Executable: "echo", Arguments: []string{"{taskFile}"}, InputMode: "argument",
		}}},
		Execution: Execution{Repetitions: 1, TimeoutSeconds: 5},
		Output:    Output{Directory: t.TempDir()},
	}
	var found bool
	for _, diagnostic := range m.Validate() {
		if diagnostic.Path == "variants[0].process.arguments[0]" && strings.Contains(diagnostic.Message, "unsupported placeholder {taskFile}") {
			found = true
		}
	}
	if !found {
		t.Fatalf("legacy placeholder was not rejected: %+v", m.Validate())
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
