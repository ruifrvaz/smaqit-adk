// tests deterministic expectation types and grading-copy isolation.
package bench

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicExpectationTypes(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "answer.txt"), "Hello\n")
	write(t, filepath.Join(root, "data.json"), `{"name":"bench","nested":{"ok":true}}`)
	write(t, filepath.Join(root, "tree", "one.txt"), "one")
	hash, _ := digestFile(filepath.Join(root, "answer.txt"))
	exit := 7
	expectations := []Expectation{
		{ID: "text", Type: "text", Actual: "stdout", Operator: "contains", Value: "hello", IgnoreCase: true},
		{ID: "file", Type: "file", Actual: "file:answer.txt", Operator: "sha256", SHA256: hash},
		{ID: "directory", Type: "directory", Actual: "directory:tree", Operator: "paths", RequiredPaths: []string{"one.txt"}, ForbiddenPaths: []string{"no.txt"}},
		{ID: "json", Type: "json", Actual: "file:data.json", Operator: "subset", Value: `{"nested":{"ok":true}}`},
		{ID: "runtime", Type: "runtime", Actual: "exitCode", ExitCode: &exit},
		{ID: "image", Type: "image", Actual: "file:answer.txt", Operator: "sha256", SHA256: hash},
	}
	request := RunRequest{Workspace: &Workspace{Root: root}, TraceDir: filepath.Join(root, "traces")}
	grades, err := gradeExpectations(context.Background(), expectations, HarnessResult{Stdout: "HELLO world", ExitCode: 7}, root, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, grade := range grades {
		if !grade.Passed {
			t.Errorf("%s failed: %s", grade.ID, grade.Message)
		}
	}
}

func TestJSONRejectsTrailingGarbage(t *testing.T) {
	var value any
	if err := decodeStrictJSON([]byte(`{} trailing`), &value); err == nil {
		t.Fatal("expected strict JSON error")
	}
}

func TestWorkspaceDoesNotExposeOracle(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	oracle := filepath.Join(root, "oracle.txt")
	write(t, input, "visible")
	write(t, oracle, "secret expected value")
	workspace, err := prepareWorkspace(Case{ID: "case", Given: Given{Prompt: Prompt{Text: "task"}, Files: []InputAsset{{ID: "input", Source: input}}}, Expect: []Expectation{{ID: "hidden", Type: "text", Actual: "stdout", ValueFile: oracle}}})
	if err != nil {
		t.Fatal(err)
	}
	defer removeWorkspace(workspace.Root)
	err = filepath.Walk(workspace.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if string(data) == "secret expected value" {
				t.Fatalf("oracle leaked into workspace at %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
