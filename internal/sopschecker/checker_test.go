// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package sopschecker

import (
	"os"
	"path/filepath"
	"testing"
)

const testAgeRecipient = "age1s7t2vk2crlxaumgm7cacs568xwutkjs535pla69kt6w006t7wgzqhkfwvp"

func TestCheckFailsForMatchingPlaintextFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `creation_rules:
  - path_regex: \.secret\.yaml$
    age: `+testAgeRecipient+`
`)
	writeFile(t, dir, "app.secret.yaml", "apiVersion: v1\nkind: Secret\n")

	failures, err := Check(Options{
		WorkDir: dir,
		Files:   []string{"app.secret.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected one failure, got %d", len(failures))
	}
	if failures[0].Path != "app.secret.yaml" {
		t.Fatalf("unexpected failure path: %s", failures[0].Path)
	}
}

func TestCheckPassesForMatchingEncryptedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `creation_rules:
  - path_regex: \.secret\.yaml$
    age: `+testAgeRecipient+`
`)
	writeFile(t, dir, "app.secret.yaml", encryptedYAML())

	failures, err := Check(Options{
		WorkDir: dir,
		Files:   []string{"app.secret.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d", len(failures))
	}
}

func TestCheckSkipsFilesWithoutMatchingCreationRule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `creation_rules:
  - path_regex: \.secret\.yaml$
    age: `+testAgeRecipient+`
`)
	writeFile(t, dir, "deployment.yaml", "apiVersion: apps/v1\nkind: Deployment\n")

	failures, err := Check(Options{
		WorkDir: dir,
		Files:   []string{"deployment.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d", len(failures))
	}
}

func TestCheckSkipsConfigWithoutCreationRules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `destination_rules:
  - s3_bucket: example
`)
	writeFile(t, dir, "secret.yaml", "apiVersion: v1\nkind: Secret\n")

	failures, err := Check(Options{
		WorkDir: dir,
		Files:   []string{"secret.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d", len(failures))
	}
}

func TestCheckRequireAllFailsWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "secret.yaml", "apiVersion: v1\nkind: Secret\n")

	failures, err := Check(Options{
		WorkDir:    dir,
		Files:      []string{"secret.yaml"},
		RequireAll: true,
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected one failure, got %d", len(failures))
	}
}

func TestCheckExpandsGlobs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `creation_rules:
  - path_regex: secrets/.*\.yaml$
    age: `+testAgeRecipient+`
`)
	writeFile(t, dir, "secrets/app.secret.yaml", "apiVersion: v1\nkind: Secret\n")
	writeFile(t, dir, "secrets/app.enc.yaml", encryptedYAML())
	writeFile(t, dir, "other/ignored.yaml", "apiVersion: v1\nkind: Secret\n")

	failures, err := Check(Options{
		WorkDir: dir,
		Globs:   []string{"secrets/*.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected one failure, got %d", len(failures))
	}
	if failures[0].Path != "secrets/app.secret.yaml" {
		t.Fatalf("unexpected failure path: %s", failures[0].Path)
	}
}

func TestCheckIgnoresFilesWhenGlobsAreProvided(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".sops.yaml", `creation_rules:
  - path_regex: .*\.yaml$
    age: `+testAgeRecipient+`
`)
	writeFile(t, dir, "selected.yaml", encryptedYAML())
	writeFile(t, dir, "ignored-positional.yaml", "apiVersion: v1\nkind: Secret\n")

	failures, err := Check(Options{
		WorkDir: dir,
		Files:   []string{"ignored-positional.yaml"},
		Globs:   []string{"selected.yaml"},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %d", len(failures))
	}
}

func writeFile(t *testing.T, dir, name, data string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func encryptedYAML() string {
	return `apiVersion: ENC[AES256_GCM,data:abcd,iv:abcd,tag:abcd,type:str]
sops:
    age:
        - recipient: ` + testAgeRecipient + `
          enc: encrypted-data-key
    lastmodified: "2025-09-02T11:21:15Z"
    mac: ENC[AES256_GCM,data:abcd,iv:abcd,tag:abcd,type:str]
    encrypted_regex: ^(data|stringData|apiVersion)$
    version: 3.8.1
`
}
