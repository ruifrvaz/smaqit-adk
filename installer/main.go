package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed agents/*.md
var adkAgentFiles embed.FS

//go:embed skills/smaqit.create-agent/SKILL.md
var adkCreateAgentSkillFile []byte

//go:embed skills/smaqit.create-skill/SKILL.md
var adkCreateSkillSkillFile []byte

//go:embed skills/smaqit.new-principle/SKILL.md
var adkNewPrincipleSkillFile []byte

//go:embed framework
var adkFrameworkFS embed.FS

//go:embed templates
var adkTemplatesFS embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "0.5.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "lite":
		targetDir := "."
		if len(os.Args) > 2 {
			targetDir = os.Args[2]
		}
		cmdLite(targetDir)
	case "advanced":
		targetDir := "."
		if len(os.Args) > 2 {
			targetDir = os.Args[2]
		}
		cmdAdvanced(targetDir)
	case "init":
		fmt.Println("'smaqit-adk init' has been replaced by explicit tier subcommands.")
		fmt.Println("  smaqit-adk lite      Install lite-tier agents and skills")
		fmt.Println("  smaqit-adk advanced  Install full ADK")
		os.Exit(1)
	case "help", "--help", "-h":
		cmdHelp()
	case "uninstall":
		tier := ""
		if len(os.Args) > 2 {
			tier = os.Args[2]
		}
		cmdUninstall(tier)
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
  lite [dir]                Install lite-tier agents and skills
  advanced [dir]            Install full ADK (includes lite + L0, L1, framework)
  help                      Show detailed help
  uninstall [lite|advanced] Remove smaqit-adk from project
  version                   Show smaqit-adk version`)
}

func cmdHelp() {
	fmt.Println("smaqit-adk - Agent Development Kit")
	fmt.Printf("Version: %s\n\n", Version)

	fmt.Println("  smaqit-adk lite [dir]")
	fmt.Println("      Install lite tier:")
	fmt.Println("        .github/agents/smaqit.L2.agent.md         (compiler agent)")
	fmt.Println("        .github/skills/smaqit.create-agent/       (create agent skill)")
	fmt.Println("        .github/skills/smaqit.create-skill/       (create skill skill)")
	fmt.Println("        .smaqit/templates/                        (compilation templates)")
	fmt.Println()
	fmt.Println("  smaqit-adk advanced [dir]")
	fmt.Println("      Install advanced tier (includes lite, plus):")
	fmt.Println("        .github/agents/smaqit.L0.agent.md         (principle curator)")
	fmt.Println("        .github/agents/smaqit.L1.agent.md         (template compiler)")
	fmt.Println("        .github/skills/smaqit.new-principle/      (framework authoring skill)")
	fmt.Println("        .smaqit/framework/                        (framework principle files)")
	fmt.Println()
	fmt.Println("  smaqit-adk uninstall [lite|advanced]")
	fmt.Println("      lite:     removes L2 agent, create skills, .smaqit/ entirely")
	fmt.Println("      advanced: removes L0/L1 agents, new-principle skill, .smaqit/framework/")
	fmt.Println("      (no arg): removes everything installed")
	fmt.Println()
	fmt.Println("  smaqit-adk version   Show smaqit-adk version")
	fmt.Println()
	fmt.Println("Usage in VS Code:")
	fmt.Println("  Say 'create a new agent' or use /smaqit.create-agent")
	fmt.Println("  Say 'create a new skill' or use /smaqit.create-skill")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/ruifrvaz/smaqit-adk")
}

// installLiteComponents installs the lite-tier components. The caller must have
// already changed to the target directory.
func installLiteComponents() {
	agentDir := filepath.Join(".github", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", agentDir, err)
		os.Exit(1)
	}

	// Install L2 agent
	l2Content, err := adkAgentFiles.ReadFile("agents/smaqit.L2.agent.md")
	if err != nil {
		fmt.Printf("Error reading smaqit.L2.agent.md: %v\n", err)
		os.Exit(1)
	}
	l2Path := filepath.Join(agentDir, "smaqit.L2.agent.md")
	if err := os.WriteFile(l2Path, l2Content, 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", l2Path, err)
		os.Exit(1)
	}

	// Install lite-tier skills
	type skillEntry struct {
		dir     string
		content []byte
	}
	skillEntries := []skillEntry{
		{filepath.Join(".github", "skills", "smaqit.create-agent"), adkCreateAgentSkillFile},
		{filepath.Join(".github", "skills", "smaqit.create-skill"), adkCreateSkillSkillFile},
	}
	for _, se := range skillEntries {
		if err := os.MkdirAll(se.dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", se.dir, err)
			os.Exit(1)
		}
		dstPath := filepath.Join(se.dir, "SKILL.md")
		if err := os.WriteFile(dstPath, se.content, 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", dstPath, err)
			os.Exit(1)
		}
	}

	// Install templates to .smaqit/templates/
	templatesDst := filepath.Join(".smaqit", "templates")
	if err := copyEmbedDir(adkTemplatesFS, "templates", templatesDst); err != nil {
		fmt.Printf("Error installing templates: %v\n", err)
		os.Exit(1)
	}
}

func cmdLite(targetDir string) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}
	if err := os.Chdir(targetDir); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	if _, err := os.Stat(filepath.Join(".github", "agents", "smaqit.L2.agent.md")); err == nil {
		fmt.Println("Error: smaqit-adk lite already installed")
		fmt.Println("Run 'smaqit-adk uninstall lite' first to remove existing installation")
		os.Exit(1)
	}

	installLiteComponents()

	fmt.Printf("✓ smaqit-adk %s lite installed\n", Version)
	fmt.Println("Use /smaqit.create-agent or /smaqit.create-skill in Copilot chat.")
}

func cmdAdvanced(targetDir string) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}
	if err := os.Chdir(targetDir); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	if _, err := os.Stat(filepath.Join(".github", "agents", "smaqit.L0.agent.md")); err == nil {
		fmt.Println("Error: smaqit-adk advanced already installed")
		fmt.Println("Run 'smaqit-adk uninstall advanced' first to remove existing installation")
		os.Exit(1)
	}

	// Install lite tier if not already present
	if _, err := os.Stat(filepath.Join(".github", "agents", "smaqit.L2.agent.md")); err != nil {
		installLiteComponents()
	}

	// Install L0/L1 agents to .github/agents/
	agentDir := filepath.Join(".github", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", agentDir, err)
		os.Exit(1)
	}
	for _, name := range []string{"smaqit.L0.agent.md", "smaqit.L1.agent.md"} {
		content, err := adkAgentFiles.ReadFile("agents/" + name)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", name, err)
			os.Exit(1)
		}
		dstPath := filepath.Join(agentDir, name)
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", dstPath, err)
			os.Exit(1)
		}
	}

	// Install new-principle skill to .github/skills/
	newPrincipleDir := filepath.Join(".github", "skills", "smaqit.new-principle")
	if err := os.MkdirAll(newPrincipleDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", newPrincipleDir, err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(newPrincipleDir, "SKILL.md"), adkNewPrincipleSkillFile, 0644); err != nil {
		fmt.Printf("Error writing new-principle skill: %v\n", err)
		os.Exit(1)
	}

	// Install framework to .smaqit/framework/
	frameworkDst := filepath.Join(".smaqit", "framework")
	if err := copyEmbedDir(adkFrameworkFS, "framework", frameworkDst); err != nil {
		fmt.Printf("Error installing framework: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ smaqit-adk %s advanced installed\n", Version)
	fmt.Println("Use /smaqit.create-agent, /smaqit.create-skill, and /smaqit.new-principle in Copilot chat.")
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

func cmdUninstall(tier string) {
	if tier != "" && tier != "lite" && tier != "advanced" {
		fmt.Printf("Unknown tier %q — use 'lite', 'advanced', or omit to remove all installed tiers.\n", tier)
		os.Exit(1)
	}

	// Detect what is installed
	liteInstalled := false
	if _, err := os.Stat(filepath.Join(".github", "agents", "smaqit.L2.agent.md")); err == nil {
		liteInstalled = true
	}
	advancedInstalled := false
	if _, err := os.Stat(filepath.Join(".github", "agents", "smaqit.L0.agent.md")); err == nil {
		advancedInstalled = true
	}

	removeLite := (tier == "" || tier == "lite") && liteInstalled
	removeAdvanced := (tier == "" || tier == "advanced") && advancedInstalled

	if !removeLite && !removeAdvanced {
		fmt.Println("No smaqit-adk installation found in this directory")
		os.Exit(0)
	}

	// List what will be removed
	fmt.Println("This will remove:")
	if removeAdvanced {
		fmt.Println("  \u2022 .github/agents/smaqit.L0.agent.md")
		fmt.Println("  \u2022 .github/agents/smaqit.L1.agent.md")
		fmt.Println("  \u2022 .github/skills/smaqit.new-principle/")
		fmt.Println("  \u2022 .smaqit/framework/")
	}
	if removeLite {
		fmt.Println("  \u2022 .github/agents/smaqit.L2.agent.md")
		fmt.Println("  \u2022 .github/skills/smaqit.create-agent/")
		fmt.Println("  \u2022 .github/skills/smaqit.create-skill/")
		fmt.Println("  \u2022 .smaqit/ (templates and any local definitions)")
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

	if removeAdvanced {
		// Remove L0/L1 agents
		for _, name := range []string{"smaqit.L0.agent.md", "smaqit.L1.agent.md"} {
			path := filepath.Join(".github", "agents", name)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Error removing %s: %v\n", path, err)
				errors++
			} else {
				fmt.Printf("\u2713 Removed %s\n", path)
			}
		}
		// Remove new-principle skill dir
		newPrincipleDir := filepath.Join(".github", "skills", "smaqit.new-principle")
		if err := os.RemoveAll(newPrincipleDir); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", newPrincipleDir, err)
			errors++
		} else {
			fmt.Printf("\u2713 Removed %s\n", newPrincipleDir)
		}
		// Remove framework
		frameworkDir := filepath.Join(".smaqit", "framework")
		if err := os.RemoveAll(frameworkDir); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", frameworkDir, err)
			errors++
		} else {
			fmt.Printf("\u2713 Removed %s\n", frameworkDir)
		}
	}

	if removeLite {
		// Remove L2 agent
		l2Path := filepath.Join(".github", "agents", "smaqit.L2.agent.md")
		if err := os.Remove(l2Path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", l2Path, err)
			errors++
		} else {
			fmt.Printf("\u2713 Removed %s\n", l2Path)
		}
		// Remove create-agent/create-skill skill dirs
		for _, dir := range []string{
			filepath.Join(".github", "skills", "smaqit.create-agent"),
			filepath.Join(".github", "skills", "smaqit.create-skill"),
		} {
			if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Error removing %s: %v\n", dir, err)
				errors++
			} else {
				fmt.Printf("\u2713 Removed %s\n", dir)
			}
		}
		// Remove entire .smaqit/ (templates + definitions + any scaffolding)
		if err := os.RemoveAll(".smaqit"); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing .smaqit: %v\n", err)
			errors++
		} else {
			fmt.Println("\u2713 Removed .smaqit/")
		}
	}

	// Cleanup empty parent dirs
	for _, dir := range []string{
		filepath.Join(".github", "agents"),
		filepath.Join(".github", "skills"),
		".github",
	} {
		if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
			if err := os.Remove(dir); err == nil {
				fmt.Printf("\u2713 Removed empty %s\n", dir)
			}
		}
	}

	if errors > 0 {
		fmt.Printf("\nUninstall completed with %d error(s)\n", errors)
		os.Exit(1)
	} else {
		fmt.Println("\n\u2713 Uninstall complete")
	}
}
