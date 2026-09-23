-- Typos de empresa / clientes como empresas (UY)
-- Descripción: un cliente (contacto) es una empresa:
--  * company_type en contacts -> tipo de empresa del cliente
--  * company_types -> lista de tipos de empresa precargada, compartida y ampliable
--  * aportacion en companies -> contribución de cada sub-empresa del RUT

ALTER TABLE contacts ADD COLUMN company_type TEXT NOT NULL DEFAULT '';

ALTER TABLE companies ADD COLUMN aportacion TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS company_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

INSERT OR IGNORE INTO company_types (name, created_at, updated_at) VALUES
    ('S.A.', datetime('now'), datetime('now')),
    ('S.R.L.', datetime('now'), datetime('now')),
    ('Unipersonal', datetime('now'), datetime('now')),
    ('Sociedad por Acciones', datetime('now'), datetime('now')),
    ('Sociedad Civil', datetime('now'), datetime('now')),
    ('Cooperativa', datetime('now'), datetime('now')),
    ('Sucursal Extranjera', datetime('now'), datetime('now'));