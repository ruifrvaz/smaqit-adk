package structural

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const sourceRoot = "../.."

// parseMarkdownFM splits a file with --- delimiters into a frontmatter key:value
// map and the body text after the closing ---.
func parseMarkdownFM(t *testing.T, path string) (map[string]string, string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		t.Fatalf("%s: expected frontmatter between --- markers, got %d parts", path, len(parts))
	}
	fm := make(map[string]string)
	for _, line := range strings.Split(parts[1], "\n") {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		fm[key] = val
	}
	return fm, parts[2]
}

// findFiles walks dir recursively and returns paths of files whose name ends with suffix.
func findFiles(t *testing.T, dir, suffix string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), suffix) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return paths
}

func TestSkillFrontmatter(t *testing.T) {
	nameRe := regexp.MustCompile(`^[a-z][a-z0-9.-]*$`)
	firstPersonRe := regexp.MustCompile(`(?i)^(I |You can)`)
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "skills"), "SKILL.md") {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			fm, _ := parseMarkdownFM(t, path)
			if fm["name"] == "" {
				t.Error("missing 'name' field in frontmatter")
			} else if !nameRe.MatchString(fm["name"]) {
				t.Errorf("name %q does not match ^[a-z][a-z0-9.-]*$", fm["name"])
			}
			desc := strings.Trim(fm["description"], `"'`)
			if desc == "" {
				t.Error("missing 'description' field in frontmatter")
			} else {
				if len(desc) > 1024 {
					t.Errorf("description length %d exceeds 1024 characters", len(desc))
				}
				if firstPersonRe.MatchString(desc) {
					t.Errorf("description must be written in third person, got: %q", desc[:min(60, len(desc))])
				}
				hasWhenSignal := strings.Contains(desc, "Use when") || strings.Contains(desc, "when the user")
				if !hasWhenSignal {
					t.Error("description missing when-signal: must contain \"Use when\" or \"when the user\"")
				}
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// findSection returns the content lines of the named section heading (e.g., "## Completion Criteria"),
// tracking fenced code blocks so headings inside code blocks are not treated as section boundaries.
// Returns (found, contentLines).
func findSection(body, heading string) (bool, []string) {
	inCode := false
	inSection := false
	var content []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			if inSection {
				content = append(content, line)
			}
			continue
		}
		if inCode {
			if inSection {
				content = append(content, line)
			}
			continue
		}
		if trimmed == heading {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inSection {
			content = append(content, line)
		}
	}
	return inSection, content
}

func TestSkillRequiredSections(t *testing.T) {
	required := []string{
		"## Purpose", "## Steps", "## Output",
		"## Scope", "## Completion", "## Failure Handling",
	}
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "skills"), "SKILL.md") {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			for _, section := range required {
				if found, _ := findSection(body, section); !found {
					t.Errorf("missing required section %q", section)
				}
			}
		})
	}
}

func TestSkillBodyLength(t *testing.T) {
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "skills"), "SKILL.md") {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			lines := strings.Split(body, "\n")
			if len(lines) > 500 {
				t.Errorf("body has %d lines, maximum is 500", len(lines))
			}
		})
	}
}

func TestSkillNoUnresolvedPlaceholders(t *testing.T) {
	// Matches uppercase placeholders like [AGENT_NAME], [TOOL_LIST] — resolved
	// skills must not contain these. Code blocks are skipped.
	placeholderRe := regexp.MustCompile(`\[[A-Z][A-Z_]+\]`)
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "skills"), "SKILL.md") {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			inCode := false
			for i, line := range strings.Split(body, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "```") {
					inCode = !inCode
					continue
				}
				if inCode {
					continue
				}
				if m := placeholderRe.FindString(line); m != "" {
					t.Errorf("unresolved placeholder %q on body line %d", m, i+1)
				}
			}
		})
	}
}

func TestSkillFailureHandlingTable(t *testing.T) {
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "skills"), "SKILL.md") {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			found, lines := findSection(body, "## Failure Handling")
			if !found {
				t.Fatal("missing ## Failure Handling section")
			}
			rows := 0
			for _, line := range lines {
				if strings.Contains(line, "|") {
					rows++
				}
			}
			// A valid table needs: header row + separator row + ≥2 data rows = ≥4 pipe-lines.
			if rows < 4 {
				t.Errorf("Failure Handling table has %d pipe-rows (need ≥4: header, separator, 2+ data rows)", rows)
			}
		})
	}
}
