package structural

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// templateRulesMapping maps each template filename (basename) to the ordered list of compiled
// rules files (relative to the template's directory) that collectively define its placeholders.
// Extension templates (implementation, specification) inherit base placeholders, so base.rules.md
// is included first as the fallback lookup.
var templateRulesMapping = map[string][]string{
	"base-agent.template.md":           {"compiled/base.rules.md"},
	"implementation-agent.template.md": {"compiled/base.rules.md", "compiled/implementation.rules.md"},
	"specification-agent.template.md":  {"compiled/base.rules.md", "compiled/specification.rules.md"},
	"base-skill.template.md":           {"compiled/skill.rules.md"},
}

func TestTemplatePlaceholdersDefinedInRules(t *testing.T) {
	placeholderRe := regexp.MustCompile(`\[([A-Z][A-Z_]+)\]`)

	templateDirs := []string{
		filepath.Join(sourceRoot, "templates", "agents"),
		filepath.Join(sourceRoot, "templates", "skills"),
	}

	for _, dir := range templateDirs {
		for _, path := range findFiles(t, dir, ".template.md") {
			baseName := filepath.Base(path)
			t.Run(baseName, func(t *testing.T) {
				rulesFiles, ok := templateRulesMapping[baseName]
				if !ok {
					t.Fatalf("no rules file mapping for template %q — add it to templateRulesMapping", baseName)
				}

				// Build combined rules content from all applicable rules files.
				var combined strings.Builder
				for _, rulesBase := range rulesFiles {
					rulesPath := filepath.Join(filepath.Dir(path), rulesBase)
					data, err := os.ReadFile(rulesPath)
					if err != nil {
						t.Fatalf("read rules %s: %v", rulesPath, err)
					}
					combined.Write(data)
					combined.WriteByte('\n')
				}
				rulesContent := combined.String()

				templateData, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read template %s: %v", path, err)
				}
				for _, m := range placeholderRe.FindAllStringSubmatch(string(templateData), -1) {
					placeholder := m[0] // e.g., "[AGENT_NAME]"
					if !strings.Contains(rulesContent, placeholder) {
						t.Errorf("placeholder %q in template is not defined in any of its rules files: %v", placeholder, rulesFiles)
					}
				}
			})
		}
	}
}
