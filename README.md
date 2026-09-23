# Clientes - CRM simplificado para gestión de clientes

CRM autoalojado para gestión de clientes de oficina, con timeline de notas por cliente.
Basado en [Meerkat CRM](https://github.com/fbuchner/meerkat-crm) simplificado.

---

## Qué es esta app

Un sistema interno para registrar y dar seguimiento a clientes de oficina. Cada cliente tiene una ficha con sus datos de contacto y un timeline donde se registran notas cronológicas de cada interacción o gestión realizada.

**Lo que tiene:**
- **Clientes** (`/contacts`) — lista de clientes con búsqueda, filtros, paginación
- **Ficha de cliente** (`/contacts/:id`) — datos del cliente + timeline de notas + recordatorios
- **Línea de Tiempo** (`/timeline`) — vista global de todas las notas ordenadas por fecha, con nombre del cliente
- **Configuración** (`/settings`) — perfil de usuario

**Lo que NO tiene** (fue eliminado del CRM original): Dashboard, Actividades globales, Notas globales, Network, API Tokens, Users admin, Data Settings, Relaciones entre contactos.

**Idioma:** español forzado.

---

## Cómo correr la app

### Requisitos
- [Docker](https://docs.docker.com/get-docker/) instalado

### 1. Clonar o copiar el repositorio
```sh
git clone https://github.com/fbuchner/meerkat-crm.git
cd meerkat-crm
```

### 2. Configurar variables de entorno
```sh
cp .env.example .env
nano .env   # o vim, o el editor que prefieras
```

**Variables mínimas necesarias en `.env`:**
```sh
JWT_SECRET_KEY=tu-clave-secreta-larga-aqui    # OBLIGATORIO. Cambiar esto invalida logins.
PUID=1001
PGID=1001
DATA_PATH=./data
PHOTOS_PATH=./photos
FRONTEND_PORT=7300
```

### 3. Construir e iniciar
```sh
docker compose up -d --build
```

La primera vez tarda ~5-10 min (descarga dependencias y compila). Las veces siguientes arranca en segundos.

### 4. Acceder
Abrir http://localhost:7300 en el navegador.

Al acceder por primera vez, se muestra la pantalla de registro para crear el usuario admin.

---

## Arquitectura técnica

```
┌──────────────────────────────────────────────┐
│  Contenedor Docker (nginx + supervisord)     │
│                                              │
│  ┌─────────────┐    ┌──────────────────┐     │
│  │ nginx:8080  │───▶│ Go backend:8081  │     │
│  │ (frontend)  │    │ (API + SQLite)   │     │
│  └─────────────┘    └──────────────────┘     │
│         │                    │                │
│    build/              data/meerkat.db       │
│    (React SPA)         (base de datos)       │
└──────────────────────────────────────────────┘
```

- **Frontend:** React 19 + TypeScript + Material UI 7, build via react-scripts
- **Backend:** Go 1.25 + Gin (web framework) + GORM (ORM)
- **Base de datos:** SQLite (driver puro Go, sin CGO)
- **Nginx:** sirve el frontend estático y proxea `/api/*` al backend
- **Supervisord:** maneja ambos procesos (nginx + backend)

---

## Estructura de archivos relevante

```
meerkat-crm/
├── backend/
│   ├── main.go                     # Entry point
│   ├── config/                     # Variables de entorno
│   ├── controllers/
│   │   ├── note_controller.go      # CRUD notas + GetAllNotes (timeline global)
│   │   ├── contact_controller.go   # CRUD contactos
│   │   └── ...
│   ├── models/
│   │   ├── note.go                 # Modelo Note (con campo Title)
│   │   ├── contact.go              # Modelo Contact
│   │   ├── dtos.go                 # DTOs (NoteInput con Title)
│   │   └── ...
│   ├── database/
│   │   ├── migrations/
│   │   │   ├── 000024_add_note_title.up.sql    # Agrega columna title a notes
│   │   │   └── 000024_add_note_title.down.sql
│   │   └── migrate.go              # Auto-aplica migraciones al iniciar
│   └── routes/routes.go            # Registro de rutas API
│
├── frontend/
│   └── src/
│       ├── App.tsx                 # Rutas y navegación (solo 3 items)
│       ├── ContactDetailPage.tsx   # Ficha cliente (timeline solo notas)
│       ├── ContactsPage.tsx        # Lista de clientes
│       ├── TimelinePage.tsx        # Timeline global (todas las notas)
│       ├── SettingsPage.tsx        # Perfil usuario
│       ├── api/notes.ts            # API client de notas (con title)
│       ├── components/
│       │   ├── ContactTimeline.tsx  # Render timeline (solo notas)
│       │   ├── AddNoteDialog.tsx    # Dialog nueva nota (con título)
│       │   └── EditTimelineItemDialog.tsx  # Editar nota
│       ├── hooks/
│       │   ├── useTimelineEditing.ts   # Editing de notas
│       │   └── useContactDialogs.ts    # Dialogs de contacto
│       └── i18n/
│           ├── config.ts           # Idioma forzado a español
│           └── locales/es.json     # Traducciones español
│
├── docker-compose.yml              # Para production
├── Dockerfile                      # Multi-stage build
└── .env.example                    # Template de configuración
```

---

## Cómo funciona el timeline

### Por cliente (`/contacts/:id`)
Cada cliente tiene una pestaña "Timeline" que muestra todas sus notas ordenadas por fecha (más reciente primero). Cada nota tiene título (bold) y contenido.

### Global (`/timeline`)
La línea de tiempo global muestra **todas las notas de todos los clientes**, ordenadas por fecha. Cada entrada muestra el título de la nota, el contenido, y el nombre del cliente (clickeable para ir a su ficha).

Búsqueda: por título, contenido, o nombre del cliente.
Filtros: por fecha desde/hasta.

---

## Base de datos

SQLite, ubicada en `./data/meerkat.db`. Las migraciones se aplican automáticamente al iniciar el backend. La migración más reciente es la `000024` que agrega el campo `title` a la tabla `notes`.

### Backup
```sh
cp ./data/meerkat.db ./data/meerkat.db.backup-$(date +%Y%m%d)
```

### Restaurar
```sh
# Detener la app primero
docker compose down
cp ./data/meerkat.db.backup-YYYYMMDD ./data/meerkat.db
docker compose up -d
```

---

## Desarrollo local

Si querés iterar sin Docker (requiere Go 1.25+ y Node.js 22.12+):

### Backend
```sh
cd backend
cp .env.example .env
# Editar .env (mínimo: JWT_SECRET_KEY)
export $(grep -v '^#' .env | xargs)
go run main.go
# API en http://localhost:8081
```

### Frontend
```sh
cd frontend
yarn install
REACT_APP_API_URL="" yarn start
# App en http://localhost:3000 (proxy a backend)
```

### Tests
```sh
# Frontend
cd frontend && npx vitest run

# Backend
cd backend && go test ./...
```

---

## API principal

Todas las rutas bajo `/api/v1/`.

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/register` | Crear usuario |
| POST | `/login` | Iniciar sesión |
| GET | `/contacts` | Listar clientes |
| POST | `/contacts` | Crear cliente |
| GET | `/contacts/:id` | Ver cliente |
| PUT | `/contacts/:id` | Editar cliente |
| DELETE | `/contacts/:id` | Eliminar cliente |
| GET | `/contacts/:id/notes` | Notas de un cliente |
| POST | `/contacts/:id/notes` | Crear nota para cliente |
| GET | `/notes` | Todas las notas (timeline global) |
| PUT | `/notes/:id` | Editar nota |
| DELETE | `/notes/:id` | Eliminar nota |
| GET | `/contacts/:id/reminders` | Recordatorios de un cliente |
| POST | `/contacts/:id/reminders` | Crear recordatorio |
| GET | `/settings` | Configuración usuario |
| PUT | `/settings` | Actualizar configuración |

---

## Cambios respecto a Meerkat CRM original

| Aspecto | Original | Ahora |
|---------|----------|-------|
| Módulos | Dashboard, Contacts, Activities, Notes, Network, Settings, Data, API Tokens, Users | Clientes, Línea de Tiempo, Configuración |
| Timeline por cliente | Notas + Actividades + Recordatorios completados | Solo notas (con título) |
| Notas | Sin título, solo contenido | Con título + contenido |
| Relaciones entre contactos | Sí (many-to-many) | Eliminadas |
| Activities (globales) | Sí | Eliminadas de UI (API sigue en backend) |
| Nota global | Sí | Reemplazada por Línea de Tiempo |
| Idioma | EN/DE/IT/ES/FR selectable | Español forzado |
| Nav principal | 5 items + submenu settings | 3 items (Clientes, Timeline, Config) |
| Nombre app | "Meerkat CRM" | "Clientes" |

### Archivos eliminados del frontend
```
DashboardPage.tsx, ActivitiesPage.tsx, NotesPage.tsx,
NetworkPage.tsx, ApiTokensPage.tsx, UsersPage.tsx, DataSettingsPage.tsx
```
Los hooks y componentes huérfanos (useRelationships, useActivities, useGraph, RelationshipList, AddRelationshipDialog, NetworkPage, etc.) quedaron en el repo pero no se usan ni se importan. Se pueden borrar por limpieza.

### Archivos backend no modificados pero no usados
Las rutas de Activities, Relationships, Graph, API Tokens, Webhooks, Calendars, CardDAV, Admin users siguen registradas en `routes/routes.go` pero la UI no las consume. Funcionan si se usan vía API directa.

---

## Variables de entorno (`.env`)

| Variable | Requerido | Descripción |
|----------|-----------|-------------|
| `JWT_SECRET_KEY` | Sí | Clave secreta para JWT (cambiarla invalida logins existentes) |
| `PUID` | No | UID del proceso (default: 1001) |
| `PGID` | No | GID del proceso (default: 1001) |
| `DATA_PATH` | No | Ruta de la DB SQLite (default: `./data`) |
| `PHOTOS_PATH` | No | Ruta de fotos de perfil (default: `./photos`) |
| `FRONTEND_PORT` | No | Puerto externo (default: 7300) |
| `PORT` | No | Puerto interno del backend (default: 8081, no cambiar) |
| `SMTP_HOST` | No | Servidor SMTP para notificaciones |
| `SMTP_PORT` | No | Puerto SMTP |
| `SMTP_USER` | No | Usuario SMTP |
| `SMTP_PASSWORD` | No | Contraseña SMTP |
| `RESEND_API_KEY` | No | API key de Resend (alternativa a SMTP) |
| `ADMIN_EMAIL` | No | Email para notificaciones de admin |

---

## Licencias y publicación en Git

Todo lo relativo a licencias, en un solo lugar, para poder hostear este repositorio en Git (GitHub, GitLab, etc.) sin problemas.

### Licencia del proyecto

El proyecto se publica bajo **MIT License**. El archivo [`LICENSE`](./LICENSE) en la raíz conserva el copyright original:

> Copyright (c) 2024 Frederic Buchner — proyecto **Meerkat CRM**, del que esta app es una adaptación.

MIT es una licencia **permisiva**: permite usar, copiar, modificar, publicar, distribuir y vender el software, libre y sin costo, con una única condición: **conservar el aviso de copyright y permiso**. Por lo tanto, al publicar este repositorio debés:

- Mantener el archivo `LICENSE` **intacto** (incluida la línea de copyright original).
- No reemplazarlo por otro copyright como si fuera código 100% propio, aunque podés agregar tu propio aviso adicional si modificás el código.

### Origen del código

- Base: **[Meerkat CRM](https://github.com/fbuchner/meerkat-crm)** — MIT License, © Frederic Buchner. Un CRM autoalojado del que esta app es una versión simplificada (clientes + notas/timeline + configuración).
- La importación desde **Monica CRM** se hace **vía su API remota**: no se copia ni se distribuye código de Monica en este repositorio. Monica es **AGPL-3.0** y su código **no** está incluido acá.
- No se incluyen assets de terceros con licencias restrictivas (fuentes, iconografías comerciales, imágenes). Todo el contenido gráfico/UI es de Material UI (MIT) o propio.

### Dependencias: no hay copyleft

Todas las dependencias directas e indirectas (Go modules y paquetes npm) son de licencia **permisiva** (MIT, BSD-3-Clause, ISC, Apache-2.0). No hay GPL / AGPL / LGPL en ninguna dependencia, por lo que:

- **No** existe obligación de liberar esta app por efecto "viral" de copyleft.
- Podés alojarla en un repo privado **o** público, sin obligación de compartir código fuente modificado.

El detalle de las librerías principales y sus licencias está en [`THIRD_PARTY_NOTICES.md`](./THIRD_PARTY_NOTICES.md).

### Checklist antes de publicar en Git

1. **Secretos:** `.env` está en `.gitignore` (patrón `*.env`). Nunca subas `JWT_SECRET_KEY`, `SMTP_PASSWORD` ni `RESEND_API_KEY`. Solo se versiona `.env.example` con valores de ejemplo.
2. **Datos reales de clientes:** `data/` (SQLite) y `photos/` están en `.gitignore` **desde esta versión**. Confirmá que no se suba `meerkat.db` ni fotos reales (los tests usan solamente datos de ejemplo).
3. **Licencia:** mantener `LICENSE` con el copyright original de Frederic Buchner (ver más arriba).
4. **Público vs. privado:** con MIT el repo puede ser público o privado indiferentemente; si elegís público, no incluyas tampoco información de clientes reales en fixtures, screenshots ni ejemplos — usá datos ficticios.
5. **No versionar dependencias:** `node_modules/`, `frontend/build/`, `vendor/` están ignorados. El instalador las descarga al construir.

---

## Troubleshooting

**La app no arranca:**
```sh
docker compose logs meerkat  # ver errores
```

**Puerto ya en uso:**
Cambiar `FRONTEND_PORT` en `.env` (ej: `FRONTEND_PORT=8080`).

**Resetear la base de datos completa:**
```sh
docker compose down
rm -rf ./data/meerkat.db*
docker compose up -d
```

**Actualizar la app con nuevos cambios:**
```sh
docker compose up -d --build
```

**El container no levanta (error de migración):**
Las migraciones corren automáticamente. Si falla, probablemente el SQLite está corrupto. Restaurar desde backup o resetear (ver arriba).
