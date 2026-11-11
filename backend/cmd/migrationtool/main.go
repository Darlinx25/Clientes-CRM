// Command migrationtool checks and maintains the migration safety manifest.
//
// Usage (from backend/):
//
//	go run ./cmd/migrationtool lint
//	go run ./cmd/migrationtool checksum            # verify manifest
//	go run ./cmd/migrationtool checksum -update    # append new files only
//	go run ./cmd/migrationtool checksum -update -force
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"meerkat/database"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "lint":
		runLint()
	case "checksum":
		runChecksum(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrationtool <lint|checksum> [-update] [-force]")
}

func runLint() {
	problems := database.CheckMigrations()
	if err := database.VerifyMigrationChecksums(); err != nil {
		problems = append(problems, err)
	}
	if len(problems) == 0 {
		fmt.Println("migrations: OK")
		return
	}
	for _, problem := range problems {
		fmt.Fprintf(os.Stderr, "migration error: %v\n", problem)
	}
	os.Exit(1)
}

func runChecksum(args []string) {
	flags := flag.NewFlagSet("checksum", flag.ExitOnError)
	update := flags.Bool("update", false, "write the manifest")
	force := flags.Bool("force", false, "allow changing/removing existing entries (manual review required)")
	_ = flags.Parse(args)

	current, err := database.MigrationChecksums()
	if err != nil {
		fatal(err)
	}
	manifestPath := "database/migrations.lock"

	var previous map[string]string
	raw, readErr := os.ReadFile(manifestPath)
	if readErr == nil {
		previous = parseManifest(string(raw))
	} else if !os.IsNotExist(readErr) {
		fatal(readErr)
	}

	if !*update {
		if err := database.VerifyMigrationChecksums(); err != nil {
			fatal(err)
		}
		fmt.Println("migration checksums: OK")
		return
	}

	if previous == nil {
		previous = map[string]string{}
	}
	for name, oldSum := range previous {
		newSum, ok := current[name]
		if !ok {
			if !*force {
				fatal(fmt.Errorf("refusing to remove %q from the manifest (use -force after manual review)", name))
			}
			continue
		}
		if newSum != oldSum && !*force {
			fatal(fmt.Errorf("refusing to change historical migration %q (use -force after manual review)", name))
		}
	}

	var lines []string
	lines = append(lines, "# SHA256 manifest of reviewed migration files. Do not edit by hand.")
	lines = append(lines, "# Regenerate only when adding a NEW migration: make migrations-checksum")
	names := make([]string, 0, len(current))
	for name := range current {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s  %s", current[name], name))
	}
	if err := os.WriteFile(manifestPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%d migrations)\n", manifestPath, len(current))
}

func parseManifest(raw string) map[string]string {
	sums := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 {
			sums[fields[1]] = fields[0]
		}
	}
	return sums
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "migration error: %v\n", err)
	os.Exit(1)
}
