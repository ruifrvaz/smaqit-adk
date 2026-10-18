// runs Claude Code headlessly against a per-attempt isolated config directory.
package bench

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// claudeCredentials are the only variables a claude-code variant may name as
// its credential source. Exactly one is passed through from the invoking
// environment; every other Anthropic/Claude credential variable is absent.
var claudeCredentials = map[string]bool{"ANTHROPIC_API_KEY": true, "CLAUDE_CODE_OAUTH_TOKEN": true}

var claudePermissionModes = map[string]bool{
	"acceptEdits": true, "auto": true, "bypassPermissions": true, "default": true,
	"dontAsk": true, "manual": true, "plan": true,
}

// claudeReservedFlags are owned by the adapter. `--bare` is rejected because
// it disables the skill/agent/command discovery treatments rely on.
var claudeReservedFlags = []string{
	"-p", "--print", "--output-format", "--input-format", "--verbose",
	"--permission-mode", "--dangerously-skip-permissions", "--session-id",
	"--resume", "-r", "--continue", "-c", "--fork-session", "--no-session-persistence",
	"--bare", "--safe-mode", "--model", "--max-turns", "--allowedTools", "--allowed-tools",
	"--disallowedTools", "--disallowed-tools", "--settings", "--setting-sources",
}

// managedClaudeRoots lists platform-managed (enterprise) Claude Code
// locations whose skills and agents take precedence over personal ones.
var managedClaudeRoots = func() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/Library/Application Support/ClaudeCode"}
	case "windows":
		return []string{`C:\Program Files\ClaudeCode`}
	default:
		return []string{"/etc/claude-code"}
	}
}

func validateClaudeCode(c ClaudeCodeConfig, v Variant, path string, cases []Case, caseInputs []map[string]bool, treatmentIDs map[string]bool, add func(string, string)) {
	if c.Executable == "" {
		add(path+".executable", "is required")
	}
	if strings.HasPrefix(filepath.Base(c.Executable), "smaqit-adk") {
		add(path+".executable", "direct recursive smaqit-adk execution is not allowed")
	}
	if !claudePermissionModes[c.PermissionMode] {
		add(path+".permissionMode", "must be one of acceptEdits, auto, bypassPermissions, default, dontAsk, manual, or plan")
	}
	if c.MaxTurns < 0 {
		add(path+".maxTurns", "must not be negative")
	}
	if !claudeCredentials[c.Credential] {
		add(path+".credential", "must be ANTHROPIC_API_KEY or CLAUDE_CODE_OAUTH_TOKEN")
	}
	for i, tool := range append(append([]string{}, c.AllowedTools...), c.DisallowedTools...) {
		if strings.TrimSpace(tool) == "" || strings.HasPrefix(tool, "-") {
			add(fmt.Sprintf("%s.tools[%d]", path, i), "must be a non-empty tool name or rule")
		}
	}
	validateArguments(c.Arguments, path+".arguments", add)
	for caseIndex, cs := range cases {
		validateArgumentReferences(c.Arguments, path+".arguments", caseInputs[caseIndex], treatmentIDs, func(p, message string) {
			add(p, message+" for case "+cs.ID)
		})
	}
	for i, argument := range c.Arguments {
		if claudeReservedFlag(argument) {
			add(fmt.Sprintf("%s.arguments[%d]", path, i), "flag "+argument+" is controlled by the claude-code adapter")
		}
	}
	for i, name := range c.Environment.Inherit {
		ip := fmt.Sprintf("%s.environment.inherit[%d]", path, i)
		if !envNamePattern.MatchString(name) {
			add(ip, "must be an environment variable name")
		} else if name == "HOME" || strings.HasPrefix(name, "CLAUDE_") || strings.HasPrefix(name, "ANTHROPIC_") {
			add(ip, name+" must not be inherited; the adapter isolates HOME and CLAUDE_* state and passes only the declared credential")
		}
	}
	for name := range c.Environment.Set {
		sp := path + ".environment.set." + name
		if !envNamePattern.MatchString(name) {
			add(sp, "key must be an environment variable name")
		} else if claudeReservedSetName(name) {
			add(sp, name+" is controlled by the claude-code adapter or is credential-shaped")
		}
	}
	treatments := map[string]TreatmentAsset{}
	for _, t := range v.Treatment {
		treatments[t.ID] = t
	}
	installed := map[string]bool{}
	names := map[string]string{}
	for i, install := range c.Install {
		ip := fmt.Sprintf("%s.install[%d]", path, i)
		treatment, ok := treatments[install.Treatment]
		if !ok {
			add(ip+".treatment", "must name one of this variant's treatments")
			continue
		}
		if installed[install.Treatment] {
			add(ip+".treatment", "is installed more than once")
		}
		installed[install.Treatment] = true
		if install.As != "skill" && install.As != "agent" && install.As != "command" {
			add(ip+".as", "must be skill, agent, or command")
			continue
		}
		name, err := claudeInstallName(install.As, treatment.Source)
		if err != nil {
			add(ip, err.Error())
			continue
		}
		namespace := claudeNamespace(install.As)
		if previous, exists := names[namespace+"/"+name]; exists {
			add(ip, fmt.Sprintf("installs %s %q, which treatment %s already installs", install.As, name, previous))
		}
		names[namespace+"/"+name] = install.Treatment
	}
}

func claudeReservedFlag(argument string) bool {
	for _, flag := range claudeReservedFlags {
		if argument == flag || strings.HasPrefix(argument, flag+"=") {
			return true
		}
	}
	return false
}

func claudeReservedSetName(name string) bool {
	switch name {
	case "HOME", "CLAUDE_CONFIG_DIR", "CLAUDE_CODE_OAUTH_TOKEN", "CLAUDE_CODE_SESSION_ID":
		return true
	}
	return strings.HasPrefix(name, "ANTHROPIC_") || strings.HasPrefix(name, "CLAUDE_CODE_USE_") || strings.HasPrefix(name, "CLAUDE_CODE_MESSAGING_")
}

// claudeNamespace groups install kinds that share one name space: commands
// are merged into skills by Claude Code, while agents are separate.
func claudeNamespace(kind string) string {
	if kind == "agent" {
		return "agent"
	}
	return "skill"
}

// claudeInstallName validates a treatment's shape for its install kind and
// returns the name Claude Code will discover it under.
func claudeInstallName(kind, source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("treatment source: %w", err)
	}
	switch kind {
	case "skill":
		if !info.IsDir() {
			return "", fmt.Errorf("a skill treatment must be a directory containing SKILL.md")
		}
		name, err := frontmatterName(filepath.Join(source, "SKILL.md"))
		if err != nil {
			return "", fmt.Errorf("a skill treatment must be a directory containing SKILL.md: %w", err)
		}
		if name == "" {
			name = filepath.Base(source)
		}
		return name, nil
	case "agent":
		if info.IsDir() || filepath.Ext(source) != ".md" {
			return "", fmt.Errorf("an agent treatment must be a .md file")
		}
		name, err := frontmatterName(source)
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("an agent treatment must declare a frontmatter name")
		}
		return name, nil
	default:
		if info.IsDir() || filepath.Ext(source) != ".md" {
			return "", fmt.Errorf("a command treatment must be a .md file")
		}
		return strings.TrimSuffix(filepath.Base(source), ".md"), nil
	}
}

// frontmatterName returns the `name:` value of a Markdown file's YAML
// frontmatter block, or "" when the file has no frontmatter name.
func frontmatterName(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return "", scanner.Err()
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		if value, ok := strings.CutPrefix(line, "name:"); ok {
			return strings.Trim(strings.TrimSpace(value), `"'`), nil
		}
	}
	return "", scanner.Err()
}

type claudeInstalled struct{ Kind, Name, Treatment string }

type claudeCodeAdapter struct{}

func (claudeCodeAdapter) Execute(ctx context.Context, request RunRequest) (HarnessResult, error) {
	config := request.Variant.ClaudeCode
	result := HarnessResult{Adapter: "claude-code", Executable: config.Executable, Name: "claude-code", Version: request.HarnessVersion, SessionID: request.SessionID}
	if request.Workspace.HarnessHome == "" {
		return result, fmt.Errorf("claude-code adapter requires a harness home")
	}
	credential, ok := os.LookupEnv(config.Credential)
	if !ok || credential == "" {
		return result, fmt.Errorf("credential variable %s is not set in the invoking environment", config.Credential)
	}
	home := filepath.Join(request.Workspace.HarnessHome, "home")
	configDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return result, fmt.Errorf("create isolated config directory: %w", err)
	}
	installed, err := installClaudeTreatments(*config, request.Variant, configDir)
	if err != nil {
		return result, err
	}
	if err := checkClaudeShadowing(request.Workspace.Root, installed); err != nil {
		return result, err
	}
	extra, err := renderArguments(config.Arguments, request)
	if err != nil {
		return result, err
	}
	spec := processSpec{
		Executable: config.Executable, Arguments: claudeArguments(*config, request.SessionID, extra),
		Directory: request.Workspace.Root, Environment: claudeEnvironment(*config, home, configDir, credential),
		Stdin: request.CaseBrief, UseStdin: true, TraceDir: request.TraceDir, TracePrefix: "harness",
	}
	processResult, err := runProcess(ctx, spec)
	processResult.Adapter, processResult.Name, processResult.Version, processResult.SessionID = result.Adapter, result.Name, result.Version, result.SessionID
	result = processResult
	if err != nil {
		return result, err
	}
	if err := copyClaudeTranscript(configDir, request.SessionID, request.TraceDir); err != nil {
		return result, err
	}
	stream := parseClaudeStream(result.Stdout)
	result.Usage = stream.metrics()
	if stream.Result != nil && stream.Result.IsError {
		result.ReportedError = true
	}
	if stream.Init != nil {
		var missing []string
		for _, item := range installed {
			if !stream.Init.discovered(item.Kind, item.Name) {
				missing = append(missing, fmt.Sprintf("%s %q (treatment %s)", item.Kind, item.Name, item.Treatment))
			}
		}
		if len(missing) > 0 {
			return result, fmt.Errorf("installed treatment not discovered by Claude Code: %s", strings.Join(missing, ", "))
		}
	}
	return result, nil
}

// installClaudeTreatments copies each declared install from its verified
// treatment source into the isolated config directory.
func installClaudeTreatments(config ClaudeCodeConfig, variant Variant, configDir string) ([]claudeInstalled, error) {
	sources := map[string]string{}
	for _, t := range variant.Treatment {
		sources[t.ID] = t.Source
	}
	var installed []claudeInstalled
	for _, install := range config.Install {
		source := sources[install.Treatment]
		name, err := claudeInstallName(install.As, source)
		if err != nil {
			return nil, fmt.Errorf("install treatment %s: %w", install.Treatment, err)
		}
		var target string
		switch install.As {
		case "skill":
			target = filepath.Join(configDir, "skills", name)
			err = copyDirectory(source, target, nil)
		case "agent":
			target = filepath.Join(configDir, "agents", name+".md")
			err = copyFile(source, target, 0600)
		default:
			target = filepath.Join(configDir, "commands", name+".md")
			err = copyFile(source, target, 0600)
		}
		if err != nil {
			return nil, fmt.Errorf("install treatment %s as %s: %w", install.Treatment, install.As, err)
		}
		installed = append(installed, claudeInstalled{Kind: install.As, Name: name, Treatment: install.Treatment})
	}
	return installed, nil
}

// checkClaudeShadowing fails when a same-named skill, command, or agent is
// visible to Claude Code from a location other than the isolated config
// directory: the workspace's own .claude/, any ancestor's .claude/ (Claude Code
// walks upward from its working directory), or a platform-managed location.
// Any of those would shadow or compete with the treatment under test.
func checkClaudeShadowing(workspaceRoot string, installed []claudeInstalled) error {
	if len(installed) == 0 {
		return nil
	}
	var roots []string
	for dir := filepath.Clean(workspaceRoot); ; dir = filepath.Dir(dir) {
		roots = append(roots, filepath.Join(dir, ".claude"))
		if filepath.Dir(dir) == dir {
			break
		}
	}
	for _, managed := range managedClaudeRoots() {
		roots = append(roots, managed, filepath.Join(managed, ".claude"))
	}
	var conflicts []string
	for _, root := range roots {
		skills := claudeDefinedNames(root, "skill")
		agents := claudeDefinedNames(root, "agent")
		for _, item := range installed {
			names := skills
			if item.Kind == "agent" {
				names = agents
			}
			if names[item.Name] {
				conflicts = append(conflicts, fmt.Sprintf("%s %q is also defined under %s", item.Kind, item.Name, root))
			}
		}
	}
	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return fmt.Errorf("treatment would be shadowed: %s", strings.Join(conflicts, "; "))
	}
	return nil
}

// claudeDefinedNames lists the names defined under one .claude-style root for
// a name space: skills and commands share one; agents have their own.
func claudeDefinedNames(root, namespace string) map[string]bool {
	names := map[string]bool{}
	if namespace == "agent" {
		entries, _ := os.ReadDir(filepath.Join(root, "agents"))
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			names[strings.TrimSuffix(entry.Name(), ".md")] = true
			if name, err := frontmatterName(filepath.Join(root, "agents", entry.Name())); err == nil && name != "" {
				names[name] = true
			}
		}
		return names
	}
	entries, _ := os.ReadDir(filepath.Join(root, "skills"))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names[entry.Name()] = true
		if name, err := frontmatterName(filepath.Join(root, "skills", entry.Name(), "SKILL.md")); err == nil && name != "" {
			names[name] = true
		}
	}
	commands, _ := os.ReadDir(filepath.Join(root, "commands"))
	for _, entry := range commands {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			names[strings.TrimSuffix(entry.Name(), ".md")] = true
		}
	}
	return names
}

// claudeArguments builds the vetted headless invocation. The Case brief is
// delivered on stdin, so no positional prompt follows the flags.
func claudeArguments(config ClaudeCodeConfig, sessionID string, extra []string) []string {
	arguments := []string{"-p", "--output-format", "stream-json", "--verbose", "--permission-mode", config.PermissionMode, "--session-id", sessionID}
	if config.Model != "" {
		arguments = append(arguments, "--model", config.Model)
	}
	if config.MaxTurns > 0 {
		arguments = append(arguments, "--max-turns", strconv.Itoa(config.MaxTurns))
	}
	arguments = append(arguments, extra...)
	if len(config.AllowedTools) > 0 {
		arguments = append(append(arguments, "--allowedTools"), config.AllowedTools...)
	}
	if len(config.DisallowedTools) > 0 {
		arguments = append(append(arguments, "--disallowedTools"), config.DisallowedTools...)
	}
	return arguments
}

// claudeEnvironment builds the harness environment from nothing: the
// manifest's non-secret inherit/set entries, PATH, the isolated HOME and
// CLAUDE_CONFIG_DIR, and the single declared credential. The invoking
// process's own CLAUDE_* and ANTHROPIC_* variables never pass through.
func claudeEnvironment(config ClaudeCodeConfig, home, configDir, credential string) []string {
	values := map[string]string{"DISABLE_AUTOUPDATER": "1"}
	if path, ok := os.LookupEnv("PATH"); ok {
		values["PATH"] = path
	}
	for _, entry := range explicitEnvironment(config.Environment) {
		name, value, _ := strings.Cut(entry, "=")
		values[name] = value
	}
	values["HOME"] = home
	values["CLAUDE_CONFIG_DIR"] = configDir
	values[config.Credential] = credential
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	environment := make([]string, 0, len(names))
	for _, name := range names {
		environment = append(environment, name+"="+values[name])
	}
	return environment
}

// copyClaudeTranscript copies the session's own transcript out of the
// isolated config directory into the run's trace directory, when present.
func copyClaudeTranscript(configDir, sessionID, traceDir string) error {
	matches, err := filepath.Glob(filepath.Join(configDir, "projects", "*", sessionID+".jsonl"))
	if err != nil || len(matches) == 0 {
		return err
	}
	if err := copyFile(matches[0], filepath.Join(traceDir, "transcript.jsonl"), 0600); err != nil {
		return fmt.Errorf("copy session transcript: %w", err)
	}
	return nil
}

type claudeInit struct {
	Skills        []string `json:"skills"`
	Agents        []string `json:"agents"`
	SlashCommands []string `json:"slash_commands"`
}

func (i *claudeInit) discovered(kind, name string) bool {
	lists := [][]string{i.Skills, i.SlashCommands}
	if kind == "agent" {
		lists = [][]string{i.Agents}
	} else if kind == "command" {
		lists = [][]string{i.SlashCommands}
	}
	for _, list := range lists {
		for _, item := range list {
			if item == name {
				return true
			}
		}
	}
	return false
}

type claudeResult struct {
	IsError      bool     `json:"is_error"`
	TotalCostUSD *float64 `json:"total_cost_usd"`
	Usage        *struct {
		InputTokens              int64 `json:"input_tokens"`
		OutputTokens             int64 `json:"output_tokens"`
		CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

type claudeStream struct {
	Init      *claudeInit
	Result    *claudeResult
	ToolCalls int64
}

// parseClaudeStream reads a stream-json transcript line by line, tolerating
// lines that are not JSON (for example a truncated final line).
func parseClaudeStream(stdout string) claudeStream {
	var stream claudeStream
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var envelope struct {
			Type    string `json:"type"`
			Subtype string `json:"subtype"`
			Message struct {
				Content []struct {
					Type string `json:"type"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &envelope) != nil {
			continue
		}
		switch {
		case envelope.Type == "system" && envelope.Subtype == "init":
			var init claudeInit
			if json.Unmarshal([]byte(line), &init) == nil {
				stream.Init = &init
			}
		case envelope.Type == "result":
			var result claudeResult
			if json.Unmarshal([]byte(line), &result) == nil {
				stream.Result = &result
			}
		case envelope.Type == "assistant":
			for _, content := range envelope.Message.Content {
				if content.Type == "tool_use" {
					stream.ToolCalls++
				}
			}
		}
	}
	return stream
}

// metrics maps the final result event onto Bench's nullable usage metrics.
// Input tokens include cache-creation and cache-read input tokens.
func (s claudeStream) metrics() *NullableMetrics {
	if s.Result == nil {
		return nil
	}
	metrics := &NullableMetrics{EstimatedCost: s.Result.TotalCostUSD}
	toolCalls := s.ToolCalls
	metrics.ToolCalls = &toolCalls
	if s.Result.Usage != nil {
		input := s.Result.Usage.InputTokens + s.Result.Usage.CacheCreationInputTokens + s.Result.Usage.CacheReadInputTokens
		output := s.Result.Usage.OutputTokens
		total := input + output
		metrics.InputTokens, metrics.OutputTokens, metrics.TotalTokens = &input, &output, &total
	}
	return metrics
}

// newSessionID returns a random RFC 4122 version 4 UUID for --session-id.
func newSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
