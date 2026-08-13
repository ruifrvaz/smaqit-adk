package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ruifrvaz/smaqit-adk/src/benchcli"
)

//go:embed agents-claude/*.md
var claudeAgentFiles embed.FS

//go:embed agents-codex/*.toml
var codexAgentFiles embed.FS

//go:embed skills
var adkSkillsFS embed.FS

//go:embed framework
var adkFrameworkFS embed.FS

//go:embed templates
var adkTemplatesFS embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "1.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bench":
		os.Exit(benchcli.Run(os.Args[2:]))
	case "--install-global":
		// Internal trigger for install.sh only — never documented or
		// presented as a user-facing command. install.sh is the sole
		// global-install entry point.
		installGlobal()
	case "help", "--help", "-h":
		cmdHelp()
	case "uninstall":
		cmdUninstall()
	case "version", "--version", "-v":
		fmt.Printf("smaqit-adk %s\n", Version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`smaqit-adk - Agent Development Kit

Usage: smaqit-adk <command>

Commands:
  bench <command>   Validate, plan, run, grade, compare, or report evaluations
  help              Show detailed help
  uninstall         Remove smaqit-adk's globally installed agents, skills, templates, and framework
  version           Show smaqit-adk version`)
}

func cmdHelp() {
	fmt.Println("smaqit-adk - Agent Development Kit")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("  smaqit-adk bench <validate|plan|run|grade|compare|report>")
	fmt.Println("      Run config-first local agent evaluations and benchmarks")
	fmt.Println("      Use 'smaqit-adk bench --help' for the manifest lifecycle")
	fmt.Println()
	fmt.Println("  smaqit-adk uninstall")
	fmt.Println("      Remove smaqit-adk's globally installed agents, skills, templates, and framework")
	fmt.Println()
	fmt.Println("  smaqit-adk version   Show smaqit-adk version")
	fmt.Println()
	fmt.Println("Installation:")
	fmt.Println("  curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit-adk/main/install.sh | bash")
	fmt.Println("  Installs smaqit.L0/L1/L2 to ~/.claude/agents/ and ~/.codex/agents/,")
	fmt.Println("  the 5 ADK skills to ~/.agents/skills/ and ~/.claude/skills/,")
	fmt.Println("  and compilation templates/framework to ~/.agents/smaqit-adk/.")
	fmt.Println("  Nothing is written into any project directory.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Say 'create a new agent' or use /smaqit.create-agent")
	fmt.Println("  Say 'create a new skill' or use /smaqit.create-skill")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/ruifrvaz/smaqit-adk")
}

// homeDir resolves the current user's home directory or exits with an error.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error resolving home directory: %v\n", err)
		os.Exit(1)
	}
	return h
}

// globalAgentsDir resolves the global agents directory for the given
// platform ("claude" or "codex"), respecting each platform's own
// environment-variable override.
func globalAgentsDir(platform string) string {
	switch platform {
	case "claude":
		if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
			return filepath.Join(d, "agents")
		}
		return filepath.Join(homeDir(), ".claude", "agents")
	case "codex":
		if d := os.Getenv("CODEX_HOME"); d != "" {
			return filepath.Join(d, "agents")
		}
		return filepath.Join(homeDir(), ".codex", "agents")
	default:
		panic("unknown platform: " + platform)
	}
}

// globalSkillsDirs resolves every global skills directory smaqit-adk
// installs into: the shared Copilot+Codex path, and Claude's own path.
func globalSkillsDirs() []string {
	shared := filepath.Join(homeDir(), ".agents", "skills")
	var claude string
	if d := os.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		claude = filepath.Join(d, "skills")
	} else {
		claude = filepath.Join(homeDir(), ".claude", "skills")
	}
	return []string{shared, claude}
}

// globalDataDir resolves smaqit-adk's own namespaced global data directory
// for compilation templates and framework principle files.
func globalDataDir() string {
	return filepath.Join(homeDir(), ".agents", "smaqit-adk")
}

func installGlobal() {
	if err := copyEmbedDir(claudeAgentFiles, "agents-claude", globalAgentsDir("claude")); err != nil {
		fmt.Printf("Error installing Claude agents: %v\n", err)
		os.Exit(1)
	}
	if err := copyEmbedDir(codexAgentFiles, "agents-codex", globalAgentsDir("codex")); err != nil {
		fmt.Printf("Error installing Codex agents: %v\n", err)
		os.Exit(1)
	}
	for _, dst := range globalSkillsDirs() {
		if err := copyEmbedDir(adkSkillsFS, "skills", dst); err != nil {
			fmt.Printf("Error installing skills to %s: %v\n", dst, err)
			os.Exit(1)
		}
	}
	dataDir := globalDataDir()
	if err := copyEmbedDir(adkTemplatesFS, "templates", filepath.Join(dataDir, "templates")); err != nil {
		fmt.Printf("Error installing templates: %v\n", err)
		os.Exit(1)
	}
	if err := copyEmbedDir(adkFrameworkFS, "framework", filepath.Join(dataDir, "framework")); err != nil {
		fmt.Printf("Error installing framework: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ smaqit-adk %s installed globally\n", Version)
	fmt.Println("Use /smaqit.create-agent, /smaqit.create-skill, /smaqit.new-principle, /smaqit.bench-run, and /smaqit.bench-scaffold in Claude Code or Codex CLI.")
}

// copyEmbedDir copies all files from an embed.FS rooted at src into the dst directory on disk.
func copyEmbedDir(fsys embed.FS, src, dst string) error {
	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("resolving relative path for %s: %w", path, err)
		}
		dstPath := filepath.Join(dst, relPath)
		if d.IsDir() {
			if mkErr := os.MkdirAll(dstPath, 0755); mkErr != nil {
				return fmt.Errorf("creating directory %s: %w", dstPath, mkErr)
			}
			return nil
		}
		content, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, err)
		}
		if writeErr := os.WriteFile(dstPath, content, 0644); writeErr != nil {
			return fmt.Errorf("writing file %s: %w", dstPath, writeErr)
		}
		return nil
	})
}

// embedFileNames returns the top-level file names (not directories) directly under root.
func embedFileNames(fsys embed.FS, root string) []string {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// embedDirNames returns the top-level directory names directly under root.
func embedDirNames(fsys embed.FS, root string) []string {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// uninstallTarget is a single named path this uninstall touches, plus
// whether it's a directory (RemoveAll) or a file (Remove).
type uninstallTarget struct {
	path  string
	isDir bool
}

func cmdUninstall() {
	claudeAgentsDir := globalAgentsDir("claude")
	codexAgentsDir := globalAgentsDir("codex")
	skillsDirs := globalSkillsDirs()
	dataDir := globalDataDir()

	claudeNames := embedFileNames(claudeAgentFiles, "agents-claude")
	codexNames := embedFileNames(codexAgentFiles, "agents-codex")
	skillNames := embedDirNames(adkSkillsFS, "skills")

	var candidates []uninstallTarget
	for _, n := range claudeNames {
		candidates = append(candidates, uninstallTarget{filepath.Join(claudeAgentsDir, n), false})
	}
	for _, n := range codexNames {
		candidates = append(candidates, uninstallTarget{filepath.Join(codexAgentsDir, n), false})
	}
	for _, dst := range skillsDirs {
		for _, n := range skillNames {
			candidates = append(candidates, uninstallTarget{filepath.Join(dst, n), true})
		}
	}
	candidates = append(candidates, uninstallTarget{dataDir, true})

	// Only prompt for targets that actually exist — an idempotent uninstall
	// on a machine with nothing installed shouldn't ask the user to confirm
	// removing nothing.
	var targets []uninstallTarget
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err == nil {
			targets = append(targets, c)
		}
	}
	if len(targets) == 0 {
		fmt.Println("No smaqit-adk installation found")
		os.Exit(0)
	}

	// Skills directories are shared with other tools/products (e.g. smaqit,
	// smaqit-extensions) — only named smaqit-adk entries are ever touched,
	// never the directory itself.
	fmt.Println("This will remove smaqit-adk's globally installed artifacts:")
	for _, t := range targets {
		fmt.Printf("  • %s\n", t.path)
	}
	fmt.Print("\nContinue? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("Uninstall cancelled")
		os.Exit(0)
	}

	errors := 0
	for _, t := range targets {
		var err error
		if t.isDir {
			err = os.RemoveAll(t.path)
		} else {
			err = os.Remove(t.path)
		}
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", t.path, err)
			errors++
		} else {
			fmt.Printf("✓ Removed %s\n", t.path)
		}
	}

	if errors > 0 {
		fmt.Printf("\nUninstall completed with %d error(s)\n", errors)
		os.Exit(1)
	}
	fmt.Println("\n✓ Uninstall complete")
}
