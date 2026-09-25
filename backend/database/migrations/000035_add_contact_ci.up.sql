-- Clients carry a cédula de identidad (CI) document number, stored as free text
-- (like RUT) so it can be imported and exported unchanged.
ALTER TABLE contacts ADD COLUMN ci TEXT;