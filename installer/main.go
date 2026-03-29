package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed agents/*.md
var adkAgentFiles embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "0.1.0"

func main() {
if len(os.Args) < 2 {
printUsage()
os.Exit(1)
}

switch os.Args[1] {
case "init":
targetDir := "."
if len(os.Args) > 2 {
targetDir = os.Args[2]
}
cmdInit(targetDir)
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
  init [dir] Install smaqit.create-agent and smaqit.create-skill into .github/agents/
             Optional: specify target directory (default: current)
  help       Show detailed command help
  uninstall  Remove smaqit-adk agents from project
  version    Show smaqit-adk version`)
}

func cmdHelp() {
	fmt.Println("smaqit-adk - Agent Development Kit")
	fmt.Printf("Version: %s\n\n", Version)

	fmt.Println("CLI Commands:")
	fmt.Println("  smaqit-adk init [dir] Install smaqit.create-agent and smaqit.create-skill")
	fmt.Println("                        into .github/agents/ in the target project")
	fmt.Println("                        Optional: specify target directory (created if needed)")
	fmt.Println()
	fmt.Println("  smaqit-adk help       Show this help message")
	fmt.Println()
	fmt.Println("  smaqit-adk uninstall  Remove smaqit-adk agents from project")
	fmt.Println("                        Removes smaqit.create-agent.agent.md and")
	fmt.Println("                        smaqit.create-skill.agent.md from .github/agents/")
	fmt.Println()
	fmt.Println("  smaqit-adk version    Show smaqit-adk version")
	fmt.Println()
	fmt.Println("Copilot Agents (invoke as subagents for clean, isolated context):")
	fmt.Println("  @smaqit.create-agent  Interactively gather specs and compile a new .agent.md")
	fmt.Println("  @smaqit.create-skill  Interactively gather specs and compile a new SKILL.md")
	fmt.Println()
	fmt.Println("Getting Started:")
	fmt.Println("  1. Run 'smaqit-adk init' in your project directory")
	fmt.Println("  2. Open GitHub Copilot chat in VS Code")
	fmt.Println("  3. Start a new chat and select @smaqit.create-agent or @smaqit.create-skill")
	fmt.Println("     Tip: run as a subagent for a clean context isolated from your current session")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/ruifrvaz/smaqit-adk")
}

func cmdInit(targetDir string) {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Change to target directory
	if err := os.Chdir(targetDir); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Check if already installed
	agentDir := filepath.Join(".github", "agents")
	if _, err := os.Stat(filepath.Join(agentDir, "smaqit.create-agent.agent.md")); err == nil {
		fmt.Println("Error: smaqit-adk agents already installed in .github/agents/")
		fmt.Println("Run 'smaqit-adk uninstall' first to remove existing installation")
		os.Exit(1)
	}

	fmt.Printf("Initializing smaqit-adk in %s...\n", targetDir)

	// Create .github/agents/ directory
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", agentDir, err)
		os.Exit(1)
	}

	// Copy the two lite-tier agents
	agentNames := []string{
		"smaqit.create-agent.agent.md",
		"smaqit.create-skill.agent.md",
	}
	for _, name := range agentNames {
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

	fmt.Println("✓ Installed smaqit.create-agent into .github/agents/")
	fmt.Println("✓ Installed smaqit.create-skill into .github/agents/")
	fmt.Printf("✓ smaqit-adk %s ready\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open GitHub Copilot chat in VS Code")
	fmt.Println("  2. Start a new chat and select @smaqit.create-agent to create a new agent,")
	fmt.Println("     or @smaqit.create-skill to create a new skill")
	fmt.Println("  Tip: run these as subagents for a clean, isolated context")
}



func cmdUninstall() {
	agentDir := filepath.Join(".github", "agents")
	agentNames := []string{
		"smaqit.create-agent.agent.md",
		"smaqit.create-skill.agent.md",
	}

	// Check if at least one agent is installed
	found := false
	for _, name := range agentNames {
		if _, err := os.Stat(filepath.Join(agentDir, name)); err == nil {
			found = true
			break
		}
	}

	if !found {
		fmt.Println("No smaqit-adk installation found in this directory")
		os.Exit(0)
	}

	// Prompt for confirmation
	fmt.Println("This will remove:")
	for _, name := range agentNames {
		fmt.Printf("  \u2022 .github/agents/%s\n", name)
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
	for _, name := range agentNames {
		path := filepath.Join(agentDir, name)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", path, err)
			errors++
		} else {
			fmt.Printf("\u2713 Removed %s\n", path)
		}
	}

	// Remove .github/agents/ only if now empty
	if entries, err := os.ReadDir(agentDir); err == nil && len(entries) == 0 {
		if err := os.Remove(agentDir); err == nil {
			fmt.Println("\u2713 Removed empty .github/agents/")
		}
	}

	// Remove .github/ only if now empty
	githubDir := ".github"
	if entries, err := os.ReadDir(githubDir); err == nil && len(entries) == 0 {
		if err := os.Remove(githubDir); err == nil {
			fmt.Println("\u2713 Removed empty .github/")
		}
	}

	if errors > 0 {
		fmt.Printf("\nUninstall completed with %d error(s)\n", errors)
		os.Exit(1)
	} else {
		fmt.Println("\n\u2713 Uninstall complete")
	}
}
