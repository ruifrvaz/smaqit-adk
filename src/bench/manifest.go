// defines, loads, resolves, and strictly validates benchmark manifests.
package bench

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const ManifestSchemaVersion = 1

var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

type Manifest struct {
	SchemaVersion int              `yaml:"schemaVersion" json:"schemaVersion"`
	Name          string           `yaml:"name" json:"name"`
	Cases         []Case           `yaml:"cases" json:"cases"`
	Variants      []Variant        `yaml:"variants" json:"variants"`
	Execution     Execution        `yaml:"execution" json:"execution"`
	Graders       []OptionalGrader `yaml:"graders,omitempty" json:"graders,omitempty"`
	Comparison    Comparison       `yaml:"comparison,omitempty" json:"comparison,omitempty"`
	Output        Output           `yaml:"output" json:"output"`
}

type Case struct {
	ID      string        `yaml:"id" json:"id"`
	Fixture *SourceRef    `yaml:"fixture,omitempty" json:"fixture,omitempty"`
	Given   Given         `yaml:"given" json:"given"`
	Expect  []Expectation `yaml:"expect" json:"expect"`
}

type Given struct {
	Prompt      Prompt       `yaml:"prompt" json:"prompt"`
	Specs       []InputAsset `yaml:"specs,omitempty" json:"specs,omitempty"`
	Files       []InputAsset `yaml:"files,omitempty" json:"files,omitempty"`
	Directories []InputAsset `yaml:"directories,omitempty" json:"directories,omitempty"`
	Images      []InputAsset `yaml:"images,omitempty" json:"images,omitempty"`
}

type Prompt struct {
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
	File string `yaml:"file,omitempty" json:"file,omitempty"`
}

type SourceRef struct {
	Source string `yaml:"source" json:"source"`
}

type InputAsset struct {
	ID          string `yaml:"id" json:"id"`
	Source      string `yaml:"source" json:"source"`
	Destination string `yaml:"destination,omitempty" json:"destination,omitempty"`
	MediaType   string `yaml:"mediaType,omitempty" json:"mediaType,omitempty"`
}

type Variant struct {
	ID                  string         `yaml:"id" json:"id"`
	Adapter             string         `yaml:"adapter" json:"adapter"`
	Process             *ProcessConfig `yaml:"process,omitempty" json:"process,omitempty"`
	Mock                *MockConfig    `yaml:"mock,omitempty" json:"mock,omitempty"`
	Setup               []Command      `yaml:"setup,omitempty" json:"setup,omitempty"`
	IntendedDifferences []string       `yaml:"intendedDifferences,omitempty" json:"intendedDifferences,omitempty"`
}

type ProcessConfig struct {
	Executable       string      `yaml:"executable" json:"executable"`
	Arguments        []string    `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	InputMode        string      `yaml:"inputMode,omitempty" json:"inputMode,omitempty"`
	WorkingDirectory string      `yaml:"workingDirectory,omitempty" json:"workingDirectory,omitempty"`
	Environment      Environment `yaml:"environment,omitempty" json:"environment,omitempty"`
}

type Environment struct {
	Inherit []string          `yaml:"inherit,omitempty" json:"inherit,omitempty"`
	Set     map[string]string `yaml:"set,omitempty" json:"set,omitempty"`
}

type MockConfig struct {
	Stdout   string            `yaml:"stdout,omitempty" json:"stdout,omitempty"`
	Stderr   string            `yaml:"stderr,omitempty" json:"stderr,omitempty"`
	ExitCode int               `yaml:"exitCode,omitempty" json:"exitCode,omitempty"`
	Files    map[string]string `yaml:"files,omitempty" json:"files,omitempty"`
}

type Command struct {
	Executable     string      `yaml:"executable" json:"executable"`
	Arguments      []string    `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	TimeoutSeconds int         `yaml:"timeoutSeconds,omitempty" json:"timeoutSeconds,omitempty"`
	Environment    Environment `yaml:"environment,omitempty" json:"environment,omitempty"`
}

type Expectation struct {
	ID             string   `yaml:"id" json:"id"`
	Type           string   `yaml:"type" json:"type"`
	Actual         string   `yaml:"actual" json:"actual"`
	Operator       string   `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value          string   `yaml:"value,omitempty" json:"value,omitempty"`
	ValueFile      string   `yaml:"valueFile,omitempty" json:"valueFile,omitempty"`
	Path           string   `yaml:"path,omitempty" json:"path,omitempty"`
	Golden         string   `yaml:"golden,omitempty" json:"golden,omitempty"`
	SHA256         string   `yaml:"sha256,omitempty" json:"sha256,omitempty"`
	IgnoreCase     bool     `yaml:"ignoreCase,omitempty" json:"ignoreCase,omitempty"`
	TrimFinalLine  bool     `yaml:"trimFinalNewline,omitempty" json:"trimFinalNewline,omitempty"`
	Command        *Command `yaml:"command,omitempty" json:"command,omitempty"`
	ExitCode       *int     `yaml:"exitCode,omitempty" json:"exitCode,omitempty"`
	RequiredPaths  []string `yaml:"requiredPaths,omitempty" json:"requiredPaths,omitempty"`
	ForbiddenPaths []string `yaml:"forbiddenPaths,omitempty" json:"forbiddenPaths,omitempty"`
}

type OptionalGrader struct {
	ID      string   `yaml:"id" json:"id"`
	Type    string   `yaml:"type" json:"type"`
	Weight  float64  `yaml:"weight" json:"weight"`
	Command Command  `yaml:"command,omitempty" json:"command,omitempty"`
	Assets  []string `yaml:"assets,omitempty" json:"assets,omitempty"`
}

type Execution struct {
	Repetitions    int    `yaml:"repetitions" json:"repetitions"`
	RandomizeOrder bool   `yaml:"randomizeOrder,omitempty" json:"randomizeOrder,omitempty"`
	Seed           *int64 `yaml:"seed,omitempty" json:"seed,omitempty"`
	TimeoutSeconds int    `yaml:"timeoutSeconds" json:"timeoutSeconds"`
	FailFast       bool   `yaml:"failFast,omitempty" json:"failFast,omitempty"`
}

type Comparison struct {
	MinimumRequiredPassRate float64  `yaml:"minimumRequiredPassRate,omitempty" json:"minimumRequiredPassRate,omitempty"`
	TieThreshold            float64  `yaml:"tieThreshold,omitempty" json:"tieThreshold,omitempty"`
	TieBreakers             []string `yaml:"tieBreakers,omitempty" json:"tieBreakers,omitempty"`
}

type Output struct {
	Directory string `yaml:"directory" json:"directory"`
}

type Diagnostic struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidationError struct{ Diagnostics []Diagnostic }

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Diagnostics))
	for i, d := range e.Diagnostics {
		parts[i] = d.Path + ": " + d.Message
	}
	return strings.Join(parts, "; ")
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		if diagnostics := unknownFieldDiagnostics(data, reflect.TypeOf(Manifest{})); len(diagnostics) > 0 {
			return nil, &ValidationError{Diagnostics: diagnostics}
		}
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode manifest: %w", err)
		}
		return nil, errors.New("decode manifest: multiple YAML documents are not allowed")
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("resolve manifest directory: %w", err)
	}
	m.applyDefaults()
	m.resolvePaths(base)
	if ds := m.Validate(); len(ds) > 0 {
		return nil, &ValidationError{Diagnostics: ds}
	}
	return &m, nil
}

func unknownFieldDiagnostics(data []byte, target reflect.Type) []Diagnostic {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil || len(document.Content) == 0 {
		return nil
	}
	var result []Diagnostic
	var visit func(*yaml.Node, reflect.Type, string)
	visit = func(node *yaml.Node, t reflect.Type, path string) {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			visit(node.Content[0], t, path)
			return
		}
		switch t.Kind() {
		case reflect.Struct:
			if node.Kind != yaml.MappingNode {
				return
			}
			fields := map[string]reflect.Type{}
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				name := strings.Split(field.Tag.Get("yaml"), ",")[0]
				if name != "" && name != "-" {
					fields[name] = field.Type
				}
			}
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				fieldType, ok := fields[key.Value]
				fieldPath := key.Value
				if path != "" {
					fieldPath = path + "." + key.Value
				}
				if !ok {
					result = append(result, Diagnostic{Path: fieldPath, Message: "unknown field"})
					continue
				}
				visit(value, fieldType, fieldPath)
			}
		case reflect.Slice:
			if node.Kind != yaml.SequenceNode {
				return
			}
			for i, child := range node.Content {
				visit(child, t.Elem(), fmt.Sprintf("%s[%d]", path, i))
			}
		case reflect.Map:
			return
		}
	}
	visit(&document, target, "")
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func (m *Manifest) applyDefaults() {
	if m.Execution.Repetitions == 0 {
		m.Execution.Repetitions = 1
	}
	if m.Execution.TimeoutSeconds == 0 {
		m.Execution.TimeoutSeconds = 300
	}
	if m.Comparison.MinimumRequiredPassRate == 0 {
		m.Comparison.MinimumRequiredPassRate = 1
	}
	if m.Comparison.TieThreshold == 0 {
		m.Comparison.TieThreshold = .01
	}
	for i := range m.Variants {
		if m.Variants[i].Process != nil && m.Variants[i].Process.InputMode == "" {
			m.Variants[i].Process.InputMode = "stdin"
		}
	}
}

func (m *Manifest) resolvePaths(base string) {
	resolve := func(p string) string {
		if p == "" {
			return ""
		}
		if filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
		return filepath.Clean(filepath.Join(base, p))
	}
	resolveExecutable := func(p string) string {
		if strings.ContainsAny(p, "/\\") || strings.HasPrefix(p, ".") {
			return resolve(p)
		}
		return p
	}
	m.Output.Directory = resolve(m.Output.Directory)
	for i := range m.Cases {
		c := &m.Cases[i]
		if c.Fixture != nil {
			c.Fixture.Source = resolve(c.Fixture.Source)
		}
		c.Given.Prompt.File = resolve(c.Given.Prompt.File)
		groups := []*[]InputAsset{&c.Given.Specs, &c.Given.Files, &c.Given.Directories, &c.Given.Images}
		for _, group := range groups {
			for j := range *group {
				(*group)[j].Source = resolve((*group)[j].Source)
			}
		}
		for j := range c.Expect {
			c.Expect[j].ValueFile = resolve(c.Expect[j].ValueFile)
			c.Expect[j].Golden = resolve(c.Expect[j].Golden)
			if c.Expect[j].Command != nil {
				c.Expect[j].Command.Executable = resolveExecutable(c.Expect[j].Command.Executable)
			}
		}
	}
	for i := range m.Variants {
		if m.Variants[i].Process != nil {
			m.Variants[i].Process.Executable = resolveExecutable(m.Variants[i].Process.Executable)
		}
		for j := range m.Variants[i].Setup {
			m.Variants[i].Setup[j].Executable = resolveExecutable(m.Variants[i].Setup[j].Executable)
		}
	}
	for i := range m.Graders {
		m.Graders[i].Command.Executable = resolveExecutable(m.Graders[i].Command.Executable)
		for j := range m.Graders[i].Assets {
			m.Graders[i].Assets[j] = resolve(m.Graders[i].Assets[j])
		}
	}
}

func (m *Manifest) Validate() []Diagnostic {
	var ds []Diagnostic
	add := func(path, msg string) { ds = append(ds, Diagnostic{Path: path, Message: msg}) }
	if m.SchemaVersion != ManifestSchemaVersion {
		add("schemaVersion", "must be 1")
	}
	if strings.TrimSpace(m.Name) == "" {
		add("name", "is required")
	}
	if len(m.Cases) == 0 {
		add("cases", "must contain at least one case")
	}
	if len(m.Variants) == 0 {
		add("variants", "must contain at least one variant")
	}
	seenCases := map[string]bool{}
	for i, c := range m.Cases {
		p := fmt.Sprintf("cases[%d]", i)
		validateID(c.ID, p+".id", seenCases, add)
		if (c.Given.Prompt.Text == "") == (c.Given.Prompt.File == "") {
			add(p+".given.prompt", "must set exactly one of text or file")
		}
		if c.Given.Prompt.File != "" {
			requireFile(c.Given.Prompt.File, p+".given.prompt.file", add)
		}
		if c.Fixture != nil {
			requireDir(c.Fixture.Source, p+".fixture.source", add)
			if pathWithin(c.Fixture.Source, m.Output.Directory) {
				add("output.directory", "must be outside fixture "+c.Fixture.Source)
			}
		}
		assetIDs := map[string]bool{}
		destinations := map[string]bool{}
		for _, named := range []struct {
			name      string
			assets    []InputAsset
			directory bool
		}{
			{"specs", c.Given.Specs, false}, {"files", c.Given.Files, false},
			{"directories", c.Given.Directories, true}, {"images", c.Given.Images, false},
		} {
			for j, a := range named.assets {
				ap := fmt.Sprintf("%s.given.%s[%d]", p, named.name, j)
				validateID(a.ID, ap+".id", assetIDs, add)
				requirePath(a.Source, ap+".source", named.directory, add)
				if named.directory && pathWithin(a.Source, m.Output.Directory) {
					add("output.directory", "must be outside declared input directory "+a.Source)
				}
				if a.Destination != "" && !safeRelative(a.Destination) {
					add(ap+".destination", "must be a relative contained path")
				}
				destination := filepath.Clean(a.Destination)
				if a.Destination != "" && destinations[destination] {
					add(ap+".destination", "must be unique within the case")
				}
				if a.Destination != "" {
					destinations[destination] = true
				}
			}
		}
		if len(c.Expect) == 0 {
			add(p+".expect", "must contain at least one required expectation")
		}
		expectIDs := map[string]bool{}
		for j, e := range c.Expect {
			ep := fmt.Sprintf("%s.expect[%d]", p, j)
			validateID(e.ID, ep+".id", expectIDs, add)
			validateExpectation(e, ep, add)
		}
	}
	seenVariants := map[string]bool{}
	for i, v := range m.Variants {
		p := fmt.Sprintf("variants[%d]", i)
		validateID(v.ID, p+".id", seenVariants, add)
		switch v.Adapter {
		case "process":
			if v.Mock != nil {
				add(p+".mock", "must not be set for process adapter")
			}
			if v.Process == nil {
				add(p+".process", "is required for process adapter")
			} else {
				if v.Process.Executable == "" {
					add(p+".process.executable", "is required")
				}
				if v.Process.InputMode != "stdin" && v.Process.InputMode != "argument" {
					add(p+".process.inputMode", "must be stdin or argument")
				}
				validateArguments(v.Process.Arguments, p+".process.arguments", add)
				if strings.HasPrefix(filepath.Base(v.Process.Executable), "smaqit-adk") && len(v.Process.Arguments) > 0 && v.Process.Arguments[0] == "bench" {
					add(p+".process", "direct recursive smaqit-adk bench execution is not allowed")
				}
				if v.Process.WorkingDirectory != "" && !safeRelative(v.Process.WorkingDirectory) {
					add(p+".process.workingDirectory", "must be a relative contained path")
				}
				for j, name := range v.Process.Environment.Inherit {
					if !envNamePattern.MatchString(name) {
						add(fmt.Sprintf("%s.process.environment.inherit[%d]", p, j), "must be an environment variable name")
					}
				}
				for name := range v.Process.Environment.Set {
					if !envNamePattern.MatchString(name) {
						add(p+".process.environment.set."+name, "key must be an environment variable name")
					}
				}
			}
		case "mock":
			if v.Process != nil {
				add(p+".process", "must not be set for mock adapter")
			}
			if v.Mock == nil {
				add(p+".mock", "is required for mock adapter")
			}
			if v.Mock != nil {
				for destination := range v.Mock.Files {
					if !safeRelative(destination) {
						add(p+".mock.files."+destination, "must be a relative contained path")
					}
				}
			}
		default:
			add(p+".adapter", "must be process or mock")
		}
		for j, command := range v.Setup {
			validateCommand(command, fmt.Sprintf("%s.setup[%d]", p, j), add)
		}
	}
	if m.Execution.Repetitions < 1 {
		add("execution.repetitions", "must be at least 1")
	}
	if m.Execution.TimeoutSeconds < 1 {
		add("execution.timeoutSeconds", "must be positive")
	}
	if m.Output.Directory == "" {
		add("output.directory", "is required")
	}
	if m.Comparison.MinimumRequiredPassRate < 0 || m.Comparison.MinimumRequiredPassRate > 1 {
		add("comparison.minimumRequiredPassRate", "must be between 0 and 1")
	}
	if m.Comparison.TieThreshold < 0 || m.Comparison.TieThreshold > 1 {
		add("comparison.tieThreshold", "must be between 0 and 1")
	}
	allowedTieBreakers := map[string]bool{"higherRequiredPassRate": true, "higherMedianScore": true, "lowerMedianDuration": true, "lowerMedianTokens": true, "fewerMedianFilesChanged": true}
	for i, tie := range m.Comparison.TieBreakers {
		if !allowedTieBreakers[tie] {
			add(fmt.Sprintf("comparison.tieBreakers[%d]", i), "unsupported tie-breaker")
		}
	}
	if len(m.Graders) > 0 {
		sum := 0.0
		ids := map[string]bool{}
		for i, g := range m.Graders {
			p := fmt.Sprintf("graders[%d]", i)
			validateID(g.ID, p+".id", ids, add)
			if g.Type != "command" {
				add(p+".type", "must be command")
			}
			if g.Weight <= 0 || g.Weight > 1 {
				add(p+".weight", "must be greater than 0 and at most 1")
			}
			validateCommand(g.Command, p+".command", add)
			for j, asset := range g.Assets {
				requireAny(asset, fmt.Sprintf("%s.assets[%d]", p, j), add)
			}
			sum += g.Weight
		}
		if sum != 1 {
			add("graders", "weights must sum exactly to 1.0 (got "+strconv.FormatFloat(sum, 'g', -1, 64)+")")
		}
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i].Path < ds[j].Path })
	return ds
}

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var placeholderPattern = regexp.MustCompile(`\{[^{}]+\}`)

var allowedPlaceholders = map[string]bool{
	"{task}": true, "{taskFile}": true, "{inputRoot}": true,
	"{workspace}": true, "{caseId}": true, "{variantId}": true,
}

func validateExpectation(e Expectation, path string, add func(string, string)) {
	allowedTypes := map[string]bool{"text": true, "file": true, "directory": true, "json": true, "runtime": true, "image": true, "command": true}
	if !allowedTypes[e.Type] {
		add(path+".type", "must be text, file, directory, json, runtime, image, or command")
	}
	if e.Actual == "" {
		add(path+".actual", "is required")
	}
	if e.Actual != "" && e.Actual != "stdout" && e.Actual != "stderr" && e.Actual != "exitCode" && e.Actual != "submission" && !strings.HasPrefix(e.Actual, "file:") && !strings.HasPrefix(e.Actual, "directory:") {
		add(path+".actual", "must be stdout, stderr, exitCode, submission, file:<path>, or directory:<path>")
	}
	if strings.Contains(e.Actual, ":") {
		_, rel, _ := strings.Cut(e.Actual, ":")
		if rel == "" || !safeRelative(rel) {
			add(path+".actual", "output path must be relative and contained")
		}
	}
	if e.ValueFile != "" {
		requireFile(e.ValueFile, path+".valueFile", add)
	}
	if e.Golden != "" {
		if e.Type == "directory" {
			requireDir(e.Golden, path+".golden", add)
		} else {
			requireFile(e.Golden, path+".golden", add)
		}
	}
	if e.Path != "" && !safeRelative(e.Path) {
		add(path+".path", "must be a relative contained path")
	}
	if e.Command != nil {
		validateCommand(*e.Command, path+".command", add)
	}
	valueSources := 0
	if e.Value != "" {
		valueSources++
	}
	if e.ValueFile != "" {
		valueSources++
	}
	if e.Golden != "" {
		valueSources++
	}
	if valueSources > 1 {
		add(path, "value, valueFile, and golden are mutually exclusive")
	}
	switch e.Type {
	case "text", "json":
		if e.Actual != "stdout" && e.Actual != "stderr" && !strings.HasPrefix(e.Actual, "file:") {
			add(path+".actual", "must locate stdout, stderr, or file:<path>")
		}
		if e.Value == "" && e.ValueFile == "" && e.Golden == "" {
			add(path+".value", "one of value, valueFile, or golden is required")
		}
		if e.Type == "text" && !oneOf(defaultString(e.Operator, "exact"), "exact", "contains", "regex") {
			add(path+".operator", "must be exact, contains, or regex")
		}
		if e.Type == "text" && e.Operator == "regex" && e.Value != "" {
			if _, err := regexp.Compile(e.Value); err != nil {
				add(path+".value", "must be a valid regular expression")
			}
		}
		if e.Type == "json" && !oneOf(defaultString(e.Operator, "exact"), "exact", "subset") {
			add(path+".operator", "must be exact or subset")
		}
	case "runtime":
		if !oneOf(e.Actual, "exitCode", "stdout", "stderr") {
			add(path+".actual", "must be exitCode, stdout, or stderr")
		}
		if e.Actual == "exitCode" && e.ExitCode == nil && e.Value == "" {
			add(path+".exitCode", "is required for exitCode runtime checks")
		}
		if (e.Actual == "stdout" || e.Actual == "stderr") && e.Value == "" && e.ValueFile == "" && e.Golden == "" {
			add(path+".value", "one of value, valueFile, or golden is required")
		}
		if (e.Actual == "stdout" || e.Actual == "stderr") && !oneOf(defaultString(e.Operator, "contains"), "exact", "contains", "regex") {
			add(path+".operator", "must be exact, contains, or regex")
		}
	case "file", "image":
		operator := defaultString(e.Operator, "exists")
		if e.Actual != "submission" && !strings.HasPrefix(e.Actual, "file:") {
			add(path+".actual", "must be submission with path, or file:<path>")
		}
		if e.Actual == "submission" && e.Path == "" {
			add(path+".path", "is required when actual is submission")
		}
		if !oneOf(operator, "exists", "absent", "bytes", "exact", "sha256", "size", "content") {
			add(path+".operator", "unsupported file operator")
		}
		if e.Type == "image" && !oneOf(operator, "bytes", "exact", "sha256") {
			add(path+".operator", "image checks must use bytes, exact, or sha256")
		}
		if (operator == "bytes" || operator == "exact") && e.Golden == "" {
			add(path+".golden", "is required for byte comparison")
		}
		if operator == "sha256" && !regexp.MustCompile(`^[A-Fa-f0-9]{64}$`).MatchString(e.SHA256) {
			add(path+".sha256", "must be a 64-character hexadecimal SHA-256")
		}
		if operator == "content" && e.Value == "" && e.ValueFile == "" && e.Golden == "" {
			add(path+".value", "one of value, valueFile, or golden is required")
		}
	case "directory":
		if e.Actual != "submission" && !strings.HasPrefix(e.Actual, "directory:") {
			add(path+".actual", "must be submission with path, or directory:<path>")
		}
		if e.Actual == "submission" && e.Path == "" {
			add(path+".path", "is required when actual is submission")
		}
		if e.Operator == "count" {
			if _, err := strconv.Atoi(e.Value); err != nil {
				add(path+".value", "must be an integer for count")
			}
		}
		if !oneOf(defaultString(e.Operator, "exists"), "exists", "absent", "count", "inventory", "tree", "paths") {
			add(path+".operator", "unsupported directory operator")
		}
		if (e.Operator == "inventory" || e.Operator == "tree") && e.Golden == "" {
			add(path+".golden", "is required for directory comparison")
		}
	case "command":
		if e.Actual != "submission" {
			add(path+".actual", "must be submission")
		}
		if e.Command == nil {
			add(path+".command", "is required")
		}
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func validateCommand(c Command, path string, add func(string, string)) {
	if c.Executable == "" {
		add(path+".executable", "is required")
	}
	if c.TimeoutSeconds < 0 {
		add(path+".timeoutSeconds", "must not be negative")
	}
	validateArguments(c.Arguments, path+".arguments", add)
	if strings.HasPrefix(filepath.Base(c.Executable), "smaqit-adk") && len(c.Arguments) > 0 && c.Arguments[0] == "bench" {
		add(path, "direct recursive smaqit-adk bench execution is not allowed")
	}
}

func validateArguments(arguments []string, path string, add func(string, string)) {
	for i, argument := range arguments {
		for _, placeholder := range placeholderPattern.FindAllString(argument, -1) {
			if allowedPlaceholders[placeholder] || (strings.HasPrefix(placeholder, "{input:") && idPattern.MatchString(strings.TrimSuffix(strings.TrimPrefix(placeholder, "{input:"), "}"))) {
				continue
			}
			add(fmt.Sprintf("%s[%d]", path, i), "unsupported placeholder "+placeholder)
		}
		withoutPlaceholders := placeholderPattern.ReplaceAllString(argument, "")
		if strings.ContainsAny(withoutPlaceholders, "{}") {
			add(fmt.Sprintf("%s[%d]", path, i), "contains malformed placeholder syntax")
		}
	}
}

func validateID(id, path string, seen map[string]bool, add func(string, string)) {
	if !idPattern.MatchString(id) {
		add(path, "must match "+idPattern.String())
		return
	}
	if seen[id] {
		add(path, "must be unique")
	}
	seen[id] = true
}
func requireFile(path, label string, add func(string, string)) { requirePath(path, label, false, add) }
func requireDir(path, label string, add func(string, string))  { requirePath(path, label, true, add) }
func requireAny(path, label string, add func(string, string))  { requirePathKind(path, label, nil, add) }
func requirePath(path, label string, dir bool, add func(string, string)) {
	requirePathKind(path, label, &dir, add)
}
func requirePathKind(path, label string, dir *bool, add func(string, string)) {
	info, err := os.Lstat(path)
	if err != nil {
		add(label, "does not exist")
		return
	}
	if info.Mode()&os.ModeSymlink != 0 {
		add(label, "must not be a symbolic link")
		return
	}
	if dir != nil && *dir && !info.IsDir() {
		add(label, "must be a directory")
	}
	if dir != nil && !*dir && info.IsDir() {
		add(label, "must be a file")
	}
	if info.IsDir() {
		_ = filepath.WalkDir(path, func(child string, entry os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if entry.Type()&os.ModeSymlink != 0 {
				add(label, "contains symbolic link: "+child)
				return filepath.SkipDir
			}
			return nil
		})
	}
}
func safeRelative(path string) bool {
	if filepath.IsAbs(path) {
		return false
	}
	clean := filepath.Clean(path)
	return clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func pathWithin(parent, child string) bool {
	if parent == "" || child == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	return err == nil && (rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
