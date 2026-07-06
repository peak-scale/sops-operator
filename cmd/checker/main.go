// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peak-scale/sops-operator/internal/sopschecker"
)

func main() {
	var (
		configPath string
		requireAll bool
		globs      stringListFlag
	)

	flag.StringVar(&configPath, "config", "", "Path to a .sops.yaml file. If omitted, the checker discovers .sops.yaml from each file path.")
	flag.Var(&globs, "glob", "File glob pattern to check. Can be provided multiple times. When set, positional files are ignored.")
	flag.BoolVar(&requireAll, "require-all", false, "Require every provided file to be SOPS encrypted, without checking .sops.yaml creation rules.")
	flag.Parse()

	failures, err := sopschecker.Check(sopschecker.Options{
		ConfigPath: configPath,
		Files:      flag.Args(),
		Globs:      globs,
		RequireAll: requireAll,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "sops-checker: %v\n", err)
		os.Exit(2)
	}

	if len(failures) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "sops-checker: files requiring SOPS encryption are not encrypted:")

	for _, failure := range failures {
		fmt.Fprintf(os.Stderr, "  - %s: %s\n", failure.Path, failure.Reason)
	}

	os.Exit(1)
}

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	*f = append(*f, value)

	return nil
}
