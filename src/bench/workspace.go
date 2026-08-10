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
	Root       string
	InputRoot  string
	TaskFile   string
	Inputs     map[string]string
	InputKinds map[string]string
}

func prepareWorkspace(c Case) (*Workspace, error) {
	root, err := os.MkdirTemp("", "smaqit-bench-")
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	w := &Workspace{Root: root, InputRoot: filepath.Join(root, inputDirectoryName), Inputs: map[string]string{}, InputKinds: map[string]string{}}
	fail := func(err error) (*Workspace, error) { _ = removeWorkspace(root); return nil, err }
	if c.Fixture != nil {
		if err := copyDirectory(c.Fixture.Source, root, nil); err != nil {
			return fail(fmt.Errorf("copy fixture: %w", err))
		}
	}
	if err := os.MkdirAll(w.InputRoot, 0700); err != nil {
		return fail(err)
	}
	prompt, err := promptText(c.Given.Prompt)
	if err != nil {
		return fail(err)
	}
	for _, named := range []struct {
		kind   string
		assets []InputAsset
	}{
		{"spec", c.Given.Specs}, {"file", c.Given.Files}, {"directory", c.Given.Directories}, {"image", c.Given.Images},
	} {
		for _, asset := range named.assets {
			destination := asset.Destination
			if destination == "" {
				destination = filepath.Join(named.kind+"s", asset.ID+filepath.Ext(asset.Source))
			}
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
	w.TaskFile = filepath.Join(w.InputRoot, "task.md")
	envelope := renderTaskEnvelope(c.ID, prompt, w.Inputs, w.InputKinds)
	if err := os.WriteFile(w.TaskFile, []byte(envelope), 0400); err != nil {
		return fail(err)
	}
	if err := makeReadOnly(w.InputRoot); err != nil {
		return fail(err)
	}
	return w, nil
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

func renderTaskEnvelope(caseID, prompt string, inputs, mapKinds map[string]string) string {
	var b strings.Builder
	b.WriteString("# Task\n\n")
	fmt.Fprintf(&b, "Case: %s\n\n", caseID)
	b.WriteString(prompt)
	b.WriteString("\n\n# Declared inputs\n")
	ids := make([]string, 0, len(inputs))
	for id := range inputs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "- %s (%s): %s\n", id, mapKinds[id], inputs[id])
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
