-- Record who created and who last edited a note. On unified deployments the
-- notes.user_id column holds the shared owner, so author_id/edited_by_id keep
-- the real users behind each note for display ("created by X", "edited by Y").
ALTER TABLE notes ADD COLUMN author_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE notes ADD COLUMN edited_by_id INTEGER NOT NULL DEFAULT 0;

-- Best-effort backfill: for notes written before unification user_id was the
-- actual author; after unification it is the shared owner (best guess).
UPDATE notes SET author_id = user_id WHERE author_id = 0;

CREATE INDEX IF NOT EXISTS idx_notes_author_id ON notes(author_id);
CREATE INDEX IF NOT EXISTS idx_notes_edited_by_id ON notes(edited_by_id);