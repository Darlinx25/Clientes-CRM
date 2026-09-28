-- Agrega el tipo de empresa "Exonerado" a la lista fija de tipos.
INSERT OR IGNORE INTO company_types (name, created_at, updated_at) VALUES
    ('Exonerado', datetime('now'), datetime('now'));