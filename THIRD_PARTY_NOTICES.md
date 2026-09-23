# THIRD PARTY NOTICES

As avisos de licencias de las librerías de terceros usadas por este proyecto.

## Resumen

- **Licencia del proyecto:** MIT License (ver `LICENSE` en la raíz). Copyright (c) 2024 Frederic Buchner — adaptación de [Meerkat CRM](https://github.com/fbuchner/meerkat-crm).
- **Origen del código:** adaptación de Meerkat CRM (MIT, © Frederic Buchner). No incluye código de Monica CRM (AGPL-3.0); la integración con Monica se realiza vía su API remota.
- **Política de dependencias:** todas las dependencias directas e indirectas (Go y npm) son de licencia permisiva: MIT, BSD-3-Clause, ISC o Apache-2.0. No se utilizan dependencias con licencia copyleft (GPL/AGPL/LGPL).

Cada librería mantiene su propio aviso de copyright dentro de su paquete (dirigido por `go.mod` / `package-lock.json`). Las tablas siguientes listan las dependencias **directas** principales.

## Backend (Go)

| Librería | Licencia |
|---|---|
| github.com/gin-gonic/gin | MIT |
| github.com/gin-contrib/cors | MIT |
| gorm.io/gorm | MIT |
| github.com/glebarez/sqlite | MIT |
| github.com/golang-migrate/migrate/v4 | MIT |
| github.com/go-co-op/gocron | MIT |
| github.com/go-playground/validator/v10 | MIT |
| github.com/golang-jwt/jwt/v4 | MIT |
| github.com/google/uuid | BSD-3-Clause |
| github.com/rs/zerolog | MIT |
| github.com/stretchr/testify | MIT |
| github.com/xuri/excelize/v2 | BSD-3-Clause |
| github.com/coreos/go-oidc/v3 | Apache-2.0 |
| github.com/emersion/go-ical | MIT |
| github.com/emersion/go-vcard | MIT |
| github.com/emersion/go-webdav | MIT |
| github.com/gen2brain/heic | MIT |
| github.com/nfnt/resize | ISC |
| github.com/resend/resend-go/v2 | MIT |
| golang.org/x/crypto | BSD-3-Clause |
| golang.org/x/image | BSD-3-Clause |
| golang.org/x/oauth2 | BSD-3-Clause |
| golang.org/x/time | BSD-3-Clause |

Dependencias indirectas relevantes (`modernc.org/sqlite`, `github.com/tetratelabs/wazero`, `go.mongodb.org/mongo-driver`, `google.golang.org/protobuf`, `gopkg.in/yaml.v3`, `github.com/quic-go/quic-go`, etc.): todas MIT, BSD-3-Clause o Apache-2.0. Sus avisos completos viajan dentro de cada módulo en `go.sum`.

## Frontend (npm)

| Librería | Licencia |
|---|---|
| react, react-dom | MIT |
| @mui/material, @mui/icons-material, @mui/lab | MIT |
| @emotion/react, @emotion/styled | MIT |
| react-router-dom | MIT |
| i18next, react-i18next, i18next-browser-languagedetector | MIT |
| react-easy-crop | MIT |
| react-force-graph-2d | MIT |
| d3-force | ISC |
| react-scripts | MIT |
| web-vitals | Apache-2.0 |
| typescript | Apache-2.0 |
| vitest, jsdom, @testing-library/react, @testing-library/dom | MIT |
| @playwright/test (e2e) | Apache-2.0 |

Las dependencias transitivas de npm conservan sus propios archivos `LICENSE` dentro de `node_modules` (no versionado; se instala en el build).

## Conclusión legal/práctica

Por ser todo MIT/BSD/ISC/Apache-2.0 (licencias permisivas, sin copyleft), este repositorio puede:

- Publicarse en cualquier plataforma Git, público o privado.
- Modificarse y redistribuirse libremente, con la condición de conservar el aviso de copyright del proyecto (`LICENSE`) y, donde corresponda, los avisos de las librerías MIT/BSD.

Única obligación concreta: **mantener el archivo `LICENSE`** con el aviso de copyright original del autor del proyecto base.