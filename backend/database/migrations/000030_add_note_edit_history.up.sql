-- Keep a faithful copy of a note's original title/content so edits can show
-- "edited on <date>" and let the user view the original.
ALTER TABLE notes ADD COLUMN original_title TEXT;
ALTER TABLE notes ADD COLUMN original_content TEXT;

-- Backfill existing notes: the original is their current content.
UPDATE notes SET original_title = title, original_content = content
WHERE original_content IS NULL OR original_content = '';