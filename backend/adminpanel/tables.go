package adminpanel

import (
	"errors"
	"fmt"
	"meerkat/models"
	"strings"

	"gorm.io/gorm"
)

var ErrUnknownTable = errors.New("unknown admin panel table")

type columnDef struct {
	Name       string
	Searchable bool
}

type tableDef struct {
	Name    string
	Label   string
	Model   any
	Order   string
	Columns []columnDef
}

var tableDefs = []tableDef{
	{
		Name:  "contacts",
		Label: "Clientes",
		Model: &models.Contact{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "firstname", Searchable: true},
			{Name: "lastname", Searchable: true},
			{Name: "email", Searchable: true},
			{Name: "phone", Searchable: true},
			{Name: "rut", Searchable: true},
			{Name: "documento", Searchable: true},
			{Name: "archived"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "notes",
		Label: "Notas",
		Model: &models.Note{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "title", Searchable: true},
			{Name: "content", Searchable: true},
			{Name: "contact_id"},
			{Name: "company_id"},
			{Name: "date"},
			{Name: "author_id"},
			{Name: "edited_by_id"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "activities",
		Label: "Actividades",
		Model: &models.Activity{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "title", Searchable: true},
			{Name: "description", Searchable: true},
			{Name: "location", Searchable: true},
			{Name: "date"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "reminders",
		Label: "Recordatorios",
		Model: &models.Reminder{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "message", Searchable: true},
			{Name: "remind_at"},
			{Name: "recurrence"},
			{Name: "completed"},
			{Name: "contact_id"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "reminder_completions",
		Label: "Recordatorios completados",
		Model: &models.ReminderCompletion{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "reminder_id"},
			{Name: "contact_id"},
			{Name: "message", Searchable: true},
			{Name: "completed_at"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "companies",
		Label: "Empresas",
		Model: &models.Company{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "contact_id"},
			{Name: "company_number", Searchable: true},
			{Name: "aportacion", Searchable: true},
			{Name: "custom_type", Searchable: true},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "relationships",
		Label: "Relaciones",
		Model: &models.Relationship{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "contact_id"},
			{Name: "related_contact_id"},
			{Name: "name", Searchable: true},
			{Name: "type", Searchable: true},
			{Name: "gender"},
			{Name: "birthday"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "users",
		Label: "Usuarios",
		Model: &models.User{},
		Order: "id DESC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "username", Searchable: true},
			{Name: "email", Searchable: true},
			{Name: "is_admin"},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "job_executions",
		Label: "Jobs",
		Model: &models.JobExecution{},
		Order: "job_name ASC",
		Columns: []columnDef{
			{Name: "id"},
			{Name: "job_name", Searchable: true},
			{Name: "last_run_at"},
			{Name: "locked_at"},
			{Name: "locked_by", Searchable: true},
			{Name: "created_at"},
			{Name: "updated_at"},
			{Name: "deleted_at"},
		},
	},
	{
		Name:  "schema_migrations",
		Label: "Migraciones",
		Order: "version DESC",
		Columns: []columnDef{
			{Name: "version"},
			{Name: "dirty"},
		},
	},
}

func ListTables(db *gorm.DB) ([]TableSummary, error) {
	tables := make([]TableSummary, 0, len(tableDefs))
	for _, def := range tableDefs {
		var count int64
		if db.Migrator().HasTable(def.Name) {
			if err := tableQuery(db, def, false).Count(&count).Error; err != nil {
				return nil, err
			}
		}
		tables = append(tables, TableSummary{
			Name:     def.Name,
			Label:    def.Label,
			RowCount: count,
		})
	}
	return tables, nil
}

func QueryTable(db *gorm.DB, name string, page, limit int, search string, includeDeleted bool) (*TableRows, error) {
	def, ok := findTable(name)
	if !ok {
		return nil, ErrUnknownTable
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	query := tableQuery(db, def, includeDeleted)
	query = applySearch(query, def, search)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(def.Columns))
	for _, column := range def.Columns {
		columns = append(columns, column.Name)
	}

	rows := make([]map[string]any, 0)
	err := query.
		Select(strings.Join(columns, ", ")).
		Order(def.Order).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &TableRows{
		Table:      def.Name,
		Label:      def.Label,
		Columns:    columns,
		Rows:       rows,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func findTable(name string) (tableDef, bool) {
	for _, def := range tableDefs {
		if def.Name == name {
			return def, true
		}
	}
	return tableDef{}, false
}

func tableQuery(db *gorm.DB, def tableDef, includeDeleted bool) *gorm.DB {
	if def.Model != nil {
		if includeDeleted {
			return db.Unscoped().Model(def.Model)
		}
		return db.Model(def.Model)
	}
	return db.Table(def.Name)
}

func applySearch(query *gorm.DB, def tableDef, search string) *gorm.DB {
	term := strings.TrimSpace(search)
	if term == "" {
		return query
	}
	runes := []rune(term)
	if len(runes) > 200 {
		term = string(runes[:200])
	}

	clauses := make([]string, 0)
	args := make([]any, 0)
	for _, column := range def.Columns {
		if !column.Searchable {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("%s LIKE ?", column.Name))
		args = append(args, "%"+term+"%")
	}
	if len(clauses) == 0 {
		return query
	}
	return query.Where("("+strings.Join(clauses, " OR ")+")", args...)
}
