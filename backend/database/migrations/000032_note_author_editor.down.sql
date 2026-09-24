DROP INDEX IF EXISTS idx_notes_edited_by_id;
DROP INDEX IF EXISTS idx_notes_author_id;

ALTER TABLE notes DROP COLUMN edited_by_id;
ALTER TABLE notes DROP COLUMN author_id;