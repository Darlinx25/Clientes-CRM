-- Contact-note queries paginate, count and order by (contact_id, date). A
-- dedicated composite index keeps COUNT(*) and the DESC paged read on the
-- contact's own notes instead of walking every note the user owns, which is
-- what matters as years of notes accumulate.
CREATE INDEX IF NOT EXISTS idx_notes_contact_date_id ON notes(contact_id, date DESC, id DESC);