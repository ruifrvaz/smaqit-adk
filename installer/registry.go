// Terraform-statefile-style registry tracking every ADK-compiled agent or
// skill's source dependencies and compiled-content hash, enabling drift
// detection (`smaqit-adk sync`) and automatic recompilation cascades
// (`smaqit.new-principle`/`new-template`/`new-rules`). Modeled on
// src/bench/plan.go's atomic JSON write and src/bench/hash.go's file
// hashing — not a direct import, since those helpers are unexported and
// installer is a separate Go module.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const RegistrySchemaVersion = 1

// Registry is one scope's persisted state: either the ADK-global registry
// (tracking smaqit.L0/L1/L2 and the 5 shipped skills against the base
// framework) or a single consuming project's registry (tracking its own
// create-agent/create-skill output against that project's framework
// extension).
type Registry struct {
	SchemaVersion int             `json:"schemaVersion"`
	Scope         string          `json:"scope"` // "base" | "extension"
	Entries       []RegistryEntry `json:"entries"`
}

// RegistryTarget is the compiled output path(s) for one registered resource.
// An agent has both; a skill (single cross-platform SKILL.md) has only Claude set.
type RegistryTarget struct {
	Claude string `json:"claude,omitempty"`
	Codex  string `json:"codex,omitempty"`
}

// RegistryEntry records one compiled agent or skill: what it was compiled
// from, and a content hash of its primary compiled output for drift
// detection.
type RegistryEntry struct {
	Name             string         `json:"name"`
	Target           RegistryTarget `json:"target"`
	SourceDefinition string         `json:"sourceDefinition,omitempty"`
	SourceTemplates  []string       `json:"sourceTemplates,omitempty"`
	CompiledHash     string         `json:"compiledHash"`
	LastCompiled     string         `json:"lastCompiled"`
}

// ReadRegistry reads the registry at path, or returns an empty registry if
// the file does not exist yet — a fresh install or a project that has never
// registered anything has no registry file, and that is not an error.
func ReadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{SchemaVersion: RegistrySchemaVersion}, nil
		}
		return nil, err
	}
	var r Registry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&r); err != nil {
		return nil, fmt.Errorf("parsing registry %s: %w", path, err)
	}
	if r.SchemaVersion != RegistrySchemaVersion {
		return nil, fmt.Errorf("unsupported registry schema version %d in %s", r.SchemaVersion, path)
	}
	return &r, nil
}

// WriteRegistry atomically writes the registry to path.
func WriteRegistry(path string, r *Registry) error {
	return writeRegistryJSONAtomic(path, r)
}

// PutEntry inserts or replaces the entry with the given Name, keeping
// Entries sorted by name for stable diffs.
func (r *Registry) PutEntry(e RegistryEntry) {
	for i := range r.Entries {
		if r.Entries[i].Name == e.Name {
			r.Entries[i] = e
			return
		}
	}
	r.Entries = append(r.Entries, e)
	sort.Slice(r.Entries, func(i, j int) bool { return r.Entries[i].Name < r.Entries[j].Name })
}

// DependentsOf returns every entry whose SourceTemplates includes sourcePath.
func (r *Registry) DependentsOf(sourcePath string) []RegistryEntry {
	var out []RegistryEntry
	for _, e := range r.Entries {
		for _, t := range e.SourceTemplates {
			if t == sourcePath {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

func digestFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeRegistryJSONAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.CreateTemp(filepath.Dir(path), ".registry-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
