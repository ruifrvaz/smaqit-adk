// creates isolated workspaces and stages visible inputs without oracle leakage.
package bench

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const inputDirectoryName = ".smaqit-bench-input"

type Workspace struct {
	Root           string
	InputRoot      string
	TreatmentRoot  string
	BriefFile      string
	Inputs         map[string]string
	InputKinds     map[string]string
	Treatments     map[string]string
	TreatmentKinds map[string]string
}

func prepareWorkspace(c Case) (*Workspace, error) {
	root, err := os.MkdirTemp("", "smaqit-bench-")
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	w := &Workspace{
		Root: root, InputRoot: filepath.Join(root, inputDirectoryName),
		Inputs: map[string]string{}, InputKinds: map[string]string{},
		Treatments: map[string]string{}, TreatmentKinds: map[string]string{},
	}
	w.TreatmentRoot = filepath.Join(w.InputRoot, "treatment")
	fail := func(err error) (*Workspace, error) { _ = removeWorkspace(root); return nil, err }
	if c.Fixture != nil {
		destination := root
		if c.Fixture.Destination != "" && c.Fixture.Destination != "." {
			destination, err = containedPath(root, c.Fixture.Destination)
			if err != nil {
				return fail(fmt.Errorf("resolve fixture destination: %w", err))
			}
		}
		if err := copyDirectory(c.Fixture.Source, destination, nil); err != nil {
			return fail(fmt.Errorf("copy fixture: %w", err))
		}
		if err := makeFixtureWritable(destination); err != nil {
			return fail(fmt.Errorf("make fixture writable: %w", err))
		}
	}
	return w, nil
}

func prepareReferenceWorkspace(root string, c Case, v Variant) (*Workspace, func(), error) {
	sidecarRoot, err := os.MkdirTemp("", "smaqit-bench-references-")
	if err != nil {
		return nil, nil, fmt.Errorf("create grading sidecar: %w", err)
	}
	cleanup := func() { _ = removeWorkspace(sidecarRoot) }
	w := &Workspace{
		Root: root, InputRoot: filepath.Join(sidecarRoot, inputDirectoryName),
		Inputs: map[string]string{}, InputKinds: map[string]string{},
		Treatments: map[string]string{}, TreatmentKinds: map[string]string{},
	}
	w.TreatmentRoot = filepath.Join(w.InputRoot, "treatment")
	if err := stageWorkspaceInputs(c, v, w); err != nil {
		cleanup()
		return nil, nil, err
	}
	return w, cleanup, nil
}

func stageWorkspaceInputs(c Case, v Variant, w *Workspace) error {
	fail := func(err error) error { return err }
	if _, err := os.Lstat(w.InputRoot); err == nil {
		return fail(fmt.Errorf("reserved Bench sidecar already exists before staging: %s", w.InputRoot))
	} else if !os.IsNotExist(err) {
		return fail(err)
	}
	if err := os.MkdirAll(w.InputRoot, 0700); err != nil {
		return fail(err)
	}
	prompt, err := promptText(c.Given.Prompt)
	if err != nil {
		return fail(err)
	}
	for _, named := range []struct {
		kind, group string
		assets      []InputAsset
	}{
		{"spec", "specs", c.Given.Specs}, {"file", "files", c.Given.Files}, {"directory", "directories", c.Given.Directories}, {"image", "images", c.Given.Images},
	} {
		for _, asset := range named.assets {
			destination := effectiveInputDestination(named.group, asset.ID, asset.Source, asset.Destination)
			target, err := containedPath(w.InputRoot, destination)
			if err != nil {
				return fail(fmt.Errorf("stage input %s: %w", asset.ID, err))
			}
			info, err := os.Lstat(asset.Source)
			if err != nil {
				return fail(err)
			}
			if info.IsDir() {
				err = copyDirectory(asset.Source, target, nil)
			} else {
				err = copyFile(asset.Source, target, info.Mode().Perm())
			}
			if err != nil {
				return fail(fmt.Errorf("stage input %s: %w", asset.ID, err))
			}
			w.Inputs[asset.ID] = target
			label := named.kind
			if asset.MediaType != "" {
				label += "; " + asset.MediaType
			}
			w.InputKinds[asset.ID] = label
		}
	}
	for _, treatment := range v.Treatment {
		target, err := containedPath(w.TreatmentRoot, treatment.ID)
		if err != nil {
			return fail(fmt.Errorf("stage treatment %s: %w", treatment.ID, err))
		}
		info, err := os.Lstat(treatment.Source)
		if err != nil {
			return fail(err)
		}
		if info.IsDir() {
			err = copyDirectory(treatment.Source, target, nil)
		} else {
			err = copyFile(treatment.Source, target, info.Mode().Perm())
		}
		if err != nil {
			return fail(fmt.Errorf("stage treatment %s: %w", treatment.ID, err))
		}
		w.Treatments[treatment.ID] = target
		label := "artifact"
		if info.IsDir() {
			label = "directory"
		}
		if treatment.MediaType != "" {
			label += "; " + treatment.MediaType
		}
		w.TreatmentKinds[treatment.ID] = label
	}
	w.BriefFile = filepath.Join(w.InputRoot, "brief.md")
	brief := renderCaseBrief(c.ID, prompt, w.Inputs, w.InputKinds, w.Treatments, w.TreatmentKinds)
	if err := os.WriteFile(w.BriefFile, []byte(brief), 0400); err != nil {
		return fail(err)
	}
	if err := makeReadOnly(w.InputRoot); err != nil {
		return fail(err)
	}
	return nil
}

func promptText(p Prompt) (string, error) {
	if p.Text != "" {
		return p.Text, nil
	}
	b, err := os.ReadFile(p.File)
	if err != nil {
		return "", fmt.Errorf("read prompt: %w", err)
	}
	return string(b), nil
}

func renderCaseBrief(caseID, prompt string, inputs, inputKinds, treatments, treatmentKinds map[string]string) string {
	var b strings.Builder
	b.WriteString("# Case brief\n\n")
	fmt.Fprintf(&b, "Case: %s\n\n", caseID)
	b.WriteString(prompt)
	b.WriteString("\n\n# Shared declared inputs\n")
	ids := make([]string, 0, len(inputs))
	for id := range inputs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "- %s (%s): %s\n", id, inputKinds[id], inputs[id])
	}
	if len(ids) == 0 {
		b.WriteString("- None.\n")
	}
	b.WriteString("\n# Variant treatment artifacts\n")
	ids = ids[:0]
	for id := range treatments {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "- %s (%s): %s\n", id, treatmentKinds[id], treatments[id])
	}
	if len(ids) == 0 {
		b.WriteString("- None.\n")
	}
	return b.String()
}

func containedPath(root, rel string) (string, error) {
	if !safeRelative(rel) {
		return "", fmt.Errorf("path escapes root: %s", rel)
	}
	target := filepath.Join(root, filepath.Clean(rel))
	relToRoot, err := filepath.Rel(root, target)
	if err != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", rel)
	}
	return target, nil
}

func copyFile(source, target string, mode fs.FileMode) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not allowed: %s", source)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("non-regular files are not allowed: %s", source)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func copyDirectory(source, target string, exclude func(string) bool) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel != "." && exclude != nil && exclude(filepath.ToSlash(rel)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not allowed: %s", path)
		}
		destination := target
		if rel != "." {
			destination = filepath.Join(target, rel)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(destination, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular files are not allowed: %s", path)
		}
		return copyFile(path, destination, info.Mode().Perm())
	})
}

func makeReadOnly(root string) error {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		if entry.IsDir() {
			return nil
		}
		return os.Chmod(path, 0400)
	})
	if err != nil {
		return err
	}
	for i := len(paths) - 1; i >= 0; i-- {
		info, err := os.Stat(paths[i])
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := os.Chmod(paths[i], 0500); err != nil {
				return err
			}
		}
	}
	return nil
}

func makeEvidenceReadOnly(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0700)
		}
		return os.Chmod(path, 0400)
	})
}

func makeWritable(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0700)
		}
		return os.Chmod(path, 0600)
	})
}

func makeFixtureWritable(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := info.Mode().Perm()
		if entry.IsDir() {
			mode |= 0700
		} else {
			mode |= 0200
		}
		return os.Chmod(path, mode)
	})
}

// removeWorkspace restores the permissions deliberately removed from staged
// inputs before deleting an ephemeral attempt workspace. os.RemoveAll cannot
// remove a file from a read-only input directory on its own.
func removeWorkspace(root string) error {
	if _, err := os.Lstat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := makeWritable(root); err != nil {
		return fmt.Errorf("make workspace writable for cleanup: %w", err)
	}
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	return nil
}
