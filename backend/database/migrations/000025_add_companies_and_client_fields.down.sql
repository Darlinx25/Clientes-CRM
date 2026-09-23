-- Remove companies table
DROP TABLE IF EXISTS companies;

-- Remove client-specific fields from contacts
ALTER TABLE contacts DROP COLUMN IF EXISTS rut;
ALTER TABLE contacts DROP COLUMN IF EXISTS contact_person;
