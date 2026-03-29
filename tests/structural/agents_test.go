package structural

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentFrontmatter(t *testing.T) {
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "agents"), ".agent.md") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			fm, _ := parseMarkdownFM(t, path)
			if fm["name"] == "" {
				t.Error("missing 'name' field in frontmatter")
			}
			tools := strings.Trim(strings.TrimSpace(fm["tools"]), "[]")
			if tools == "" {
				t.Error("'tools' field is missing or empty")
			}
		})
	}
}

func TestAgentRequiredSections(t *testing.T) {
	required := []string{
		"## Role", "## Input", "## Output", "## Directives", "## Completion Criteria",
	}
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "agents"), ".agent.md") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			for _, section := range required {
				if found, _ := findSection(body, section); !found {
					t.Errorf("missing required section %q", section)
				}
			}
		})
	}
}

func TestAgentCompletionCriteria(t *testing.T) {
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "agents"), ".agent.md") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			found, lines := findSection(body, "## Completion Criteria")
			if !found {
				t.Fatal("missing ## Completion Criteria section")
			}
			hasCheckbox := false
			for _, line := range lines {
				if strings.Contains(line, "- [ ]") {
					hasCheckbox = true
					break
				}
			}
			if !hasCheckbox {
				t.Error("Completion Criteria section has no '- [ ]' checklist items")
			}
		})
	}
}

func TestAgentDirectivesHasMust(t *testing.T) {
	for _, path := range findFiles(t, filepath.Join(sourceRoot, "agents"), ".agent.md") {
		t.Run(filepath.Base(path), func(t *testing.T) {
			_, body := parseMarkdownFM(t, path)
			found, lines := findSection(body, "## Directives")
			if !found {
				t.Fatal("missing ## Directives section")
			}
			hasMust := false
			for _, line := range lines {
				if strings.TrimSpace(line) == "### MUST" {
					hasMust = true
					break
				}
			}
			if !hasMust {
				t.Error("Directives section has no '### MUST' subsection")
			}
		})
	}
}
