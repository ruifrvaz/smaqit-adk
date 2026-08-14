//go:build ignore

// Renders smaqit-adk's shared-body agent sources (../agents/*.md) plus their
// per-platform metadata (../.smaqit/definitions/agents/*.frontmatter.yaml)
// into Claude Code (.md) and Codex CLI (.toml) compiled output under
// agents-claude/ and agents-codex/. Mirrors the split-source pattern
// validated by smaqit's scripts/generate-agents.py: one shared body per
// agent, never hand-duplicated across platforms. Invoked by `make prepare`
// via `go run generate-agents.go`; output is gitignored and regenerated on
// every build.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type platformMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tools       string `yaml:"tools,omitempty"`
}

type agentFrontmatter struct {
	Claude platformMeta `yaml:"claude"`
	Codex  platformMeta `yaml:"codex"`
}

var agentNames = []string{"smaqit.L0", "smaqit.L1", "smaqit.L2"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "generate-agents:", err)
		os.Exit(1)
	}
}

func run() error {
	claudeDir := "agents-claude"
	codexDir := "agents-codex"
	if err := os.RemoveAll(claudeDir); err != nil {
		return err
	}
	if err := os.RemoveAll(codexDir); err != nil {
		return err
	}
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return err
	}

	for _, name := range agentNames {
		body, err := os.ReadFile(filepath.Join("..", "agents", name+".md"))
		if err != nil {
			return fmt.Errorf("reading body for %s: %w", name, err)
		}
		fmContent, err := os.ReadFile(filepath.Join("..", ".smaqit", "definitions", "agents", name+".frontmatter.yaml"))
		if err != nil {
			return fmt.Errorf("reading frontmatter for %s: %w", name, err)
		}
		var fm agentFrontmatter
		if err := yaml.Unmarshal(fmContent, &fm); err != nil {
			return fmt.Errorf("parsing frontmatter for %s: %w", name, err)
		}

		bodyStr := strings.TrimRight(string(body), "\n") + "\n"

		if err := writeClaude(claudeDir, fm.Claude, bodyStr); err != nil {
			return fmt.Errorf("rendering claude output for %s: %w", name, err)
		}
		if err := writeCodex(codexDir, name, fm.Codex, bodyStr); err != nil {
			return fmt.Errorf("rendering codex output for %s: %w", name, err)
		}
	}

	fmt.Printf("generate-agents: rendered %d agents to %s/ and %s/\n", len(agentNames), claudeDir, codexDir)
	return nil
}

func writeClaude(dir string, meta platformMeta, body string) error {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: " + meta.Name + "\n")
	sb.WriteString("description: " + meta.Description + "\n")
	if meta.Tools != "" {
		sb.WriteString("tools: " + meta.Tools + "\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(body)

	dst := filepath.Join(dir, meta.Name+".md")
	return os.WriteFile(dst, []byte(sb.String()), 0644)
}

func writeCodex(dir string, filenameBase string, meta platformMeta, body string) error {
	if strings.Contains(body, "'''") {
		return fmt.Errorf("body contains a literal ''' sequence, which breaks TOML's literal multi-line string delimiter — cannot render %s", filenameBase)
	}

	var sb strings.Builder
	sb.WriteString("name = " + toTOMLString(meta.Name) + "\n")
	sb.WriteString("description = " + toTOMLString(meta.Description) + "\n")
	sb.WriteString("developer_instructions = '''\n")
	sb.WriteString(body)
	sb.WriteString("'''\n")

	dst := filepath.Join(dir, filenameBase+".toml")
	return os.WriteFile(dst, []byte(sb.String()), 0644)
}

func toTOMLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
