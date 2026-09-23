-- Modelo de empresas final (UY):
--  * un cliente (RUT) puede tener varios NUMEROS de empresa
--  * cada numero de empresa puede tener uno o varios TIPOS de empresa
--    (many-to-many, lista fija de 19 tipos)
--  * company_types pasa a ser lista cerrada de exactamente 19 tipos
--  * companies.company_type / contacts.company_type quedan inactivas

DELETE FROM company_types;

INSERT INTO company_types (name, created_at, updated_at) VALUES
    ('Servicios Personales', datetime('now'), datetime('now')),
    ('Monotributo', datetime('now'), datetime('now')),
    ('Rural', datetime('now'), datetime('now')),
    ('SAS', datetime('now'), datetime('now')),
    ('Administración', datetime('now'), datetime('now')),
    ('Obra', datetime('now'), datetime('now')),
    ('SRL', datetime('now'), datetime('now')),
    ('IyC', datetime('now'), datetime('now')),
    ('Literal E', datetime('now'), datetime('now')),
    ('Serv Dom', datetime('now'), datetime('now')),
    ('SA', datetime('now'), datetime('now')),
    ('Casa', datetime('now'), datetime('now')),
    ('ZF', datetime('now'), datetime('now')),
    ('Imp Pat', datetime('now'), datetime('now')),
    ('PF', datetime('now'), datetime('now')),
    ('Jubilado', datetime('now'), datetime('now')),
    ('PJExt', datetime('now'), datetime('now')),
    ('CRF', datetime('now'), datetime('now')),
    ('Rentas Exterior', datetime('now'), datetime('now'));

CREATE TABLE IF NOT EXISTS company_company_types (
    company_id      INTEGER NOT NULL,
    company_type_id INTEGER NOT NULL,
    PRIMARY KEY (company_id, company_type_id),
    FOREIGN KEY (company_id)      REFERENCES companies(id)      ON DELETE CASCADE,
    FOREIGN KEY (company_type_id) REFERENCES company_types(id)  ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_company_company_types_type ON company_company_types(company_type_id);