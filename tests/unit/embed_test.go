package unit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// expectedFiles is the exhaustive list of files that must be present after init.
var expectedFiles = []string{
	// Lite-tier compiled agents
	".github/agents/smaqit.create-agent.agent.md",
	".github/agents/smaqit.create-skill.agent.md",
}

// sourceDirMap maps installed path prefixes (relative to the init target dir)
// to source path prefixes (relative to the tests/unit/ package directory).
var sourceDirMap = []struct {
	installedDir string
	sourceDir    string
}{
	{".github/agents", "../../agents"},
}

// TestEmbedCompleteness verifies that every expected file is present after init.
func TestEmbedCompleteness(t *testing.T) {
	dir := t.TempDir()
	mustInit(t, dir)

	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected file not installed: %s", f)
		}
	}
}

// TestEmbedContentMatchesSource verifies that every installed file is
// byte-for-byte identical to its source in the repo root. This catches
// drift between make prepare output and the source artifacts.
func TestEmbedContentMatchesSource(t *testing.T) {
	dir := t.TempDir()
	mustInit(t, dir)

	for _, mapping := range sourceDirMap {
		installedAbs := filepath.Join(dir, mapping.installedDir)
		err := filepath.WalkDir(installedAbs, func(absPath string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(installedAbs, absPath)

			installed, readErr := os.ReadFile(absPath)
			if readErr != nil {
				t.Errorf("cannot read installed file %s/%s: %v", mapping.installedDir, relPath, readErr)
				return nil
			}

			sourcePath := filepath.Join(mapping.sourceDir, relPath)
			source, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Errorf("cannot read source file %s/%s: %v", mapping.sourceDir, relPath, readErr)
				return nil
			}

			if !bytes.Equal(installed, source) {
				t.Errorf("content mismatch: installed %s/%s differs from source %s/%s",
					mapping.installedDir, relPath, mapping.sourceDir, relPath)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("walking installed dir %s: %v", mapping.installedDir, err)
		}
	}
}
