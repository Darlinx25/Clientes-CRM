import {
  apiFetchAdminPanel,
  API_BASE_URL,
  ApiError,
  getAuthHeaders,
  parseErrorResponse,
} from './client';
import type {
  AdminPanelBackupsResponse,
  AdminPanelOverview,
  AdminPanelTableQuery,
  AdminPanelTableRows,
  AdminPanelTableSummary,
} from '../types/adminPanel';

export async function startAdminPanelSession(
  password: string
): Promise<{ expires_in: number }> {
  const response = await apiFetchAdminPanel(`${API_BASE_URL}/admin/panel/session`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ password }),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

export async function endAdminPanelSession(): Promise<void> {
  const response = await apiFetchAdminPanel(`${API_BASE_URL}/admin/panel/session`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }
}

export async function getAdminPanelOverview(): Promise<AdminPanelOverview> {
  const response = await apiFetchAdminPanel(`${API_BASE_URL}/admin/panel/overview`, {
    method: 'GET',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

export async function getAdminPanelTables(): Promise<AdminPanelTableSummary[]> {
  const response = await apiFetchAdminPanel(`${API_BASE_URL}/admin/panel/tables`, {
    method: 'GET',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  const data: { tables: AdminPanelTableSummary[] } = await response.json();
  return data.tables;
}

export async function getAdminPanelTable(
  table: string,
  query: AdminPanelTableQuery = {}
): Promise<AdminPanelTableRows> {
  const params = new URLSearchParams();
  if (query.page) params.set('page', String(query.page));
  if (query.limit) params.set('limit', String(query.limit));
  if (query.search) params.set('search', query.search);
  if (query.includeDeleted) params.set('include_deleted', 'true');

  const suffix = params.toString() ? `?${params.toString()}` : '';
  const response = await apiFetchAdminPanel(
    `${API_BASE_URL}/admin/panel/tables/${encodeURIComponent(table)}${suffix}`,
    {
      method: 'GET',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

export async function getAdminPanelBackups(): Promise<AdminPanelBackupsResponse> {
  const response = await apiFetchAdminPanel(`${API_BASE_URL}/admin/panel/backups`, {
    method: 'GET',
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

export async function downloadAdminPanelBackup(filename: string): Promise<Blob> {
  const response = await apiFetchAdminPanel(
    `${API_BASE_URL}/admin/panel/backups/${encodeURIComponent(filename)}`,
    {
      method: 'GET',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.blob();
}

export function isAdminSessionRequiredError(err: unknown): boolean {
  return err instanceof ApiError && err.status === 401 && err.code === 'ADMIN_SESSION_REQUIRED';
}
