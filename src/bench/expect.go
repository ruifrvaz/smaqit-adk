// evaluates deterministic required expectations and command graders.
package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type GradeResult struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Passed  bool    `json:"passed"`
	Score   float64 `json:"score"`
	Message string  `json:"message,omitempty"`
}

func gradeExpectations(ctx context.Context, expectations []Expectation, harness HarnessResult, submission string, request RunRequest) ([]GradeResult, error) {
	results := make([]GradeResult, 0, len(expectations))
	for _, expectation := range expectations {
		result, err := gradeExpectation(ctx, expectation, harness, submission, request)
		if err != nil {
			return nil, fmt.Errorf("grade expectation %s: %w", expectation.ID, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func gradeExpectation(ctx context.Context, expectation Expectation, harness HarnessResult, submission string, request RunRequest) (GradeResult, error) {
	result := GradeResult{ID: expectation.ID, Type: expectation.Type}
	pass := false
	message := "expectation did not match"
	switch expectation.Type {
	case "text":
		actual, err := actualText(expectation.Actual, harness, submission)
		if err != nil {
			return result, err
		}
		expected, err := expectedText(expectation)
		if err != nil {
			return result, err
		}
		actual, expected = normalizeText(actual, expected, expectation)
		switch defaultString(expectation.Operator, "exact") {
		case "exact":
			pass = actual == expected
		case "contains":
			pass = strings.Contains(actual, expected)
		case "regex":
			matched, err := regexp.MatchString(expected, actual)
			if err != nil {
				return result, err
			}
			pass = matched
		default:
			return result, fmt.Errorf("unsupported text operator %q", expectation.Operator)
		}
	case "json":
		actualText, err := actualText(expectation.Actual, harness, submission)
		if err != nil {
			return result, err
		}
		expectedText, err := expectedText(expectation)
		if err != nil {
			return result, err
		}
		var actual, expected any
		if err := decodeStrictJSON([]byte(actualText), &actual); err != nil {
			return result, fmt.Errorf("actual JSON: %w", err)
		}
		if err := decodeStrictJSON([]byte(expectedText), &expected); err != nil {
			return result, fmt.Errorf("expected JSON: %w", err)
		}
		if defaultString(expectation.Operator, "exact") == "subset" {
			pass = jsonSubset(expected, actual)
		} else {
			pass = valuesEqual(expected, actual)
		}
	case "runtime":
		switch expectation.Actual {
		case "exitCode":
			expected, err := expectedExitCode(expectation)
			if err != nil {
				return result, err
			}
			pass = harness.ExitCode == expected
		case "stdout", "stderr":
			actual, err := actualText(expectation.Actual, harness, submission)
			if err != nil {
				return result, err
			}
			expected, err := expectedText(expectation)
			if err != nil {
				return result, err
			}
			actual, expected = normalizeText(actual, expected, expectation)
			pass = textCompare(defaultString(expectation.Operator, "contains"), actual, expected)
		default:
			return result, fmt.Errorf("runtime actual must be exitCode, stdout, or stderr")
		}
	case "file", "image":
		path, err := actualPath(expectation, submission)
		if err != nil {
			return result, err
		}
		pass, message, err = gradeFile(expectation, path)
		if err != nil {
			return result, err
		}
	case "directory":
		path, err := actualPath(expectation, submission)
		if err != nil {
			return result, err
		}
		pass, message, err = gradeDirectory(expectation, path)
		if err != nil {
			return result, err
		}
	case "command":
		if expectation.Command == nil {
			return result, fmt.Errorf("command is required")
		}
		copyRoot, err := os.MkdirTemp("", "smaqit-bench-grade-")
		if err != nil {
			return result, err
		}
		defer os.RemoveAll(copyRoot)
		if err := copyDirectory(submission, copyRoot, nil); err != nil {
			return result, err
		}
		if err := makeWritable(copyRoot); err != nil {
			return result, err
		}
		gradeRequest := request
		gradeRequest.Workspace = &Workspace{Root: copyRoot, InputRoot: request.Workspace.InputRoot, TreatmentRoot: request.Workspace.TreatmentRoot, BriefFile: request.Workspace.BriefFile, Inputs: request.Workspace.Inputs, InputKinds: request.Workspace.InputKinds, Treatments: request.Workspace.Treatments, TreatmentKinds: request.Workspace.TreatmentKinds}
		commandResult, err := executeCommand(ctx, *expectation.Command, gradeRequest, "expect-"+expectation.ID)
		if err != nil {
			return result, err
		}
		expected, err := expectedExitCode(expectation)
		if err != nil {
			return result, err
		}
		pass = commandResult.ExitCode == expected
	default:
		return result, fmt.Errorf("unsupported type %q", expectation.Type)
	}
	result.Passed = pass
	if pass {
		result.Score = 1
		result.Message = "passed"
	} else {
		result.Message = message
	}
	return result, nil
}

func actualText(locator string, harness HarnessResult, submission string) (string, error) {
	switch locator {
	case "stdout":
		return harness.Stdout, nil
	case "stderr":
		return harness.Stderr, nil
	}
	if strings.HasPrefix(locator, "file:") {
		b, err := os.ReadFile(filepath.Join(submission, filepath.FromSlash(strings.TrimPrefix(locator, "file:"))))
		return string(b), err
	}
	return "", fmt.Errorf("unsupported text locator %q", locator)
}

func actualPath(expectation Expectation, submission string) (string, error) {
	rel := expectation.Path
	if strings.Contains(expectation.Actual, ":") {
		_, rel, _ = strings.Cut(expectation.Actual, ":")
	}
	if rel == "" {
		return "", fmt.Errorf("path is required")
	}
	return containedPath(submission, filepath.FromSlash(rel))
}

func expectedText(expectation Expectation) (string, error) {
	if expectation.ValueFile != "" {
		b, err := os.ReadFile(expectation.ValueFile)
		return string(b), err
	}
	if expectation.Golden != "" {
		b, err := os.ReadFile(expectation.Golden)
		return string(b), err
	}
	return expectation.Value, nil
}
func normalizeText(actual, expected string, e Expectation) (string, string) {
	if e.TrimFinalLine {
		actual = strings.TrimSuffix(strings.TrimSuffix(actual, "\n"), "\r")
		expected = strings.TrimSuffix(strings.TrimSuffix(expected, "\n"), "\r")
	}
	if e.IgnoreCase {
		actual = strings.ToLower(actual)
		expected = strings.ToLower(expected)
	}
	return actual, expected
}
func textCompare(operator, actual, expected string) bool {
	switch operator {
	case "exact":
		return actual == expected
	case "contains":
		return strings.Contains(actual, expected)
	case "regex":
		ok, _ := regexp.MatchString(expected, actual)
		return ok
	}
	return false
}
func expectedExitCode(e Expectation) (int, error) {
	if e.ExitCode != nil {
		return *e.ExitCode, nil
	}
	if e.Value != "" {
		return strconv.Atoi(e.Value)
	}
	return 0, nil
}

func gradeFile(e Expectation, path string) (bool, string, error) {
	info, err := os.Stat(path)
	operator := defaultString(e.Operator, "exists")
	if operator == "absent" {
		return os.IsNotExist(err), "file exists", nil
	}
	if err != nil {
		return false, "file does not exist", nil
	}
	if info.IsDir() {
		return false, "path is a directory", nil
	}
	switch operator {
	case "exists":
		return true, "", nil
	case "bytes", "exact":
		expected, err := os.ReadFile(e.Golden)
		if err != nil {
			return false, "", err
		}
		actual, err := os.ReadFile(path)
		return bytes.Equal(actual, expected), "file bytes differ", err
	case "sha256":
		digest, err := digestFile(path)
		return strings.EqualFold(digest, e.SHA256), "file hash differs", err
	case "size":
		size, err := strconv.ParseInt(e.Value, 10, 64)
		if err != nil {
			return false, "", err
		}
		return info.Size() == size, "file size differs", nil
	case "content":
		actual, err := os.ReadFile(path)
		if err != nil {
			return false, "", err
		}
		expected, err := expectedText(e)
		if err != nil {
			return false, "", err
		}
		a, x := normalizeText(string(actual), expected, e)
		return textCompare("exact", a, x), "file content differs", nil
	}
	return false, "", fmt.Errorf("unsupported file operator %q", operator)
}

func gradeDirectory(e Expectation, path string) (bool, string, error) {
	info, err := os.Stat(path)
	operator := defaultString(e.Operator, "exists")
	if operator == "absent" {
		return os.IsNotExist(err), "directory exists", nil
	}
	if err != nil {
		return false, "directory does not exist", nil
	}
	if !info.IsDir() {
		return false, "path is not a directory", nil
	}
	if operator == "exists" {
		return true, "", nil
	}
	inventory, err := directoryInventory(path)
	if err != nil {
		return false, "", err
	}
	switch operator {
	case "count":
		expected, err := strconv.Atoi(e.Value)
		if err != nil {
			return false, "", err
		}
		return len(inventory) == expected, "directory count differs", nil
	case "inventory":
		if e.Golden == "" {
			return false, "", fmt.Errorf("golden directory is required")
		}
		golden, err := directoryInventory(e.Golden)
		if err != nil {
			return false, "", err
		}
		return valuesEqual(inventory, golden), "directory inventory differs", nil
	case "tree":
		if e.Golden == "" {
			return false, "", fmt.Errorf("golden directory is required")
		}
		actualDigest, err := digestPath(path)
		if err != nil {
			return false, "", err
		}
		goldenDigest, err := digestPath(e.Golden)
		if err != nil {
			return false, "", err
		}
		return actualDigest == goldenDigest, "directory tree differs", nil
	case "paths":
		set := map[string]bool{}
		for _, p := range inventory {
			set[p] = true
		}
		for _, p := range e.RequiredPaths {
			if !set[p] {
				return false, "required path missing: " + p, nil
			}
		}
		for _, p := range e.ForbiddenPaths {
			if set[p] {
				return false, "forbidden path present: " + p, nil
			}
		}
		return true, "", nil
	}
	return false, "", fmt.Errorf("unsupported directory operator %q", operator)
}
func directoryInventory(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in output: %s", path)
		}
		rel, _ := filepath.Rel(root, path)
		suffix := ""
		if d.IsDir() {
			suffix = "/"
		}
		out = append(out, filepath.ToSlash(rel)+suffix)
		return nil
	})
	sort.Strings(out)
	return out, err
}
func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("multiple JSON values")
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("multiple JSON values")
	} else if err != io.EOF {
		return err
	}
	return nil
}
func jsonSubset(expected, actual any) bool {
	switch e := expected.(type) {
	case map[string]any:
		a, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for key, value := range e {
			actualValue, exists := a[key]
			if !exists || !jsonSubset(value, actualValue) {
				return false
			}
		}
		return true
	case []any:
		a, ok := actual.([]any)
		if !ok || len(e) > len(a) {
			return false
		}
		for i := range e {
			if !jsonSubset(e[i], a[i]) {
				return false
			}
		}
		return true
	default:
		return valuesEqual(expected, actual)
	}
}
func valuesEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
