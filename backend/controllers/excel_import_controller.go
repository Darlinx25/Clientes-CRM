package controllers

import (
	"fmt"
	apperrors "meerkat/errors"
	"meerkat/logger"
	"meerkat/models"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const (
	excelColumnCliente    = "Client"
	excelColumnTipo       = "CompanyType"
	excelColumnPersona    = "ContactPerson"
	excelColumnEmail      = "Email"
	excelColumnCelular    = "Phone"
	excelColumnRUT        = "Rut"
	excelColumnNumEmpresa = "CompanyNumber"
)

// ExcelImportResult summarizes the outcome of a bulk Excel import
type ExcelImportResult struct {
	ContactsCreated int      `json:"contacts_created"`
	ContactsUpdated int      `json:"contacts_updated"`
	CompaniesAdded  int      `json:"companies_added"`
	RowsProcessed   int      `json:"rows_processed"`
	RowsSkipped     int      `json:"rows_skipped"`
	Errors          []string `json:"errors"`
}

// excelRow holds the parsed values of a single spreadsheet row
type excelRow struct {
	lineNumber      int
	cliente         string
	tipoEmpresa     string
	personaContacto string
	email           string
	celular         string
	rut             string
	numeroEmpresa   string
}

type excelHeaders struct {
	cliente    int
	tipo       int
	persona    int
	email      int
	celular    int
	rut        int
	numEmpresa int
	found      map[string]bool
}

// ImportExcelContacts processes an uploaded Excel file (.xlsx) for bulk client import.
// Columns are matched by their header name (case-insensitive):
//
//	Cliente, Tipo de Empresa, Persona de Contacto, Email, Celular, RUT, Número de Empresa
//
// Clients are deduplicated by RUT: multiple rows sharing the same RUT are merged into a
// single client, a Número de Empresa per row is attached, and every Tipo de Empresa of
// that row is linked to it.
func ImportExcelContacts(c *gin.Context) {
	log := logger.FromContext(c)

	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	file, err := c.FormFile("file")
	if err != nil {
		log.Warn().Err(err).Msg("No file uploaded")
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("file", "No file uploaded"))
		return
	}

	if file.Size > 20*1024*1024 {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("file", "File too large. Maximum size is 20 MB"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".xlsx" && ext != ".xlsm" {
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("file", "File must be an .xlsx spreadsheet"))
		return
	}

	f, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open uploaded file")
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to process file"))
		return
	}
	defer f.Close()

	tf, err := os.CreateTemp("", "import-*.xlsx")
	if err != nil {
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to process file"))
		return
	}
	path := tf.Name()
	defer tf.Close()
	defer os.Remove(path)

	if _, err := tf.ReadFrom(f); err != nil {
		log.Error().Err(err).Msg("Failed to buffer uploaded file")
		apperrors.AbortWithError(c, apperrors.ErrInternal("Failed to process file"))
		return
	}

	result, err := processExcelImport(db, userID, path)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to process Excel import")
		apperrors.AbortWithError(c, apperrors.ErrInvalidInput("file", err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}

// processExcelImport parses the spreadsheet, groups rows by RUT and upserts clients + companies.
func processExcelImport(db *gorm.DB, userID uint, path string) (*ExcelImportResult, error) {
	result := &ExcelImportResult{}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el archivo Excel: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("el archivo Excel no contiene ninguna hoja")
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la hoja: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("el archivo no tiene filas de datos (solo la fila de encabezados)")
	}

	headers, err := parseExcelHeaders(rows[0])
	if err != nil {
		return nil, err
	}

	// Fixed list of company types, resolved by name (case-insensitive)
	var companyTypes []models.CompanyType
	if err := db.Order("name ASC").Find(&companyTypes).Error; err != nil {
		return nil, fmt.Errorf("no se pudo cargar los tipos de empresa: %w", err)
	}
	typeByName := make(map[string]models.CompanyType, len(companyTypes))
	for _, t := range companyTypes {
		typeByName[strings.ToLower(strings.TrimSpace(t.Name))] = t
	}

	// Map normalized RUT -> accumulated client data
	type clientAccumulator struct {
		cliente         string
		personaContacto string
		email           string
		celular         string
		companies       map[string]map[string]bool // company number -> set of type names
	}
	clients := make(map[string]*clientAccumulator)
	order := make([]string, 0)

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isExcelRowEmpty(row) {
			continue
		}
		result.RowsProcessed++

		parsed := parseExcelRow(row, headers, i+1)
		if parsed.cliente == "" {
			result.RowsSkipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Fila %d: falta el campo Cliente", parsed.lineNumber))
			continue
		}

		key := normalizeRUT(parsed.rut)
		if key == "" {
			// No RUT: dedup by client name (case-insensitive)
			key = "name:" + strings.ToLower(strings.TrimSpace(parsed.cliente))
		}

		acc, exists := clients[key]
		if !exists {
			acc = &clientAccumulator{companies: make(map[string]map[string]bool)}
			clients[key] = acc
			order = append(order, key)
		}

		if acc.cliente == "" {
			acc.cliente = parsed.cliente
		}
		if parsed.personaContacto != "" {
			acc.personaContacto = parsed.personaContacto
		}
		if parsed.email != "" {
			acc.email = parsed.email
		}
		if parsed.celular != "" {
			acc.celular = parsed.celular
		}

		if parsed.numeroEmpresa != "" {
			tipos := acc.companies[parsed.numeroEmpresa]
			if tipos == nil {
				tipos = make(map[string]bool)
				acc.companies[parsed.numeroEmpresa] = tipos
			}
			if parsed.tipoEmpresa != "" {
				tipos[parsed.tipoEmpresa] = true
			}
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, key := range order {
			acc := clients[key]
			normRUT := ""
			if !strings.HasPrefix(key, "name:") {
				normRUT = key
			}

			var contact models.Contact
			query := tx
			if normRUT != "" {
				if err := query.Where("rut = ?", normRUT).First(&contact).Error; err != nil {
					if err != gorm.ErrRecordNotFound {
						return err
					}
				}
			}

			isNew := contact.ID == 0
			if isNew && normRUT == "" {
				// Match existing contact by same name when no RUT is present
				nameLike := strings.ToLower(strings.TrimSpace(acc.cliente))
				var byName models.Contact
				if err := tx.Where("LOWER(firstname) = ?", nameLike).First(&byName).Error; err == nil {
					contact = byName
					isNew = false
				}
			}

			if isNew {
				contact = models.Contact{
					UserID:        userID,
					Firstname:     acc.cliente,
					Rut:           normRUT,
					ContactPerson: acc.personaContacto,
				}
				if acc.email != "" {
					contact.Emails = []models.ContactEmail{{Type: "home", Value: acc.email}}
				}
				if acc.celular != "" {
					contact.Phones = []models.ContactPhone{{Type: "cell", Value: acc.celular}}
				}
				if err := tx.Create(&contact).Error; err != nil {
					return err
				}
				result.ContactsCreated++
				for num, tipos := range acc.companies {
					comp := models.Company{ContactID: contact.ID, CompanyNumber: num}
					if err := tx.Create(&comp).Error; err != nil {
						return err
					}
					tlist := resolveTypeEntities(typeByName, tipos)
					if len(tlist) > 0 {
						if err := tx.Model(&comp).Association("Types").Replace(tlist); err != nil {
							return err
						}
					}
					result.CompaniesAdded++
				}
			} else {
				changed := false
				if acc.personaContacto != "" && contact.ContactPerson != acc.personaContacto {
					contact.ContactPerson = acc.personaContacto
					changed = true
				}
				if acc.email != "" && contact.Email != acc.email {
					contact.Emails = []models.ContactEmail{{Type: "home", Value: acc.email}}
					changed = true
				}
				if acc.celular != "" && contact.Phone != acc.celular {
					contact.Phones = []models.ContactPhone{{Type: "cell", Value: acc.celular}}
					changed = true
				}
				if changed {
					if err := tx.Save(&contact).Error; err != nil {
						return err
					}
				}
				result.ContactsUpdated++

				// Attach company numbers (merging types into existing numbers)
				var existing []models.Company
				if err := tx.Where("contact_id = ?", contact.ID).Find(&existing).Error; err != nil {
					return err
				}
				existingNums := make(map[string]*models.Company, len(existing))
				for i := range existing {
					existingNums[strings.ToLower(strings.TrimSpace(existing[i].CompanyNumber))] = &existing[i]
				}
				for num, tipos := range acc.companies {
					comp, exists := existingNums[strings.ToLower(strings.TrimSpace(num))]
					if !exists {
						comp = &models.Company{ContactID: contact.ID, CompanyNumber: num}
						if err := tx.Create(comp).Error; err != nil {
							return err
						}
						result.CompaniesAdded++
					}
					tlist := resolveTypeEntities(typeByName, tipos)
					if len(tlist) > 0 {
						if err := mergeCompanyTypes(tx, comp, tlist); err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al guardar los clientes: %w", err)
	}

	return result, nil
}

// resolveTypeEntities maps a set of type names to the matching CompanyType rows from the fixed list.
func resolveTypeEntities(typeByName map[string]models.CompanyType, tipos map[string]bool) []models.CompanyType {
	list := make([]models.CompanyType, 0, len(tipos))
	for nome := range tipos {
		if t, ok := typeByName[strings.ToLower(strings.TrimSpace(nome))]; ok {
			list = append(list, t)
		}
	}
	return list
}

// mergeCompanyTypes unions tlist into the company's existing types.
func mergeCompanyTypes(tx *gorm.DB, comp *models.Company, tlist []models.CompanyType) error {
	var cur models.Company
	if err := tx.Preload("Types").First(&cur, comp.ID).Error; err != nil {
		return err
	}
	have := make(map[string]bool, len(cur.Types))
	for _, t := range cur.Types {
		have[strings.ToLower(t.Name)] = true
	}
	final := cur.Types
	for _, t := range tlist {
		if !have[strings.ToLower(t.Name)] {
			final = append(final, t)
		}
	}
	if len(final) > 0 {
		return tx.Model(comp).Association("Types").Replace(final)
	}
	return nil
}

func parseExcelHeaders(headerRow []string) (*excelHeaders, error) {
	allowSet := map[string]string{
		"cliente":             excelColumnCliente,
		"tipo de empresa":     excelColumnTipo,
		"tipo empresa":        excelColumnTipo,
		"tipo":                excelColumnTipo,
		"persona de contacto": excelColumnPersona,
		"persona contacto":    excelColumnPersona,
		"contacto":            excelColumnPersona,
		"email":               excelColumnEmail,
		"correo":              excelColumnEmail,
		"e-mail":              excelColumnEmail,
		"mail":                excelColumnEmail,
		"celular":             excelColumnCelular,
		"telefono":            excelColumnCelular,
		"teléfono":            excelColumnCelular,
		"cel":                 excelColumnCelular,
		"rut":                 excelColumnRUT,
		"numero de empresa":   excelColumnNumEmpresa,
		"número de empresa":   excelColumnNumEmpresa,
		"numero empresa":      excelColumnNumEmpresa,
		"número empresa":      excelColumnNumEmpresa,
		"n de empresa":        excelColumnNumEmpresa,
		"número":              excelColumnNumEmpresa,
		"numero":              excelColumnNumEmpresa,
	}
	h := &excelHeaders{cliente: -1, tipo: -1, persona: -1, email: -1, celular: -1, rut: -1, numEmpresa: -1, found: map[string]bool{}}

	for idx, cell := range headerRow {
		normalized := normalizeHeader(cell)
		if normalized == "" {
			continue
		}
		if target, ok := allowSet[normalized]; ok {
			switch target {
			case excelColumnCliente:
				if h.cliente == -1 {
					h.cliente = idx
				}
			case excelColumnTipo:
				if h.tipo == -1 {
					h.tipo = idx
				}
			case excelColumnPersona:
				if h.persona == -1 {
					h.persona = idx
				}
			case excelColumnEmail:
				if h.email == -1 {
					h.email = idx
				}
			case excelColumnCelular:
				if h.celular == -1 {
					h.celular = idx
				}
			case excelColumnRUT:
				if h.rut == -1 {
					h.rut = idx
				}
			case excelColumnNumEmpresa:
				if h.numEmpresa == -1 {
					h.numEmpresa = idx
				}
			}
			h.found[target] = true
		}
	}

	if h.cliente == -1 {
		return nil, fmt.Errorf("no se encontró la columna requerida \"Cliente\" en la primera fila")
	}
	missing := []string{}
	definitions := map[string]string{
		excelColumnTipo:       "Tipo de Empresa",
		excelColumnPersona:    "Persona de Contacto",
		excelColumnEmail:      "Email",
		excelColumnCelular:    "Celular",
		excelColumnRUT:        "RUT",
		excelColumnNumEmpresa: "Número de Empresa",
	}
	for target, label := range definitions {
		if !h.found[target] {
			missing = append(missing, label)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("faltan columnas en el encabezado: %s", strings.Join(missing, ", "))
	}

	return h, nil
}

func parseExcelRow(row []string, h *excelHeaders, lineNumber int) *excelRow {
	cell := func(idx int) string {
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}
	return &excelRow{
		lineNumber:      lineNumber,
		cliente:         cell(h.cliente),
		tipoEmpresa:     cell(h.tipo),
		personaContacto: cell(h.persona),
		email:           cell(h.email),
		celular:         cell(h.celular),
		rut:             cell(h.rut),
		numeroEmpresa:   cell(h.numEmpresa),
	}
}

func isExcelRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-zñáéíóú0-9]`)
	s = re.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func normalizeRUT(rut string) string {
	rut = strings.TrimSpace(rut)
	if rut == "" {
		return ""
	}
	re := regexp.MustCompile(`[^0-9kK]`)
	return re.ReplaceAllString(rut, "")
}
