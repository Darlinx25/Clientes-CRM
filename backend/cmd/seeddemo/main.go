// Command seeddemo wipes the app database and inserts a realistic demo
// dataset: 100 clients (people + societies), each with 3-12 "story-like"
// notes written by one of four demo users.
//
// It talks directly to the SQLite file with the same pure-Go driver the
// backend uses, so it must be run against a stopped database.
//
//	DB_PATH=./data/meerkat.db go run ./cmd/seeddemo
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/glebarez/sqlite"
)

const adminID = 6

type event struct {
	title string
	body  string
	ago   int // days ago, oldest first
}

type storyline struct {
	name   string
	events []event
}

var storylines = []storyline{
	{"agro", []event{
		{"Primer contacto", "Se comunicó por teléfono para asesoramiento rural. Se abrió el legajo del cliente.", 700},
		{"Apertura de empresa", "Se confeccionó la apertura de la empresa en el padrón @p1 y se presentó toda la documentación.", 690},
		{"Baja de padrones", "Se bajaron 2 padrones del registro: @p2 y @p3. Quedó activo solo el @p1.", 600},
		{"Solicitud de CUD", "Se pidió el CUD por el padrón @p4. Se adjuntaron el plano y la cédula catastral.", 520},
		{"Expediente por rectificativa", "Se generó el expediente @e en UBA por la rectificativa de superficie.", 450},
		{"Cambio de domicilio fiscal", "Se actualizó el domicilio fiscal a @dom. Se informó a UBA y a Rentas.", 300},
		{"Envío por WhatsApp", "El cliente pidió que le enviemos toda la documentación y los comprobantes por WhatsApp.", 280},
		{"Transformación societaria", "Inicia la transformación de @t1 a SA; se está armando el acta de asamblea.", 150},
		{"Vencimientos UBA", "Se planificaron los vencimientos de UBA y del CUD del próximo cuatrimestre.", 60},
		{"Facturación CFE", "Se emitieron las facturas electrónicas del semestre por el servicio de asesoramiento.", 15},
		{"Declaración anual", "Se presentó la declaración jurada anual y se planificaron las anticipaciones.", 6},
		{"Renovación de CUD", "Se renovó el CUD del padrón @p4 con el nuevo catastro.", 2},
	}},
	{"obra", []event{
		{"Apertura de obra", "Se abrió la obra en el padrón @p1 y se presentó el plano de mensura.", 500},
		{"Baja de padrones", "Parte de la obra se baja del padrón @p2 por fraccionamiento.", 400},
		{"Expediente MTOP", "Se generó el expediente @e en el MTOP por la construcción del puente.", 320},
		{"Habilitación", "Se inició el trámite de habilitación de la obra ante la Intendencia.", 200},
		{"Rectificativa catastral", "Se presentó la rectificativa por la superficie construida.", 120},
		{"Solicitud de CUD", "Se solicitó CUD por el padrón @p3.", 80},
		{"Presupuesto de obra", "Se liquidó el presupuesto de obra y el IVA correspondiente.", 30},
	}},
	{"domestico", []event{
		{"Alta de servicio doméstico", "Se registró el trabajo de servicio doméstico y se emitió el título de trabajo.", 400},
		{"Liquidación mensual", "Liquidación del mes: aportes y salario calculados según la tabla vigente.", 30},
		{"Envío de comprobantes", "Se enviaron los recibos y comprobantes por WhatsApp.", 10},
		{"Fin de año", "Se generaron los aguinaldos y los recibos de fin de año.", 0},
	}},
	{"inactivo", []event{
		{"Apertura de empresa", "Se abrió la empresa en el padrón @p1 (tipo @t1).", 800},
		{"Solicitud de CUD", "Se solicitó el CUD por el padrón @p2.", 700},
		{"Expediente por rectificativa", "Se generó el expediente @e por la rectificativa del @p1.", 500},
		{"Baja de padrones", "Se bajaron los padrones @p2 y @p3.", 300},
		{"Inactivo", "El cliente está inactivo: se archiva el legajo hasta nuevo aviso.", 45},
	}},
	{"credito", []event{
		{"Consultas iniciales", "Preliminares por la compra del padrón @p1.", 300},
		{"Certificado catastral", "Se tramitó el certificado catastral del padrón.", 280},
		{"Crédito BROU", "Se armó la carpeta para el crédito y se adjuntó el informe técnico.", 200},
		{"Escritura", "Se confeccionó la escritura de compraventa.", 90},
		{"Cambio de titularidad", "Se actualizó la titularidad en UBA y Rentas.", 60},
	}},
	{"diversificacion", []event{
		{"Gestión ante UBA", "Se actualizó y volvió a presentar el legajo UBA del padrón @p1.", 600},
		{"Baja de padrones", "Se bajaron 2 padrones: @p2 y @p3.", 500},
		{"Solicitud de CUD", "Se pidió el CUD por el padrón @p4.", 440},
		{"Expediente por rectificativa", "Expediente @e por la rectificativa del loteo.", 340},
		{"Cambio de domicilio fiscal", "Se cambió el domicilio fiscal a @dom.", 220},
		{"Envío por WhatsApp", "Pidió que le enviemos todo por WhatsApp.", 200},
		{"Declaración jurada", "Se presentó la declaración jurada anual de la empresa.", 100},
		{"Alta de servicio doméstico", "Se dio de alta el servicio doméstico de la casa de la estancia.", 70},
		{"Cambio de tipo societario", "Se realizó el pasaje de @t1 a SA con nueva acta.", 40},
		{"Facturación e IVA", "Se cerró el cuatrimestre con las declaraciones de IVA.", 10},
		{"Servicio doméstico", "Se actualizó el sueldo del servicio doméstico.", 4},
	}},
	{"sociedad", []event{
		{"Constitución de sociedad", "Se constituyó la sociedad @who con tres socios.", 900},
		{"Aporte de padrones", "Se aportaron los padrones @p1 y @p2 a la sociedad.", 850},
		{"Solicitud de CUD", "Se solicitó el CUD por el padrón @p3.", 700},
		{"Expediente por rectificativa", "Se generó el expediente @e por la rectificativa en la bodega.", 500},
		{"Cambio de domicilio fiscal", "Domicilio fiscal actualizado a @dom.", 250},
		{"Vencimientos", "Se anotaron los vencimientos de impuestos del ejercicio.", 90},
		{"Envío por WhatsApp", "Pidió que le enviemos los informes por WhatsApp.", 20},
	}},
	{"campo", []event{
		{"Apertura de empresa", "Apertura de la empresa en el padrón @p1 de secano.", 800},
		{"Baja de padrones", "Se bajaron los padrones @p2 y @p3.", 650},
		{"Solicitud de CUD", "CUD solicitado por el padrón @p4.", 590},
		{"Rectificativa", "Expediente @e por la rectificativa de superficie arrocera.", 430},
		{"Cambio de tipo societario", "Cambio de @t1 a SA.", 220},
		{"Facturación de cosecha", "Se emitieron las facturas electrónicas de la cosecha.", 60},
		{"Próximo ejercicio", "Se preparó el plan del próximo ejercicio: siembra y contratos.", 5},
	}},
	{"turismo", []event{
		{"Apertura de empresa", "Se abrió la empresa de turismo rural en el padrón @p1.", 700},
		{"Solicitud de CUD", "Se solicitó el CUD por el padrón @p2.", 640},
		{"Expediente", "Expediente @e ante el Ministerio de Turismo.", 500},
		{"Cambio de domicilio fiscal", "Cambio del domicilio fiscal a @dom.", 300},
		{"Alta de servicio doméstico", "Alta de servicio doméstico para la estancia.", 150},
		{"Temporada alta", "Se coordinaron las reservas de la temporada y los cobros.", 25},
	}},
	{"exportador", []event{
		{"Apertura de empresa", "Apertura de la empresa exportadora.", 600},
		{"Baja de padrones", "Se bajaron los padrones @p1 y @p2.", 520},
		{"Solicitud de CUD", "CUD por el padrón @p3.", 470},
		{"Expediente por rectificativa", "Expediente @e por la rectificativa.", 380},
		{"Cambio de domicilio fiscal", "Se cambió el domicilio fiscal a @dom.", 250},
		{"Certificados de origen", "Se tramitaron los certificados de origen para la exportación.", 120},
		{"Envío por WhatsApp", "Pidió que le enviemos los documentos por WhatsApp.", 30},
	}},
	{"auditoria", []event{
		{"Revisión anual", "Se hizo la revisión anual de la contabilidad y del libro IVA.", 250},
		{"Expediente", "Expediente @e iniciado por observaciones de Rentas.", 180},
		{"Rectificativa", "Se presentó rectificativa por las diferencias detectadas.", 90},
		{"Domicilio fiscal", "Se regularizó el domicilio fiscal a @dom.", 40},
		{"Cierre de ejercicio", "Se cerró el ejercicio y se planificó la asamblea.", 10},
	}},
}

var firstNames = []string{
	"Rosana", "Carlos", "Marta", "José", "Silvia", "Héctor", "Mónica", "Raúl",
	"Graciela", "Néstor", "Élida", "Washington", "Daniela", "Pedro", "Milka",
	"Alberto", "Lucía", "Enrique", "Beatriz", "Oscar", "Ana", "Fabián",
	"Susana", "Gustavo", "Carina", "Rubén", "Natalia", "Julio", "Verónica",
	"Mario", "Andrea", "Sergio", "Delia", "Walter", "Romina",
}

var lastNames = []string{
	"Pérez", "Rodríguez", "González", "Silveira", "Fernández", "López",
	"Martínez", "Acosta", "Olivera", "Díaz", "Suárez", "Blanco", "Cabrera",
	"Viera", "Román", "Ríos", "Machado", "Beron", "Etchegaray", "Píriz",
	"Ferreira", "Britos", "Albornoz", "Cáceres", "Franco", "Burgueño",
	"Silva", "Mederos", "Borges", "Amaro",
}

var societies = []string{
	"Construcciones SA",
	"Agropecuaria La Cascada SRL",
	"José Luis y Otros",
	"Estancia Los Talas SA",
	"Cooperativa Agrícola del Este",
	"Transporte El Cóndor SA",
	"Estancia San Rafael SRL",
	"Frigorífico Patria",
	"Laboratorio Rural Rondeau",
	"Agroindustria Del Alba",
	"Sociedad Campera Las Margaritas",
	"Polo & Pampa Uruguay SA",
	"Servicios Rurales Rocha",
	"Granja El Llanito",
	"Molino Victoria SA",
	"Construcciones Del Plata SRL",
	"Agromaderas Bella Unión",
	"Sociedad Anónima Campo Nuevo",
	"Tambería Santa Ana",
	"Exportadora Cítricos Norte SA",
	"Horticultura El Jagüel",
	"Camping y Estancia Los Aromos SRL",
}

var streets = []string{
	"18 de Julio 1240",
	"Ruta 8 km 27.5",
	"Sarandí 456",
	"Camino Gral. Flores 890",
	"Bvar. Artigas 1122",
	"Ruta 3 km 214",
	"Av. Italia 3421",
	"Calle Colonia 976",
}

var companyNames = []string{
	"UBA", "Rural", "Bodega", "Estancia", "Obra", "Servicio Doméstico",
	"Casa", "Frigorífico", "Citrícola", "Tambería", "Transporte",
}

var companyTypes = []string{
	"Rural", "SA", "SRL", "SAS", "Monotributo", "Serv Dom", "Obra", "Casa",
	"Administración",
}

var contactPeople = []string{
	"José Luis García", "María Inés Bello", "Sr. Ricardo Etcheverry",
	"Sra. Ana Victoria Crosa", "Ing. Pedro Almeida",
}

var authors = []string{"julia", "rober", "marcela", "andres"}

func fill(s string, p [4]int, exp string, dom string, t1 string) string {
	return strings.NewReplacer(
		"@p1", fmt.Sprintf("%d", p[0]),
		"@p2", fmt.Sprintf("%d", p[1]),
		"@p3", fmt.Sprintf("%d", p[2]),
		"@p4", fmt.Sprintf("%d", p[3]),
		"@e", exp,
		"@dom", dom,
		"@t1", t1,
	).Replace(s)
}

func wipe(db *sql.DB) {
	tables := []string{
		"notes", "activities", "activity_contacts", "relationships",
		"reminders", "reminder_completions", "contacts", "companies",
		"company_company_types", "webhooks", "webhook_deliveries",
		"calendar_subscriptions", "calendar_event_links",
		"carddav_connections", "carddav_contact_links", "carddav_sync",
		"api_tokens", "job_executions",
	}
	if _, err := db.Exec("PRAGMA foreign_keys=OFF"); err != nil {
		log.Fatalf("pragma: %v", err)
	}
	for _, t := range tables {
		if _, err := db.Exec("DELETE FROM " + t); err != nil {
			log.Fatalf("delete %s: %v", t, err)
		}
	}
	if _, err := db.Exec("DELETE FROM users WHERE id != ?", adminID); err != nil {
		log.Fatalf("delete users: %v", err)
	}
	seq := append(tables, "users")
	// sqlite_sequence updates below fail when a table has no AUTOINCREMENT alias
	// or the table is new; ignore those.
	if _, err := db.Exec("DELETE FROM sqlite_sequence"); err == nil {
		reset := make([]string, 0, len(seq)+1)
		for _, t := range seq {
			reset = append(reset, fmt.Sprintf("(%q, 0)", t))
		}
		_, _ = db.Exec("INSERT INTO sqlite_sequence (name, seq) VALUES " + strings.Join(reset, ","))
		_, _ = db.Exec(fmt.Sprintf("UPDATE sqlite_sequence SET seq = COALESCE((SELECT MAX(id) FROM users), 0) WHERE name = 'users'"))
	}
	fmt.Println("- base de datos vaciada (se mantienen admin, company_types y migrations)")
}

func insertUsers(db *sql.DB) map[string]int64 {
	hashes := []string{
		"$2a$10$C3OOYvEb2kItZw/8YzdPkO14MJ5R65iGeChxlb2ztFR.GifKFidzK",
		"$2a$10$rI9QzvrXnDezQiCKZeSuJe6ElXxygXoni.ABMoEDxQ/e1250nWI5W",
		"$2a$10$WeKQSb.3HfB2eUdKdyi.tOIsVBB6bNmT4EzFFIBYPwvebizreELue",
		"$2a$10$VaWZKVav3XcMTy1wgZ..Nepcnc4gu5eBPbr.OB5CWcXIfFduzW1ZG",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	ids := map[string]int64{}
	for i, uname := range authors {
		res, err := db.Exec(
			`INSERT INTO users (created_at, updated_at, username, password, email, language, custom_field_names, is_admin, enabled_contact_fields)
			 VALUES (?, ?, ?, ?, ?, 'es', '[]', 0, '["rut","documento","contact_person","email","phone","anniversary"]')`,
			now, now, uname, hashes[i], uname+"@historial.demo",
		)
		if err != nil {
			log.Fatalf("insert user %s: %v", uname, err)
		}
		id, _ := res.LastInsertId()
		ids[uname] = id
	}
	return ids
}

func main() {
	rand.Seed(time.Now().UnixNano())

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "meerkat.db")
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	wipe(db)
	userIDs := insertUsers(db)

	now := time.Now().UTC()
	totalNotes := 0
	used := map[string]bool{}

	for i := 0; i < 100; i++ {
		isSociety := i%3 == 2
		var first, last, slug string
		if isSociety {
			first = societies[rand.Intn(len(societies))]
			if used[first] {
				first += " " + []string{"II", "III", "Norte", "Sur"}[rand.Intn(4)]
			}
			used[first] = true
			slug = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(first, " ", "-")))
		} else {
			first = firstNames[rand.Intn(len(firstNames))]
			last = lastNames[rand.Intn(len(lastNames))]
			for used[first+" "+last] {
				last = lastNames[rand.Intn(len(lastNames))]
			}
			used[first+" "+last] = true
			slug = strings.ToLower(first + "." + last)
		}

		var p [4]int
		for j := 0; j < 4; j++ {
			p[j] = 100 + rand.Intn(9899)
		}
		exp := fmt.Sprintf("E-%d/%04d", 2025+rand.Intn(2), 1000+rand.Intn(9000))
		dom := streets[rand.Intn(len(streets))]
		t1 := []string{"SRL", "SA", "SAS"}[rand.Intn(3)]

		st := storylines[rand.Intn(len(storylines))]
		n := 3 + rand.Intn(10)
		if n > len(st.events) {
			n = len(st.events)
		}
		inclInactive := i%9 == 0
		events := append([]event{}, st.events[:n]...)
		if inclInactive {
			events = append(events, event{"Inactivo", "El cliente está inactivo: se archiva el legajo hasta nuevo aviso.", 30})
		}

		// contact
		created := now.AddDate(0, 0, -(st.events[0].ago + 5))
		var rut string
		if isSociety {
			rut = fmt.Sprintf("21000000%03d", rand.Intn(1000))
		} else {
			cit := 5000000 + rand.Intn(3000000)
			rut = fmt.Sprintf("%d.%03d.%03d-%d", cit/1000000, (cit/1000)%1000, cit%1000, rand.Intn(9))
		}
		birthday := ""
		if !isSociety {
			birthday = fmt.Sprintf("%04d-%02d-%02d", 1960+rand.Intn(40), 1+rand.Intn(12), 1+rand.Intn(28))
			if rand.Intn(4) == 0 {
				d := now.Add(time.Duration(1+rand.Intn(12)) * 24 * time.Hour)
				birthday = fmt.Sprintf("1975-%02d-%02d", int(d.Month()), d.Day())
			}
		}
		anniv := created.Format("2006-01-02")
		archivado := 0
		if inclInactive {
			archivado = 1
		}
		phone := fmt.Sprintf("09%02d %03d-%03d", 0+rand.Intn(1), rand.Intn(100)+100, rand.Intn(999))
		contactPerson := ""
		if isSociety {
			contactPerson = contactPeople[rand.Intn(len(contactPeople))]
		}
		email := "info@" + slug + ".com.uy"
		if !isSociety {
			email = slug + "@example.com"
		}
		emailsJSON := fmt.Sprintf(`[{"type":"WORK","value":"%s"}]`, email)
		phonesJSON := fmt.Sprintf(`[{"type":"CELL","value":"%s"}]`, phone)

		res, err := db.Exec(`INSERT INTO contacts
			(created_at, updated_at, firstname, lastname, nickname, email, phone, birthday, rut, documento,
			 contact_person, anniversary, custom_fields, emails, phones, user_id, archived)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?, ?)`,
			created.Format(time.RFC3339), now.Format(time.RFC3339),
			first, last, "", email, phone, birthday, rut, "",
			contactPerson, anniv, emailsJSON, phonesJSON, adminID, archivado,
		)
		if err != nil {
			log.Fatalf("insert contact: %v", err)
		}
		contactID, _ := res.LastInsertId()

		// companies (padrones / empresas)
		padronToCompany := map[int]int64{}
		if i%3 != 0 {
			nc := 1 + rand.Intn(2)
			for j := 0; j < nc; j++ {
				pn := p[rand.Intn(4)]
				num := fmt.Sprintf("%d", pn)
				cname := companyNames[rand.Intn(len(companyNames))]
				ctype := companyTypes[rand.Intn(len(companyTypes))]
				cres, err := db.Exec(`INSERT INTO companies (contact_id, company_name, company_type, company_number, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?)`,
					contactID, cname, ctype, num, created.Format(time.RFC3339), now.Format(time.RFC3339))
				if err != nil {
					log.Fatalf("insert company: %v", err)
				}
				cid, _ := cres.LastInsertId()
				padronToCompany[pn] = cid
			}
		}

		// notes, oldest first, one note per event
		authorIdx := rand.Intn(len(authors))
		for eIdx, ev := range events {
			noteDate := now.Add(time.Duration(-ev.ago)*24*time.Hour + time.Duration(rand.Intn(10))*time.Hour)
			if eIdx > 0 {
				prev := now.Add(time.Duration(-events[eIdx-1].ago) * 24 * time.Hour)
				if noteDate.Before(prev) {
					noteDate = prev.Add(30 * time.Minute)
				}
			}
			body := fill(ev.body, p, exp, dom, t1)
			body = strings.ReplaceAll(body, "@who", first)
			var companyID *int64
			if cid, ok := padronToCompany[p[0]]; ok {
				companyID = &cid
			}
			author := authors[(authorIdx+eIdx)%len(authors)]
			_, err := db.Exec(`INSERT INTO notes
				(created_at, updated_at, content, date, contact_id, user_id, title, company_id, original_title, original_content, author_id, edited_by_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
				noteDate.Format(time.RFC3339), noteDate.Format(time.RFC3339),
				body, noteDate.Format(time.RFC3339), contactID, adminID,
				ev.title, companyID, ev.title, body, userIDs[author],
			)
			if err != nil {
				log.Fatalf("insert note: %v", err)
			}
			totalNotes++
		}
	}

	fmt.Printf("seed ok: 100 clientes, %d notas, 4 usuarios (demo1234)\n", totalNotes)
}