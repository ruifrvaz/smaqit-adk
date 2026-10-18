// captures stable filesystem inventories for submissions and directory checks.
package bench

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TreeEntry is one inventoried workspace path. Kind is "file", "directory",
// or "special" for a non-regular file (socket, FIFO, or device) that a harness
// left behind; a special entry carries its Type but is never digested, copied
// into a frozen submission, or counted as a file.
type TreeEntry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Type   string `json:"type,omitempty"`
	Size   int64  `json:"size,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}
type RepositoryMetrics struct {
	FilesCreated  int   `json:"filesCreated"`
	FilesModified int   `json:"filesModified"`
	FilesDeleted  int   `json:"filesDeleted"`
	FinalFiles    int   `json:"finalFiles"`
	FinalBytes    int64 `json:"finalBytes"`
	SpecialFiles  int   `json:"specialFiles,omitempty"`
}

func snapshotTree(root string) ([]TreeEntry, error) {
	var result []TreeEntry
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == inputDirectoryName || strings.HasPrefix(rel, inputDirectoryName+"/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		item := TreeEntry{Path: rel, Size: info.Size()}
		if entry.IsDir() {
			item.Kind = "directory"
		} else if special := specialFileType(path, info); special != "" {
			item.Kind, item.Type, item.Size = "special", special, 0
		} else {
			item.Kind = "file"
			item.SHA256, err = digestFile(path)
			if err != nil {
				return err
			}
		}
		result = append(result, item)
		return nil
	})
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, err
}

// specialFileType names a non-regular file that cannot be opened and digested
// like a regular one, or returns "" for regular files and symbolic links. A
// symbolic link keeps its existing handling (digested through its target)
// unless that target is itself non-regular.
func specialFileType(path string, info os.FileInfo) string {
	mode := info.Mode()
	if mode&os.ModeSymlink != 0 {
		target, err := os.Stat(path)
		if err != nil || target.Mode().IsRegular() || target.IsDir() {
			return ""
		}
		mode = target.Mode()
	}
	switch {
	case mode.IsRegular():
		return ""
	case mode&os.ModeSocket != 0:
		return "socket"
	case mode&os.ModeNamedPipe != 0:
		return "fifo"
	case mode&os.ModeDevice != 0:
		return "device"
	default:
		return "other"
	}
}

// specialPaths returns the workspace-relative paths of every special entry so
// the freeze step can leave them out of the submission copy.
func specialPaths(tree []TreeEntry) map[string]bool {
	paths := map[string]bool{}
	for _, item := range tree {
		if item.Kind == "special" {
			paths[item.Path] = true
		}
	}
	return paths
}

func compareTrees(before, after []TreeEntry) RepositoryMetrics {
	old := map[string]TreeEntry{}
	for _, item := range before {
		if item.Kind == "file" {
			old[item.Path] = item
		}
	}
	metrics := RepositoryMetrics{}
	seen := map[string]bool{}
	for _, item := range after {
		if item.Kind == "special" {
			metrics.SpecialFiles++
		}
		if item.Kind != "file" {
			continue
		}
		metrics.FinalFiles++
		metrics.FinalBytes += item.Size
		seen[item.Path] = true
		previous, exists := old[item.Path]
		if !exists {
			metrics.FilesCreated++
		} else if previous.SHA256 != item.SHA256 {
			metrics.FilesModified++
		}
	}
	for path := range old {
		if !seen[path] {
			metrics.FilesDeleted++
		}
	}
	return metrics
}
