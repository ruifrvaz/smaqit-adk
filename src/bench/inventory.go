// captures stable filesystem inventories for submissions and directory checks.
package bench

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TreeEntry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Size   int64  `json:"size,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}
type RepositoryMetrics struct {
	FilesCreated  int   `json:"filesCreated"`
	FilesModified int   `json:"filesModified"`
	FilesDeleted  int   `json:"filesDeleted"`
	FinalFiles    int   `json:"finalFiles"`
	FinalBytes    int64 `json:"finalBytes"`
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
