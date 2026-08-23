package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// registryPath resolves which registry file a scope writes to. "base" is
// the ADK-global registry (this repo's own L0/L1/L2 + 5 shipped skills,
// or — once installed globally — smaqit-adk's own base framework);
// "extension" is a consuming project's own registry, tracking that
// project's create-agent/create-skill output against its own
// .smaqit/framework/*.md extension.
func registryPath(scope string) (string, error) {
	switch scope {
	case "base":
		return globalDataDir() + string(os.PathSeparator) + "registry.json", nil
	case "extension":
		return ".smaqit" + string(os.PathSeparator) + "registry.json", nil
	default:
		return "", fmt.Errorf("unknown registry scope %q (want \"base\" or \"extension\")", scope)
	}
}

// cmdRegistry dispatches the hidden `registry get`/`registry put`
// subcommands. Undocumented, like --install-global: these exist for
// smaqit.L0/L1/L2 and the new-principle/new-template/new-rules skills to
// shell out to via Bash, not for direct end-user use.
func cmdRegistry(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: smaqit-adk registry <get|put> [flags]")
		os.Exit(1)
	}
	switch args[0] {
	case "get":
		cmdRegistryGet(args[1:])
	case "put":
		cmdRegistryPut(args[1:])
	default:
		fmt.Printf("unknown registry subcommand %q\n", args[0])
		os.Exit(1)
	}
}

func cmdRegistryGet(args []string) {
	fs := flag.NewFlagSet("registry get", flag.ExitOnError)
	scope := fs.String("scope", "", "registry scope: base or extension")
	dependsOn := fs.String("depends-on", "", "filter to entries whose sourceTemplates includes this path")
	fs.Parse(args)

	path, err := registryPath(*scope)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	reg, err := ReadRegistry(path)
	if err != nil {
		fmt.Println("Error reading registry:", err)
		os.Exit(1)
	}
	entries := reg.Entries
	if *dependsOn != "" {
		entries = reg.DependentsOf(*dependsOn)
	}
	if entries == nil {
		entries = []RegistryEntry{}
	}
	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

func cmdRegistryPut(args []string) {
	fs := flag.NewFlagSet("registry put", flag.ExitOnError)
	scope := fs.String("scope", "", "registry scope: base or extension")
	name := fs.String("name", "", "registered resource name, e.g. smaqit.L2 or smaqit.create-agent")
	claudeTarget := fs.String("claude", "", "compiled Claude Code output path")
	codexTarget := fs.String("codex", "", "compiled Codex CLI output path")
	sourceDefinition := fs.String("source-definition", "", "path to the definition file this was compiled from")
	sourceTemplates := fs.String("source-templates", "", "comma-separated paths to L1 template/rules files this depends on")
	hashOf := fs.String("hash-of", "", "path to hash for drift detection (usually the primary compiled output)")
	fs.Parse(args)

	if *scope == "" || *name == "" || *hashOf == "" {
		fmt.Println("registry put requires --scope, --name, and --hash-of")
		os.Exit(1)
	}
	path, err := registryPath(*scope)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	reg, err := ReadRegistry(path)
	if err != nil {
		fmt.Println("Error reading registry:", err)
		os.Exit(1)
	}
	if reg.SchemaVersion == 0 {
		reg.SchemaVersion = RegistrySchemaVersion
	}
	if reg.Scope == "" {
		reg.Scope = *scope
	}
	hash, err := digestFileSHA256(*hashOf)
	if err != nil {
		fmt.Println("Error hashing", *hashOf, ":", err)
		os.Exit(1)
	}
	var templates []string
	if *sourceTemplates != "" {
		templates = strings.Split(*sourceTemplates, ",")
	}
	reg.PutEntry(RegistryEntry{
		Name:             *name,
		Target:           RegistryTarget{Claude: *claudeTarget, Codex: *codexTarget},
		SourceDefinition: *sourceDefinition,
		SourceTemplates:  templates,
		CompiledHash:     hash,
		LastCompiled:     time.Now().UTC().Format(time.RFC3339),
	})
	if err := WriteRegistry(path, reg); err != nil {
		fmt.Println("Error writing registry:", err)
		os.Exit(1)
	}
	fmt.Printf("✓ registered %s in %s registry (%s)\n", *name, *scope, path)
}

// cmdSync walks every registry this machine/project can see and reports
// drift — a registered target whose current content hash no longer matches
// what was recorded at compile time, meaning it was hand-edited outside the
// L0→L1→L2 chain since. It cannot recompile: recompilation is a semantic
// L1/L2 transformation only an agent can perform, not a deterministic file
// operation this plain Go binary can carry out headlessly in a raw shell or
// during install.sh. The smaqit.sync skill is the interactive counterpart
// that actually resolves what this command can only report — it shells out
// to this same binary's `registry get` from inside an agent session, where
// a Task-tool subagent invocation is available to do the real recompile.
func cmdSync() {
	drifted := false
	for _, scope := range []string{"base", "extension"} {
		path, err := registryPath(scope)
		if err != nil {
			continue
		}
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			continue
		}
		reg, err := ReadRegistry(path)
		if err != nil {
			fmt.Printf("Error reading %s registry (%s): %v\n", scope, path, err)
			continue
		}
		fmt.Printf("Checking %s registry (%s)...\n", scope, path)
		if len(reg.Entries) == 0 {
			fmt.Println("  (no registered entries)")
			continue
		}
		driftCount := 0
		for _, e := range reg.Entries {
			target := e.Target.Claude
			if target == "" {
				target = e.Target.Codex
			}
			if target == "" {
				continue
			}
			current, hashErr := digestFileSHA256(target)
			if hashErr != nil {
				fmt.Printf("  ✗ %s: %v\n", e.Name, hashErr)
				driftCount++
				drifted = true
				continue
			}
			if current != e.CompiledHash {
				fmt.Printf("  ⚠ %s: %s changed outside the compilation chain since %s\n", e.Name, target, e.LastCompiled)
				driftCount++
				drifted = true
			}
		}
		if driftCount == 0 {
			fmt.Printf("  ✓ %d entries, no drift\n", len(reg.Entries))
		} else {
			fmt.Printf("  %d/%d entries drifted — recompile via the appropriate create/new skill, then re-run sync\n", driftCount, len(reg.Entries))
		}
	}
	if drifted {
		os.Exit(1)
	}
}
