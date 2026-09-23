-- Add client-specific fields to contacts
ALTER TABLE contacts ADD COLUMN rut TEXT NOT NULL DEFAULT '';
ALTER TABLE contacts ADD COLUMN contact_person TEXT NOT NULL DEFAULT '';

-- Create companies table (one client can have multiple companies)
CREATE TABLE IF NOT EXISTS companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contact_id INTEGER NOT NULL,
    company_name TEXT NOT NULL DEFAULT '',
    company_type TEXT NOT NULL DEFAULT '',
    company_number TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_companies_contact_id ON companies(contact_id);
