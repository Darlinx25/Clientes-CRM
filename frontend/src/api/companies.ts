// Companies-related API calls
import { apiFetch, API_BASE_URL, getAuthHeaders, parseErrorResponse } from './client';

export interface Company {
  ID: number;
  company_number?: string;
  aportacion?: string;
  contact_id: number;
  types?: { ID: number; name: string }[];
  contact?: {
    ID: number;
    firstname: string;
    lastname: string;
  };
}

// List all companies across the current user's contacts (for the timeline note picker)
export async function getAllCompanies(): Promise<Company[]> {
  const response = await apiFetch(`${API_BASE_URL}/companies`, {
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}