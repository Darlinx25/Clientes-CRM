export interface AdminPanelSchema {
  current: number;
  expected: number;
  dirty: boolean;
  up_to_date: boolean;
}

export interface AdminPanelCounts {
  users: number;
  contacts: number;
  archived_contacts: number;
  deleted_contacts: number;
  notes: number;
  deleted_notes: number;
  activities: number;
  reminders: number;
  reminder_completions: number;
  companies: number;
  relationships: number;
}

export interface AdminPanelJob {
  job_name: string;
  last_run_at: string;
  locked_at: string | null;
  locked_by: string;
}

export interface AdminPanelBackupSummary {
  count: number;
  latest_filename: string;
  latest_created_at: string | null;
}

export interface AdminPanelOverview {
  app_version: string;
  schema: AdminPanelSchema;
  db_path: string;
  db_size_bytes: number;
  wal_size_bytes: number;
  backup_dir: string;
  counts: AdminPanelCounts;
  jobs: AdminPanelJob[];
  backups: AdminPanelBackupSummary;
}

export interface AdminPanelTableSummary {
  name: string;
  label: string;
  row_count: number;
}

export interface AdminPanelTableRows {
  table: string;
  label: string;
  columns: string[];
  rows: Record<string, unknown>[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface AdminPanelBackup {
  filename: string;
  kind: 'backup' | 'pre_upgrade';
  size_bytes: number;
  created_at: string;
}

export interface AdminPanelBackupsResponse {
  backups: AdminPanelBackup[];
  backup_dir: string;
}

export interface AdminPanelTableQuery {
  page?: number;
  limit?: number;
  search?: string;
  includeDeleted?: boolean;
}
