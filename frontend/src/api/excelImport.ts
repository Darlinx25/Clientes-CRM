// Excel bulk import API
import { apiFetch, API_BASE_URL, parseErrorResponse } from './client';

export interface ExcelImportResult {
  contacts_created: number;
  contacts_updated: number;
  companies_added: number;
  rows_processed: number;
  rows_skipped: number;
  errors: string[];
}

// Upload an Excel file (.xlsx) for bulk client import
export async function uploadExcelForImport(file: File): Promise<ExcelImportResult> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await apiFetch(
    `${API_BASE_URL}/contacts/import/excel`,
    {
      method: 'POST',
      body: formData,
    },
    120000
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}