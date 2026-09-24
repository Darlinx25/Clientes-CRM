-- Companies record the free-text "Otro" exception (e.g. "Sociedad Nueva SA")
-- on the company row itself without creating a new shared CompanyType.
ALTER TABLE companies ADD COLUMN custom_type TEXT DEFAULT '';