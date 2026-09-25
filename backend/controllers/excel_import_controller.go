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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const (
	excelColumnCliente     = "Client"
	excelColumnTipo        = "CompanyType"
	excelColumnPersona     = "ContactPerson"
	excelColumnEmail       = "Email"
	excelColumnCelular     = "Phone"
excelColumnRUT         = "Rut"
excelColumnDocumento = "Documento"
excelColumnNumEmpresa  = "CompanyNumber"
	excelColumnAniversario = "Anniversary"
	excelColumnComentario  = "ContactInformation"
	excelColumnAportacion  = "Aportacion"
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
	documento       string
	numeroEmpresa   string
	aniversario     string
	comentario      string
	aportacion      string
}

type excelHeaders struct {
	cliente     int
	tipo        int
	persona     int
	email       int
	celular     int
	rut         int
	documento   int
	numEmpresa  int
	aniversario int
	comentario  int
	aportacion  int
	found       map[string]bool
}

// ImportExcelContacts processes an uploaded Excel file (.xlsx) for bulk client import.
// Columns are matched by their header name (case-insensitive):
//
//	Cliente, Tipo de Empresa, Persona de Contacto, Email, Celular, RUT, Número de Empresa
//
// Aniversario, Comentario and Aportación are optional extra columns: when present
// in the header they are imported per client (aniversario/comentario) or per
// company (aportación). Email and Celular may hold several values separated by
// comma, semicolon or pipe; each becomes a separate email/phone entry.
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
	type companyAccumulator struct {
		tipos      map[string]bool
		aportacion string
	}
	type clientAccumulator struct {
		cliente         string
		documento     string
		personaContacto string
		emails          []string
		phones          []string
		aniversario     string
		comentario      string
		companies       map[string]*companyAccumulator // company number -> aportacion + set of type names
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
			acc = &clientAccumulator{companies: make(map[string]*companyAccumulator)}
			clients[key] = acc
			order = append(order, key)
		}

		if acc.cliente == "" {
			acc.cliente = parsed.cliente
		}
		if parsed.personaContacto != "" {
			acc.personaContacto = parsed.personaContacto
		}
		if parsed.documento != "" && acc.documento == "" {
			acc.documento = parsed.documento
		}
		if parsed.email != "" {
			acc.emails = appendUnique(acc.emails, splitMultiValues(parsed.email))
		}
		if parsed.celular != "" {
			acc.phones = appendUnique(acc.phones, splitMultiValues(parsed.celular))
		}
		if n := normalizeExcelDate(parsed.aniversario); n != "" && acc.aniversario == "" {
			acc.aniversario = n
		}
		if parsed.comentario != "" && acc.comentario == "" {
			acc.comentario = trimMax(parsed.comentario, 1000)
		}

		if parsed.numeroEmpresa != "" {
			ca, ok := acc.companies[parsed.numeroEmpresa]
			if !ok {
				ca = &companyAccumulator{tipos: make(map[string]bool)}
				acc.companies[parsed.numeroEmpresa] = ca
			}
			if parsed.tipoEmpresa != "" {
				ca.tipos[parsed.tipoEmpresa] = true
			}
			if parsed.aportacion != "" && ca.aportacion == "" {
				ca.aportacion = trimMax(parsed.aportacion, 100)
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
					UserID:             userID,
					Firstname:          acc.cliente,
					Rut:                normRUT,
					Documento:           acc.documento,
					ContactPerson:      acc.personaContacto,
					Anniversary:        acc.aniversario,
					ContactInformation: acc.comentario,
				}
				if len(acc.emails) > 0 {
					contact.Emails = make([]models.ContactEmail, 0, len(acc.emails))
					for _, v := range acc.emails {
						contact.Emails = append(contact.Emails, models.ContactEmail{Type: "home", Value: v})
					}
				}
				if len(acc.phones) > 0 {
					contact.Phones = make([]models.ContactPhone, 0, len(acc.phones))
					for _, v := range acc.phones {
						contact.Phones = append(contact.Phones, models.ContactPhone{Type: "cell", Value: v})
					}
				}
				if err := tx.Create(&contact).Error; err != nil {
					return err
				}
				result.ContactsCreated++
				for num, ca := range acc.companies {
					comp := models.Company{ContactID: contact.ID, CompanyNumber: num, Aportacion: ca.aportacion}
					if err := tx.Create(&comp).Error; err != nil {
						return err
					}
					tlist := resolveTypeEntities(typeByName, ca.tipos)
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
				if acc.documento != "" && contact.Documento != acc.documento {
					contact.Documento = acc.documento
					changed = true
				}
				if acc.aniversario != "" && contact.Anniversary != acc.aniversario {
					contact.Anniversary = acc.aniversario
					changed = true
				}
				if acc.comentario != "" && contact.ContactInformation != acc.comentario {
					contact.ContactInformation = acc.comentario
					changed = true
				}
				if len(acc.emails) > 0 && contact.Email != acc.emails[0] {
					contact.Emails = make([]models.ContactEmail, 0, len(acc.emails))
					for _, v := range acc.emails {
						contact.Emails = append(contact.Emails, models.ContactEmail{Type: "home", Value: v})
					}
					changed = true
				}
				if len(acc.phones) > 0 && contact.Phone != acc.phones[0] {
					contact.Phones = make([]models.ContactPhone, 0, len(acc.phones))
					for _, v := range acc.phones {
						contact.Phones = append(contact.Phones, models.ContactPhone{Type: "cell", Value: v})
					}
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
				for num, ca := range acc.companies {
					normNum := strings.ToLower(strings.TrimSpace(num))
					comp, exists := existingNums[normNum]
					if !exists {
						comp = &models.Company{ContactID: contact.ID, CompanyNumber: num, Aportacion: ca.aportacion}
						if err := tx.Create(comp).Error; err != nil {
							return err
						}
						result.CompaniesAdded++
					} else if ca.aportacion != "" && strings.ToLower(strings.TrimSpace(comp.Aportacion)) != strings.ToLower(strings.TrimSpace(ca.aportacion)) {
						comp.Aportacion = ca.aportacion
						if err := tx.Save(comp).Error; err != nil {
							return err
						}
					}
					tlist := resolveTypeEntities(typeByName, ca.tipos)
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
		"documento":          excelColumnDocumento,
		"ci":                 excelColumnDocumento,
		"cédula":             excelColumnDocumento,
		"cedula":             excelColumnDocumento,
		"cedula de identidad": excelColumnDocumento,
		"numero de empresa":   excelColumnNumEmpresa,
		"número de empresa":   excelColumnNumEmpresa,
		"numero empresa":      excelColumnNumEmpresa,
		"número empresa":      excelColumnNumEmpresa,
		"n de empresa":        excelColumnNumEmpresa,
		"número":              excelColumnNumEmpresa,
		"numero":              excelColumnNumEmpresa,
		"aniversario":         excelColumnAniversario,
		"fecha de inicio":     excelColumnAniversario,
		"inicio":              excelColumnAniversario,
		"comentario":          excelColumnComentario,
		"comentarios":         excelColumnComentario,
		"notas":               excelColumnComentario,
		"observaciones":       excelColumnComentario,
		"aportacion":          excelColumnAportacion,
		"aportación":          excelColumnAportacion,
	}
	h := &excelHeaders{cliente: -1, tipo: -1, persona: -1, email: -1, celular: -1, rut: -1, documento: -1, numEmpresa: -1, aniversario: -1, comentario: -1, aportacion: -1, found: map[string]bool{}}

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
			case excelColumnDocumento:
				if h.documento == -1 {
					h.documento = idx
				}
			case excelColumnNumEmpresa:
				if h.numEmpresa == -1 {
					h.numEmpresa = idx
				}
			case excelColumnAniversario:
				if h.aniversario == -1 {
					h.aniversario = idx
				}
			case excelColumnComentario:
				if h.comentario == -1 {
					h.comentario = idx
				}
			case excelColumnAportacion:
				if h.aportacion == -1 {
					h.aportacion = idx
				}
			}
			h.found[target] = true
		}
	}

	if h.cliente == -1 {
		return nil, fmt.Errorf("no se encontró la columna requerida \"Cliente\" en la primera fila")
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
		documento:       cell(h.documento),
		numeroEmpresa:   cell(h.numEmpresa),
		aniversario:     cell(h.aniversario),
		comentario:      cell(h.comentario),
		aportacion:      cell(h.aportacion),
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

// splitMultiValues splits a cell holding several values ("a@x.cl; b@y.cl")
// into trimmed, de-duplicated pieces. White space, comma, semicolon and pipe
// are accepted as separators so pasted lists keep working.
func splitMultiValues(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == '|' || r == ',' || r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// appendUnique appends vals to dst, keeping existing entries (order preserved).
func appendUnique(dst []string, vals []string) []string {
	seen := make(map[string]bool, len(dst)+len(vals))
	for _, v := range dst {
		seen[v] = true
	}
	for _, v := range vals {
		if seen[v] {
			continue
		}
		seen[v] = true
		dst = append(dst, v)
	}
	return dst
}

// normalizeExcelDate converts the common date formats found in spreadsheets to
// the yyyy-mm-dd string the app stores. Unparseable input returns "".
func normalizeExcelDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Strict year-first formats first, then day-first.
	layouts := []string{
		"2006-01-02",
		"2006-1-2",
		"2006/01/02",
		"2006.01.02",
		"02/01/2006",
		"2/1/2006",
		"02.01.2006",
		"2.1.2006",
		"02-01-2006",
		"2-1-2006",
	}
	for _, l := range layouts {
		if ts, err := time.Parse(l, s); err == nil {
			return ts.Format("2006-01-02")
		}
	}
	return ""
}

// trimMax trims whitespace and caps the value at max runes (avoids cutting a
// multi-byte character in half).
func trimMax(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}
