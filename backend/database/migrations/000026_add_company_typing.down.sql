DROP TABLE IF EXISTS company_types;

ALTER TABLE companies DROP COLUMN IF EXISTS aportacion;

ALTER TABLE contacts DROP COLUMN IF EXISTS company_type;