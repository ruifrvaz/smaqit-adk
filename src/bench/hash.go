// hashes files, directories, and executable identities for plan drift checks.
package bench

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AssetDigest struct {
	Kind     string   `json:"kind"`
	Path     string   `json:"path"`
	SHA256   string   `json:"sha256"`
	Excludes []string `json:"excludes,omitempty"`
}

func digestPath(path string) (string, error) {
	return digestPathExcluding(path, nil)
}

func digestPathExcluding(path string, excludes []string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("symlinks are not allowed: %s", path)
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("non-regular files are not allowed: %s", path)
		}
		return digestFile(path)
	}
	var paths []string
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not allowed: %s", p)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular files are not allowed: %s", p)
		}
		if p != path && matchesExclusion(p, excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if p != path {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, p := range paths {
		rel, _ := filepath.Rel(path, p)
		info, err := os.Stat(p)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), info.Mode().Type())
		if !info.IsDir() {
			f, err := os.Open(p)
			if err != nil {
				return "", err
			}
			_, copyErr := io.Copy(h, f)
			closeErr := f.Close()
			if copyErr != nil {
				return "", copyErr
			}
			if closeErr != nil {
				return "", closeErr
			}
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func matchesExclusion(path string, excludes []string) bool {
	for _, excluded := range excludes {
		rel, err := filepath.Rel(excluded, path)
		if err == nil && (rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return true
		}
	}
	return false
}

func digestFile(path string) (string, error) {
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
