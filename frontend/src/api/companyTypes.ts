// Company-type API calls (fixed shared list)
import { apiFetch, API_BASE_URL, getAuthHeaders, parseErrorResponse } from './client';

export interface CompanyType {
  ID: number;
  name: string;
  CreatedAt: string;
  UpdatedAt: string;
}

// Get all company types (fixed list, shared by all users)
export async function getCompanyTypes(): Promise<CompanyType[]> {
  const response = await apiFetch(
    `${API_BASE_URL}/company-types`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Add a new company type to the shared list (used by the "Otro" option).
// Returns the type row (existing match when the name already exists).
export async function createCompanyType(name: string): Promise<CompanyType> {
  const response = await apiFetch(
    `${API_BASE_URL}/company-types`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ name }),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}