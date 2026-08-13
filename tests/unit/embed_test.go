package unit_test

import (
	"os"
	"path/filepath"
	"testing"
)

// expectedFiles is the exhaustive list of files that must be present after
// --install-global, relative to HOME.
var expectedFiles = []string{
	filepath.Join(".claude", "agents", "smaqit-L0.md"),
	filepath.Join(".claude", "agents", "smaqit-L1.md"),
	filepath.Join(".claude", "agents", "smaqit-L2.md"),
	filepath.Join(".codex", "agents", "smaqit.L0.toml"),
	filepath.Join(".codex", "agents", "smaqit.L1.toml"),
	filepath.Join(".codex", "agents", "smaqit.L2.toml"),
	filepath.Join(".agents", "skills", "smaqit.create-agent", "SKILL.md"),
	filepath.Join(".claude", "skills", "smaqit.create-agent", "SKILL.md"),
	filepath.Join(".agents", "smaqit-adk", "templates", "agents", "base-agent.template.md"),
	filepath.Join(".agents", "smaqit-adk", "framework", "SMAQIT.md"),
}

// TestEmbedCompleteness verifies that every expected file is present after
// a global install.
func TestEmbedCompleteness(t *testing.T) {
	home := mustInstallGlobal(t)

	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(home, f)); err != nil {
			t.Errorf("expected file not installed: %s", f)
		}
	}
}

// TestEmbedSkillsContentMatchesSource verifies that installed skill files
// (format-identical across platforms, so a single source of truth) are
// byte-for-byte identical to their source in the repo root. This catches
// drift between make prepare output and the source artifacts.
func TestEmbedSkillsContentMatchesSource(t *testing.T) {
	home := mustInstallGlobal(t)

	for _, installedRoot := range []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "skills"),
	} {
		err := filepath.WalkDir(installedRoot, func(absPath string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(installedRoot, absPath)

			installed, readErr := os.ReadFile(absPath)
			if readErr != nil {
				t.Errorf("cannot read installed file %s: %v", absPath, readErr)
				return nil
			}

			sourcePath := filepath.Join("..", "..", "skills", relPath)
			source, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Errorf("cannot read source file %s: %v", sourcePath, readErr)
				return nil
			}

			if string(installed) != string(source) {
				t.Errorf("content mismatch: installed %s differs from source %s", absPath, sourcePath)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("walking installed dir %s: %v", installedRoot, err)
		}
	}
}
