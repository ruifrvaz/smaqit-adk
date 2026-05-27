// validate-skill validates a single SKILL.md file against the ADK skill format rules.
//
// Usage: go run validate-skill.go <path-to-SKILL.md>
// Exit codes: 0 = valid, 1 = invalid or usage error.
//
// Rules enforced (mirrors tests/structural/skills_test.go):
//   - Frontmatter: name format, description presence + length, no first-person or anti-patterns
//   - Required sections: Steps, Output, Scope, Completion, Failure Handling
//   - Body length: max 500 lines
//   - No unresolved uppercase placeholders (e.g. [SKILL_NAME])
//   - Failure Handling table: header + separator + ≥2 data rows

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run validate-skill.go <path-to-SKILL.md>")
		os.Exit(1)
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		os.Exit(1)
	}

	violations := validateSkill(string(data))
	if len(violations) == 0 {
		fmt.Printf("✓ %s passed all checks\n", path)
		os.Exit(0)
	}
	fmt.Printf("✗ %s has %d violation(s):\n", path, len(violations))
	for _, v := range violations {
		fmt.Printf("  - %s\n", v)
	}
	os.Exit(1)
}

func validateSkill(content string) []string {
	var violations []string

	// Parse frontmatter
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return []string{"missing frontmatter between --- markers"}
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
	body := parts[2]

	// Validate name
	nameRe := regexp.MustCompile(`^[a-z][a-z0-9.-]*$`)
	name := fm["name"]
	if name == "" {
		violations = append(violations, "missing 'name' field in frontmatter")
	} else if !nameRe.MatchString(name) {
		violations = append(violations, fmt.Sprintf("name %q does not match ^[a-z][a-z0-9.-]*$", name))
	}

	// Validate description
	desc := strings.Trim(fm["description"], `"'`)
	if desc == "" {
		violations = append(violations, "missing 'description' field in frontmatter")
	} else {
		if len(desc) > 1024 {
			violations = append(violations, fmt.Sprintf("description length %d exceeds 1024 characters", len(desc)))
		}
		firstPersonRe := regexp.MustCompile(`(?i)^(I |You can)`)
		if firstPersonRe.MatchString(desc) {
			n := 60
			if n > len(desc) {
				n = len(desc)
			}
			violations = append(violations, fmt.Sprintf("description must not use first-person phrasing: %q", desc[:n]))
		}
		antiPatternRe := regexp.MustCompile(`(?i)(Use when|Use this skill when|when the user|helps users|allow.*user to|This skill (can|will|may))`)
		if m := antiPatternRe.FindString(desc); m != "" {
			violations = append(violations, fmt.Sprintf(
				"description contains conversational/user-centric anti-pattern %q — use capability-oriented, declarative, present-tense phrasing instead", m))
		}
	}

	// Validate required sections
	for _, section := range []string{"## Steps", "## Output", "## Scope", "## Completion", "## Failure Handling"} {
		if found, _ := findSection(body, section); !found {
			violations = append(violations, fmt.Sprintf("missing required section %q", section))
		}
	}

	// Validate body length
	lines := strings.Split(body, "\n")
	if len(lines) > 500 {
		violations = append(violations, fmt.Sprintf("body has %d lines, maximum is 500", len(lines)))
	}

	// Validate no unresolved placeholders (skip code blocks)
	placeholderRe := regexp.MustCompile(`\[[A-Z][A-Z_]+\]`)
	inCode := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		if m := placeholderRe.FindString(line); m != "" {
			violations = append(violations, fmt.Sprintf("unresolved placeholder %q on body line %d", m, i+1))
		}
	}

	// Validate Failure Handling table structure
	if found, fhLines := findSection(body, "## Failure Handling"); found {
		rows := 0
		for _, line := range fhLines {
			if strings.Contains(line, "|") {
				rows++
			}
		}
		// A valid table needs: header row + separator row + ≥2 data rows = ≥4 pipe-lines.
		if rows < 4 {
			violations = append(violations, fmt.Sprintf(
				"Failure Handling table has %d pipe-rows (need ≥4: header, separator, 2+ data rows)", rows))
		}
	}

	return violations
}

// findSection returns the content lines of a named section heading (e.g. "## Completion"),
// skipping headings that appear inside fenced code blocks.
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
