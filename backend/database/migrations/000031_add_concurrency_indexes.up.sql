-- Composite indexes for the queries that serve the contact list, the global
-- timeline and the reminders list. Single-column indexes existed; these cover
-- the (user_id, filter, order) patterns actually used, so SQLite can serve the
-- pages without a full scan on shared deployments.
CREATE INDEX IF NOT EXISTS idx_contacts_user_archived_id ON contacts(user_id, archived, id DESC);

CREATE INDEX IF NOT EXISTS idx_notes_user_date_id ON notes(user_id, date DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_reminders_user_remind_at ON reminders(user_id, remind_at);