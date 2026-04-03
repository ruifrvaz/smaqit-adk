package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	copilot "github.com/github/copilot-sdk/go"
)

//go:embed agents/*.md
var adkAgentFiles embed.FS

//go:embed agents/smaqit.create-agent.agent.md
var adkCreateAgentFile []byte

//go:embed agents/smaqit.create-skill.agent.md
var adkCreateSkillFile []byte

//go:embed skills/smaqit.create-agent/SKILL.md
var adkCreateAgentSkillFile []byte

//go:embed skills/smaqit.create-skill/SKILL.md
var adkCreateSkillSkillFile []byte

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "0.3.2"

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
	case "create-agent":
		outputDir := ""
		if len(os.Args) > 2 && os.Args[2] == "--output" && len(os.Args) > 3 {
			outputDir = os.Args[3]
		}
		cmdCreate("agent", outputDir)
	case "create-skill":
		outputDir := ""
		if len(os.Args) > 2 && os.Args[2] == "--output" && len(os.Args) > 3 {
			outputDir = os.Args[3]
		}
		cmdCreate("skill", outputDir)
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
  init [dir]                     Install lite-tier agents and skills into .github/
  create-agent [--output <dir>]  Create a new agent interactively
  create-skill [--output <dir>]  Create a new skill interactively
  help                           Show detailed command help
  uninstall                      Remove smaqit-adk agents and skills from project
  version                        Show smaqit-adk version`)
}

func cmdHelp() {
	fmt.Println("smaqit-adk - Agent Development Kit")
	fmt.Printf("Version: %s\n\n", Version)

	fmt.Println("CLI Commands:")
	fmt.Println("  smaqit-adk init [dir]")
	fmt.Println("      Install lite-tier agents and skills into .github/agents/ and .github/skills/")
	fmt.Println("      Optional: specify target directory (created if needed)")
	fmt.Println()
	fmt.Println("  smaqit-adk create-agent [--output <dir>]")
	fmt.Println("      Interactively gather agent specs and compile a .agent.md into the project.")
	fmt.Println("      Runs in an isolated LLM context — no project agent instructions in scope.")
	fmt.Println("      Output defaults to ./.github/agents/")
	fmt.Println()
	fmt.Println("  smaqit-adk create-skill [--output <dir>]")
	fmt.Println("      Interactively gather skill specs and compile a SKILL.md into the project.")
	fmt.Println("      Runs in an isolated LLM context — no project agent instructions in scope.")
	fmt.Println("      Output defaults to ./.github/skills/<name>/")
	fmt.Println()
	fmt.Println("  smaqit-adk help       Show this help message")
	fmt.Println()
	fmt.Println("  smaqit-adk uninstall  Remove smaqit-adk agents and skills from project")
	fmt.Println("                        Removes agents and routing skills from .github/agents/")
	fmt.Println("                        and .github/skills/")
	fmt.Println()
	fmt.Println("  smaqit-adk version    Show smaqit-adk version")
	fmt.Println()
	fmt.Println("VS Code Lite Tier (install with 'init'):")
	fmt.Println("  Say 'create a new agent' or 'create a new skill' in Copilot chat.")
	fmt.Println("  Or use the slash commands: /smaqit.create-agent, /smaqit.create-skill")
	fmt.Println("  Copilot activates the skill, which invokes the agent as a subagent (isolated context).")
	fmt.Println()
	fmt.Println("Auth: set COPILOT_GITHUB_TOKEN, GH_TOKEN, or GITHUB_TOKEN,")
	fmt.Println("      or log in with 'gh auth login' / VS Code GitHub Copilot extension.")
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

	// Copy the two lite-tier routing skills
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

	fmt.Println("✓ Installed smaqit.create-agent into .github/agents/ and .github/skills/")
	fmt.Println("✓ Installed smaqit.create-skill into .github/agents/ and .github/skills/")
	fmt.Printf("✓ smaqit-adk %s ready\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open GitHub Copilot chat in VS Code")
	fmt.Println("  2. Say 'create a new agent' or 'create a new skill'")
	fmt.Println("     or use /smaqit.create-agent and /smaqit.create-skill")
	fmt.Println("  Copilot activates the skill, which runs the agent in an isolated context.")
}



func cmdUninstall() {
	agentDir := filepath.Join(".github", "agents")
	agentNames := []string{
		"smaqit.create-agent.agent.md",
		"smaqit.create-skill.agent.md",
	}

	skillsDir := filepath.Join(".github", "skills")
	skillDirs := []string{
		filepath.Join(skillsDir, "smaqit.create-agent"),
		filepath.Join(skillsDir, "smaqit.create-skill"),
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
	for _, dir := range skillDirs {
		fmt.Printf("  \u2022 %s/SKILL.md\n", dir)
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

	for _, dir := range skillDirs {
		skillPath := filepath.Join(dir, "SKILL.md")
		if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing %s: %v\n", skillPath, err)
			errors++
		} else {
			fmt.Printf("\u2713 Removed %s\n", skillPath)
		}
		// Remove the skill subdirectory if now empty
		if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
			if err := os.Remove(dir); err == nil {
				fmt.Printf("\u2713 Removed empty %s\n", dir)
			}
		}
	}

	// Remove .github/agents/ only if now empty
	if entries, err := os.ReadDir(agentDir); err == nil && len(entries) == 0 {
		if err := os.Remove(agentDir); err == nil {
			fmt.Println("\u2713 Removed empty .github/agents/")
		}
	}

	// Remove .github/skills/ only if now empty
	if entries, err := os.ReadDir(skillsDir); err == nil && len(entries) == 0 {
		if err := os.Remove(skillsDir); err == nil {
			fmt.Println("\u2713 Removed empty .github/skills/")
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

// cmdCreate drives an interactive create-agent or create-skill session via the Copilot SDK.
// kind must be "agent" or "skill". outputDir overrides the default output location.
func cmdCreate(kind, outputDir string) {
	var systemContent string
	var initialPrompt string
	var defaultOutputDir string

	switch kind {
	case "agent":
		systemContent = string(adkCreateAgentFile)
		initialPrompt = "The user wants to create a new agent. Begin by asking for the agent name."
		defaultOutputDir = filepath.Join(".github", "agents")
	case "skill":
		systemContent = string(adkCreateSkillFile)
		initialPrompt = "The user wants to create a new skill. Begin by asking for the skill name."
		defaultOutputDir = filepath.Join(".github", "skills")
	default:
		fmt.Fprintf(os.Stderr, "unknown kind: %s\n", kind)
		os.Exit(1)
	}

	if outputDir == "" {
		outputDir = defaultOutputDir
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// No timeout — interactive sessions are human-paced. Ctrl-C is the exit path.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl-C cancels cleanly via context.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nCancelled.")
		cancel()
	}()

	token := resolveToken()

	client := copilot.NewClient(&copilot.ClientOptions{
		Cwd:         cwd,
		GitHubToken: token,
	})

	sessionCfg := &copilot.SessionConfig{
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "replace",
			Content: systemContent,
		},
		WorkingDirectory: cwd,
		OnPermissionRequest: func(req copilot.PermissionRequest, _ copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
			// Deny shell execution; approve file reads and writes.
			if req.Kind == copilot.PermissionRequestKindShell {
				return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindDeniedByRules}, nil
			}
			return copilot.PermissionRequestResult{Kind: copilot.PermissionRequestResultKindApproved}, nil
		},
		OnUserInputRequest: func(req copilot.UserInputRequest, _ copilot.UserInputInvocation) (copilot.UserInputResponse, error) {
			if req.Question != "" {
				fmt.Printf("\n%s\n\n", req.Question)
			}
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			if err != nil {
				return copilot.UserInputResponse{}, fmt.Errorf("reading input: %w", err)
			}
			return copilot.UserInputResponse{Answer: strings.TrimSpace(answer)}, nil
		},
	}

	session, err := client.CreateSession(ctx, sessionCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("smaqit-adk create-%s — working directory: %s\n", kind, cwd)
	fmt.Printf("Output: %s\n", filepath.Join(cwd, outputDir))
	fmt.Println("Type your answers when prompted. Ctrl-C to cancel.")
	fmt.Println()

	_, sendErr := session.SendAndWait(ctx, copilot.MessageOptions{Prompt: initialPrompt})

	if sendErr != nil {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\nCancelled.")
		} else {
			fmt.Fprintf(os.Stderr, "session error: %v\n", sendErr)
		}
		os.Exit(1)
	}

	// Print any final assistant messages not delivered via ask_user.
	if events, err := session.GetMessages(ctx); err == nil {
		for _, ev := range events {
			if ev.Type == copilot.SessionEventTypeAssistantMessage && ev.Data.Content != nil {
				fmt.Printf("\n%s\n", *ev.Data.Content)
			}
		}
	}

	fmt.Printf("\n✓ Done. Output written to %s\n", filepath.Join(cwd, outputDir))
}

// resolveToken returns a GitHub token from environment variables, or empty string
// to use the SDK's default UseLoggedInUser behaviour (VS Code / gh CLI credentials).
func resolveToken() string {
	for _, env := range []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN", "GITHUB_TOKEN"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}
