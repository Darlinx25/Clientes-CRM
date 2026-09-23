-- Allow notes to be linked to a company (empresa) so timeline notes are not left unassigned.
ALTER TABLE notes ADD COLUMN company_id INTEGER;
CREATE INDEX IF NOT EXISTS idx_notes_company_id ON notes(company_id);