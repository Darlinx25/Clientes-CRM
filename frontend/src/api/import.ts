// Shared types for import previews / confirmations.
// Kept here because the Monica importer reuses the same shapes formerly used by the
// (now removed) CSV/VCF import dialog.

// Match info for duplicate detection
export interface DuplicateMatch {
  existing_contact_id: number;
  existing_firstname: string;
  existing_lastname: string;
  existing_email: string;
  existing_phone: string;
  match_reason: 'name' | 'email' | 'phone' | 'vcard_uid';
  existing_deleted: boolean;
}

// Preview row with parsed contact and status
export interface ImportRowPreview {
  row_index: number;
  parsed_contact: Record<string, string>;
  validation_errors: string[];
  duplicate_match: DuplicateMatch | null;
  suggested_action: 'add' | 'skip' | 'update';
}

// Action for a specific row
export interface RowImportAction {
  row_index: number;
  action: 'skip' | 'add' | 'update';
}

// Final import result
export interface ImportResult {
  total_processed: number;
  created: number;
  updated: number;
  skipped: number;
  errors: string[];
}