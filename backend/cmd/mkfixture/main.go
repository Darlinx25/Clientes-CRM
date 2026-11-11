// Command mkfixture builds a representative SQLite database at a specific
// migration version, used by the upgrade-rehearsal tests.
//
// Usage (from backend/):
//
//	go run ./cmd/mkfixture -version 23 -out database/testdata/fixtures/v000023.db
//	go run ./cmd/mkfixture -version 39 -out database/testdata/fixtures/v000039.db
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"meerkat/database"
)

func main() {
	version := flag.Uint("version", 0, "target migration version")
	out := flag.String("out", "", "output database path")
	force := flag.Bool("force", false, "overwrite an existing fixture")
	flag.Parse()

	if *version == 0 || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: mkfixture -version N -out path.db [-force]")
		os.Exit(2)
	}
	if !*force {
		if _, err := os.Stat(*out); err == nil {
			fatal(fmt.Errorf("%s already exists (use -force to overwrite)", *out))
		}
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	_ = os.Remove(*out)

	if err := database.MigrateUpTo(*out, *version); err != nil {
		fatal(err)
	}
	if err := seed(*out); err != nil {
		fatal(err)
	}
	fmt.Printf("fixture %s created at migration version %d\n", *out, *version)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "mkfixture: %v\n", err)
	os.Exit(1)
}

func seed(path string) error {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(ON)")
	if err != nil {
		return err
	}
	defer db.Close()

	cache := &columnCache{db: db, columns: map[string]map[string]bool{}}
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	if err := insert(db, cache, "users", map[string]any{
		"id": 1, "created_at": now, "updated_at": now,
		"username": "admin", "password": "fixture-password-hash",
		"email": "admin@fixture.invalid", "language": "es", "is_admin": 1,
		"custom_field_names": `["Teléfono fijo"]`,
	}); err != nil {
		return err
	}

	contacts := []map[string]any{
		{
			"id": 1, "created_at": now, "updated_at": now, "user_id": 1,
			"firstname": "María", "lastname": "González", "nickname": "Mary",
			"email": "maria@example.invalid", "phone": "+598 99 111 222",
			"birthday": "1985-03-14", "address": "Av. Siempre Viva 742",
			"how_we_met": "Cliente histórico", "circles": `["Clientes"]`,
			"custom_fields": `{"Teléfono fijo":"2400 1234"}`,
			"emails":        `[{"type":"work","value":"maria@example.invalid"}]`,
			"phones":        `[{"type":"cell","value":"+598 99 111 222"}]`,
			"prefix":        "", "middle_name": "", "suffix": "",
			"organization": "González y Cía.", "department": "Administración",
			"job_title": "Directora", "role": "", "anniversary": "2010-11-02",
			"archived": 0,
		},
		{
			"id": 2, "created_at": now, "updated_at": now, "user_id": 1,
			"firstname": "José", "lastname": "Pérez", "email": "jose@example.invalid",
			"phone": "+598 98 333 444", "circles": `["Clientes"]`, "custom_fields": `{}`,
			"emails": `[]`, "phones": `[]`, "archived": 1,
		},
		{
			"id": 3, "created_at": now, "updated_at": now, "user_id": 1,
			"firstname": "Ana", "lastname": "Silva", "email": "ana@example.invalid",
			"circles": `[]`, "custom_fields": `{}`, "emails": `[]`, "phones": `[]`, "archived": 0,
		},
	}
	for _, contact := range contacts {
		if err := insert(db, cache, "contacts", contact); err != nil {
			return err
		}
	}

	notes := []map[string]any{
		{"id": 1, "created_at": now, "updated_at": now, "user_id": 1, "contact_id": 1, "title": "Llamada inicial", "content": "Se conversó sobre la declaración jurada.", "date": "2024-01-10 10:00:00"},
		{"id": 2, "created_at": now, "updated_at": now, "user_id": 1, "contact_id": 1, "title": "", "content": "Documentación recibida.", "date": "2024-02-15 12:30:00"},
		{"id": 3, "created_at": now, "updated_at": now, "user_id": 1, "contact_id": 2, "title": "Nota antigua", "content": "Sin tilde para probar compatibilidad.", "date": "2023-12-01 09:15:00"},
	}
	for _, note := range notes {
		if err := insert(db, cache, "notes", note); err != nil {
			return err
		}
	}

	if err := insert(db, cache, "reminders", map[string]any{
		"id": 1, "created_at": now, "updated_at": now, "user_id": 1, "contact_id": 1,
		"message": "Revisar balance", "remind_at": "2026-12-01 09:00:00",
		"recurrence": "yearly", "by_mail": 0,
	}); err != nil {
		return err
	}

	if err := insert(db, cache, "activities", map[string]any{
		"id": 1, "created_at": now, "updated_at": now, "user_id": 1,
		"title": "Reunión", "description": "Reunión de seguimiento", "date": "2024-03-01 15:00:00",
	}); err != nil {
		return err
	}

	if err := insert(db, cache, "relationships", map[string]any{
		"id": 1, "created_at": now, "updated_at": now, "user_id": 1,
		"name": "Hermana", "type": "Sister", "contact_id": 1, "related_contact_id": 3,
	}); err != nil {
		return err
	}

	if hasTable(db, "companies") {
		for id, number := range []string{"214567890012", "214567890013"} {
			if err := insert(db, cache, "companies", map[string]any{
				"id": id + 1, "created_at": now, "updated_at": now,
				"contact_id": 1, "company_number": number, "aportacion": "BPS",
			}); err != nil {
				return err
			}
		}
		if hasTable(db, "company_company_types") && hasTable(db, "company_types") {
			rows, err := db.Query("SELECT id FROM company_types ORDER BY id LIMIT 2")
			if err != nil {
				return err
			}
			defer rows.Close()
			var typeIDs []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					return err
				}
				typeIDs = append(typeIDs, id)
			}
			for companyID := 1; companyID <= 2 && len(typeIDs) > 0; companyID++ {
				if _, err := db.Exec("INSERT OR IGNORE INTO company_company_types (company_id, company_type_id) VALUES (?, ?)", companyID, typeIDs[(companyID-1)%len(typeIDs)]); err != nil {
					return err
				}
			}
		}
	}

	if hasTable(db, "app_settings") {
		for key, value := range map[string]string{
			"auto_backup_enabled": "0", "auto_backup_weekday": "friday",
			"auto_backup_time": "18:00", "auto_backup_timezone": "America/Montevideo",
		} {
			if err := insert(db, cache, "app_settings", map[string]any{"key": key, "value": value}); err != nil {
				return err
			}
		}
	}

	return nil
}

type columnCache struct {
	db      *sql.DB
	columns map[string]map[string]bool
}

func (c *columnCache) has(table, column string) bool {
	cols, ok := c.columns[table]
	if !ok {
		cols = map[string]bool{}
		rows, err := c.db.Query("PRAGMA table_info(" + table + ")")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cid int
				var name, typ string
				var notNull, pk int
				var dflt any
				if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err == nil {
					cols[name] = true
				}
			}
		}
		c.columns[table] = cols
	}
	return cols[column]
}

func insert(db *sql.DB, cache *columnCache, table string, values map[string]any) error {
	if !hasTable(db, table) {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if cache.has(table, key) {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = values[key]
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(keys, ", "), strings.Join(placeholders, ", "))
	if _, err := db.Exec(query, args...); err != nil {
		return fmt.Errorf("insert into %s: %w", table, err)
	}
	return nil
}

func hasTable(db *sql.DB, table string) bool {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count); err != nil {
		return false
	}
	return count > 0
}
