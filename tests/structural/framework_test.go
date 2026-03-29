package structural

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestFrameworkNoDirectiveLanguage(t *testing.T) {
	// Framework files describe principles and invariants (declarative).
	// Directive language (MUST, MUST NOT, SHOULD) belongs in L1 compiled rules.
	directiveRe := regexp.MustCompile(`\b(MUST NOT|MUST|SHOULD)\b`)

	frameworkDir := filepath.Join(sourceRoot, "framework")
	entries, err := os.ReadDir(frameworkDir)
	if err != nil {
		t.Fatalf("read framework dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(frameworkDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			for i, line := range strings.Split(string(data), "\n") {
				if m := directiveRe.FindString(line); m != "" {
					t.Errorf("line %d contains directive keyword %q: %s", i+1, m, strings.TrimSpace(line))
				}
			}
		})
	}
}
