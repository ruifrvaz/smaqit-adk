package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed framework/*.md
var frameworkFiles embed.FS

//go:embed templates/agents/*.md templates/agents/compiled/*.md
var agentTemplateFiles embed.FS

//go:embed agents/*.md
var adkAgentFiles embed.FS

//go:embed prompts/*.md
var promptFiles embed.FS

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
fmt.Println(`smaqit-adk - Generic Agent Development Kit

Usage: smaqit-adk <command>

Commands:
  init [dir] Scaffold ADK structure (.smaqit/framework, .smaqit/templates, .github/agents)
             Optional: specify target directory (default: current)
  help       Show detailed command help
  uninstall  Remove smaqit-adk from project
  version    Show smaqit-adk version`)
}

func cmdHelp() {
fmt.Println("smaqit-adk - Generic Agent Development Kit")
fmt.Printf("Version: %s\n\n", Version)

fmt.Println("CLI Commands:")
fmt.Println("  smaqit-adk init [dir] Scaffold smaqit-adk project structure")
fmt.Println("                        Creates .smaqit/ and .github/ directories with")
fmt.Println("                        framework files, generic agent templates, and Level agents")
fmt.Println("                        Optional: specify target directory (created if needed)")
fmt.Println()
fmt.Println("  smaqit-adk help       Show this help message")
fmt.Println()
fmt.Println("  smaqit-adk uninstall  Remove smaqit-adk from project")
fmt.Println("                        Removes .smaqit/framework/, .smaqit/templates/,")
fmt.Println("                        .github/agents/, .github/prompts/")
fmt.Println()
fmt.Println("  smaqit-adk version    Show smaqit-adk version")
fmt.Println()
fmt.Println("Copilot Agents (use in GitHub Copilot chat with /):")
fmt.Println("  /smaqit.L0         Principle Curator (maintains framework purity)")
fmt.Println("  /smaqit.L1         Template Compiler (compiles principles to directives)")
fmt.Println("  /smaqit.L2         Agent Compiler (compiles templates to product agents)")
fmt.Println()
fmt.Println("Getting Started:")
fmt.Println("  1. Run 'smaqit-adk init' in your project directory")
fmt.Println("  2. Open GitHub Copilot chat in VS Code")
fmt.Println("  3. Fill .github/prompts/smaqit.new-agent.prompt.md with agent requirements")
fmt.Println("  4. Type '/smaqit.L2' to compile your custom agent")
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

// Check if .smaqit already exists
if _, err := os.Stat(".smaqit"); err == nil {
fmt.Println("Error: .smaqit/ directory already exists")
fmt.Println("Run 'smaqit-adk uninstall' first to remove existing installation")
os.Exit(1)
}

fmt.Printf("Initializing smaqit-adk in %s...\n", targetDir)

// Create directory structure
dirs := []string{
".smaqit/framework",
".smaqit/templates/agents/compiled",
".github/agents",
".github/prompts",
}

for _, dir := range dirs {
if err := os.MkdirAll(dir, 0755); err != nil {
fmt.Printf("Error creating directory %s: %v\n", dir, err)
os.Exit(1)
}
}

// Copy framework files
if err := copyEmbeddedDir(frameworkFiles, "framework", ".smaqit/framework"); err != nil {
fmt.Printf("Error copying framework files: %v\n", err)
os.Exit(1)
}

// Copy agent templates
if err := copyEmbeddedDir(agentTemplateFiles, "templates/agents", ".smaqit/templates/agents"); err != nil {
fmt.Printf("Error copying agent templates: %v\n", err)
os.Exit(1)
}

// Copy ADK agents (Level agents)
if err := copyEmbeddedDir(adkAgentFiles, "agents", ".github/agents"); err != nil {
fmt.Printf("Error copying ADK agents: %v\n", err)
os.Exit(1)
}

// Copy prompts
if err := copyEmbeddedDir(promptFiles, "prompts", ".github/prompts"); err != nil {
fmt.Printf("Error copying prompts: %v\n", err)
os.Exit(1)
}

fmt.Println("✓ Created .smaqit/ directory structure")
fmt.Println("✓ Copied framework files (5 generic principle files)")
fmt.Println("✓ Copied agent templates (3 generic templates + 3 compilation rules)")
fmt.Println("✓ Copied Level agents (L0, L1, L2)")
fmt.Println("✓ Copied prompts (new-agent template)")
fmt.Printf("✓ Initialized smaqit-adk %s\n\n", Version)
fmt.Println("Next steps:")
fmt.Println("  1. Open GitHub Copilot chat in VS Code")
fmt.Println("  2. Fill .github/prompts/smaqit.new-agent.prompt.md with agent requirements")
fmt.Println("  3. Type '/smaqit.L2' to compile your custom agent")
}

// copyEmbeddedDir copies files from an embedded FS to a target directory
func copyEmbeddedDir(embeddedFS embed.FS, srcDir, dstDir string) error {
return fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
if err != nil {
return err
}

if d.IsDir() {
return nil
}

// Read embedded file
content, err := embeddedFS.ReadFile(path)
if err != nil {
return fmt.Errorf("reading %s: %w", path, err)
}

// Calculate destination path
relPath := strings.TrimPrefix(path, srcDir+"/")
dstPath := filepath.Join(dstDir, relPath)

// Ensure destination directory exists
dstFileDir := filepath.Dir(dstPath)
if err := os.MkdirAll(dstFileDir, 0755); err != nil {
return fmt.Errorf("creating directory %s: %w", dstFileDir, err)
}

// Write file
if err := os.WriteFile(dstPath, content, 0644); err != nil {
return fmt.Errorf("writing %s: %w", dstPath, err)
}

return nil
})
}

func cmdUninstall() {
// Check if .smaqit exists
if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
fmt.Println("No smaqit-adk installation found in this directory")
os.Exit(0)
}

// Prompt for confirmation
fmt.Println("This will remove:")
fmt.Println("  • .smaqit/framework/")
fmt.Println("  • .smaqit/templates/agents/")
fmt.Println("  • .github/agents/")
fmt.Println("  • .github/prompts/")
fmt.Print("\nContinue? [y/N]: ")

var response string
fmt.Scanln(&response)
response = strings.ToLower(strings.TrimSpace(response))

if response != "y" && response != "yes" {
fmt.Println("Uninstall cancelled")
os.Exit(0)
}

// Remove directories
errors := 0

// Remove ADK-specific directories
adkDirs := []string{
".smaqit/framework",
".smaqit/templates/agents",
filepath.Join(".github", "agents"),
filepath.Join(".github", "prompts"),
}

for _, dir := range adkDirs {
if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
fmt.Printf("Error removing %s: %v\n", dir, err)
errors++
} else if err == nil {
fmt.Printf("✓ Removed %s\n", dir)
}
}

// Check if .smaqit/templates is empty and remove parent directories if so
templatesDir := ".smaqit/templates"
if entries, err := os.ReadDir(templatesDir); err == nil && len(entries) == 0 {
if err := os.Remove(templatesDir); err == nil {
fmt.Println("✓ Removed empty .smaqit/templates/")
}
}

// Check if .smaqit is empty and remove it
smaqitDir := ".smaqit"
if entries, err := os.ReadDir(smaqitDir); err == nil && len(entries) == 0 {
if err := os.Remove(smaqitDir); err == nil {
fmt.Println("✓ Removed empty .smaqit/")
}
}

// Check if .github is empty and remove it
githubDir := ".github"
if entries, err := os.ReadDir(githubDir); err == nil && len(entries) == 0 {
if err := os.Remove(githubDir); err == nil {
fmt.Println("✓ Removed empty .github/")
}
}

if errors > 0 {
fmt.Printf("\nUninstall completed with %d error(s)\n", errors)
os.Exit(1)
} else {
fmt.Println("\n✓ Uninstall complete")
}
}
