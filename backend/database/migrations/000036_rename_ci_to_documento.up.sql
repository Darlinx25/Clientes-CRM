-- The identifier column was labelled "ci"; it really holds any free-text
-- document number (e.g. "psp A5324FB"), so it is renamed "documento".
ALTER TABLE contacts RENAME COLUMN ci TO documento;