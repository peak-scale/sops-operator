// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package sopschecker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/config"
)

// Options configures a SOPS encryption check.
type Options struct {
	// ConfigPath is an optional path to a .sops.yaml file. When set, files are
	// only required to be encrypted if they match a creation rule in the config.
	ConfigPath string
	// Files contains the files to check.
	Files []string
	// Globs contains file glob patterns to check. When set, Files is ignored.
	Globs []string
	// RequireAll requires every file in Files to be encrypted, independent of
	// any .sops.yaml creation rules.
	RequireAll bool
	// WorkDir is used to resolve relative file and config paths. It defaults to
	// the current working directory.
	WorkDir string
}

// Failure describes a file that should have been encrypted but was not.
type Failure struct {
	Path   string
	Reason string
}

// Check validates that files requiring SOPS encryption are encrypted.
func Check(opts Options) ([]Failure, error) {
	if len(opts.Files) == 0 && len(opts.Globs) == 0 {
		return nil, errors.New("no files or globs provided")
	}

	workDir := opts.WorkDir
	if workDir == "" {
		var err error

		workDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	var configPath string

	if opts.ConfigPath != "" {
		resolved, err := resolvePath(workDir, opts.ConfigPath)
		if err != nil {
			return nil, err
		}

		configPath = resolved
	}

	files := opts.Files

	if len(opts.Globs) > 0 {
		globbedFiles, err := expandGlobs(workDir, opts.Globs)
		if err != nil {
			return nil, err
		}

		files = globbedFiles
	}

	failures := make([]Failure, 0)

	for _, file := range files {
		path, err := resolvePath(workDir, file)
		if err != nil {
			return nil, err
		}

		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", file, err)
		}

		if info.IsDir() {
			continue
		}

		required, storesConfig, err := encryptionRequired(path, configPath, opts.RequireAll)
		if err != nil {
			return nil, err
		}

		if !required {
			continue
		}

		encrypted, reason := isEncrypted(path, storesConfig)
		if !encrypted {
			failures = append(failures, Failure{
				Path:   displayPath(workDir, path),
				Reason: reason,
			})
		}
	}

	return failures, nil
}

func expandGlobs(workDir string, patterns []string) ([]string, error) {
	seen := make(map[string]struct{})

	for _, pattern := range patterns {
		resolvedPattern := pattern
		if !filepath.IsAbs(resolvedPattern) {
			resolvedPattern = filepath.Join(workDir, resolvedPattern)
		}

		matches, err := filepath.Glob(resolvedPattern)
		if err != nil {
			return nil, fmt.Errorf("expand glob %q: %w", pattern, err)
		}

		for _, match := range matches {
			abs, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("resolve glob match %q: %w", match, err)
			}

			seen[abs] = struct{}{}
		}
	}

	files := make([]string, 0, len(seen))
	for file := range seen {
		files = append(files, file)
	}

	sort.Strings(files)

	return files, nil
}

func encryptionRequired(filePath, configuredPath string, requireAll bool) (bool, *config.StoresConfig, error) {
	if requireAll {
		return true, config.NewStoresConfig(), nil
	}

	configPath := configuredPath
	if configPath == "" {
		result, _ := config.LookupConfigFile(filePath)
		if result.Path == "" {
			return false, nil, nil
		}

		configPath = result.Path
	}

	creationRule, err := config.LoadCreationRuleForFile(configPath, filePath, nil)
	if err != nil {
		if isNoMatchingRuleError(err) {
			return false, nil, nil
		}

		return false, nil, fmt.Errorf("load creation rule for %q: %w", filePath, err)
	}

	if creationRule == nil {
		return false, nil, nil
	}

	storesConfig, err := config.LoadStoresConfig(configPath)
	if err != nil {
		return false, nil, fmt.Errorf("load stores config %q: %w", configPath, err)
	}

	return true, storesConfig, nil
}

func isEncrypted(path string, storesConfig *config.StoresConfig) (bool, string) {
	if storesConfig == nil {
		storesConfig = config.NewStoresConfig()
	}

	store := common.DefaultStoreForPath(storesConfig, path)
	if _, err := common.LoadEncryptedFile(store, path); err != nil {
		return false, err.Error()
	}

	return true, ""
}

func isNoMatchingRuleError(err error) bool {
	return strings.Contains(err.Error(), "no matching creation rules found")
}

func resolvePath(workDir, path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}

	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", path, err)
	}

	return resolved, nil
}

func displayPath(workDir, path string) string {
	rel, err := filepath.Rel(workDir, path)
	if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return path
	}

	return rel
}
