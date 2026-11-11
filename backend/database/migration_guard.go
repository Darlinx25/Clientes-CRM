package database

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// migrationsLockFS embeds the checksum manifest so a build can never silently
// diverge from the reviewed migration history.
//
//go:embed migrations.lock
var migrationsLockFS embed.FS

const migrationsLockName = "migrations.lock"

// migrationLintFromVersion is the first migration version that is subject to
// the destructive-statement lint. Everything before it is historical and
// immutable (protected by the checksum manifest); new migrations must comply
// with the upgrade-safety rules documented in docs/upgrades.md.
const migrationLintFromVersion = 40

var (
	migrationFileRe = regexp.MustCompile(`^(\d{6})_(.+)\.(up|down)\.sql$`)
	// destructiveSQLRe matches statements that are unsafe in an upgrade path.
	// A migration may opt in explicitly with a comment such as:
	//   -- migration:allow-destructive reason=historial data migration
	destructiveSQLRe   = regexp.MustCompile(`(?is)\b(DROP\s+TABLE|DROP\s+COLUMN|RENAME\s+COLUMN|RENAME\s+TO|DELETE\s+FROM|UPDATE\s+[A-Za-z_][A-Za-z0-9_]*\s+SET)\b`)
	allowDestructiveRe = regexp.MustCompile(`(?i)--\s*migration:allow-destructive\b`)
)

// MigrationFile describes one embedded .sql migration file.
type MigrationFile struct {
	Name    string
	Version uint
	Kind    string // "up" or "down"
	Content []byte
}

// ListMigrationFiles returns every embedded migration file in a stable order.
func ListMigrationFiles() ([]MigrationFile, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	files := make([]MigrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		match := migrationFileRe.FindStringSubmatch(name)
		if match == nil {
			return nil, fmt.Errorf("invalid migration filename %q (expected NNNNNN_name.up.sql / .down.sql)", name)
		}
		version, err := strconv.ParseUint(match[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", name, err)
		}
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}
		files = append(files, MigrationFile{
			Name:    name,
			Version: uint(version),
			Kind:    match[3],
			Content: content,
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return files, nil
}

// LatestMigrationVersion returns the highest embedded migration version.
func LatestMigrationVersion() (uint, error) {
	files, err := ListMigrationFiles()
	if err != nil {
		return 0, err
	}
	var latest uint
	for _, file := range files {
		if file.Kind == "up" && file.Version > latest {
			latest = file.Version
		}
	}
	return latest, nil
}

// MigrationChecksums maps every embedded migration filename to its SHA256.
func MigrationChecksums() (map[string]string, error) {
	files, err := ListMigrationFiles()
	if err != nil {
		return nil, err
	}
	sums := make(map[string]string, len(files))
	for _, file := range files {
		sum := sha256.Sum256(file.Content)
		sums[file.Name] = hex.EncodeToString(sum[:])
	}
	return sums, nil
}

// LoadMigrationChecksums parses the embedded checksum manifest. The format is
// one "<sha256>  <filename>" line per file (sha256sum-compatible).
func LoadMigrationChecksums() (map[string]string, error) {
	raw, err := migrationsLockFS.ReadFile(migrationsLockName)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", migrationsLockName, err)
	}
	sums := make(map[string]string)
	for lineNo, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("%s:%d: expected '<sha256>  <filename>'", migrationsLockName, lineNo+1)
		}
		sums[fields[1]] = fields[0]
	}
	return sums, nil
}

// VerifyMigrationChecksums compares the embedded migrations with the reviewed
// manifest. It returns an error if a historical file was modified, removed or
// added without regenerating the manifest.
func VerifyMigrationChecksums() error {
	manifest, err := LoadMigrationChecksums()
	if err != nil {
		return err
	}
	current, err := MigrationChecksums()
	if err != nil {
		return err
	}

	for name, want := range manifest {
		got, ok := current[name]
		if !ok {
			return fmt.Errorf("migration %q is in %s but missing from the embedded files (was it deleted?)", name, migrationsLockName)
		}
		if got != want {
			return fmt.Errorf("migration %q changed after review (manifest %s, current %s)", name, want, got)
		}
	}
	for name := range current {
		if _, ok := manifest[name]; !ok {
			return fmt.Errorf("migration %q is not in %s; run 'make migrations-checksum' and commit the manifest", name, migrationsLockName)
		}
	}
	return nil
}

// CheckMigrations validates structure (up/down pairs, sequential versions) and
// lints new migrations for destructive statements that would make upgrades
// unsafe. Historical migrations are exempt from the lint but still protected by
// the checksum manifest.
func CheckMigrations() []error {
	files, err := ListMigrationFiles()
	if err != nil {
		return []error{err}
	}

	var problems []error
	ups := map[uint]MigrationFile{}
	downs := map[uint]MigrationFile{}
	for _, file := range files {
		switch file.Kind {
		case "up":
			ups[file.Version] = file
		case "down":
			downs[file.Version] = file
		}
	}

	versions := make([]int, 0, len(ups))
	for version := range ups {
		versions = append(versions, int(version))
	}
	sort.Ints(versions)
	for i, version := range versions {
		want := i + 1
		if version != want {
			problems = append(problems, fmt.Errorf("migration versions are not sequential: expected %06d, found %06d", want, version))
			break
		}
		if _, ok := downs[uint(version)]; !ok {
			problems = append(problems, fmt.Errorf("migration %06d is missing its .down.sql file", version))
		}
		if uint(version) >= migrationLintFromVersion {
			up := ups[uint(version)]
			if problem := lintMigration(up); problem != nil {
				problems = append(problems, problem)
			}
		}
	}
	for version := range downs {
		if _, ok := ups[version]; !ok {
			problems = append(problems, fmt.Errorf("migration %06d has a .down.sql file but no .up.sql", version))
		}
	}
	return problems
}

func lintMigration(file MigrationFile) error {
	sqlText := string(file.Content)
	if !destructiveSQLRe.MatchString(sqlText) || allowDestructiveRe.MatchString(sqlText) {
		return nil
	}
	return fmt.Errorf("migration %q contains a destructive statement (DROP/RENAME/DELETE/UPDATE); use expand/contract or add '-- migration:allow-destructive reason=...' with justification", file.Name)
}
