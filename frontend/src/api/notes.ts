// Notes-related API calls
import { apiFetch, API_BASE_URL, getAuthHeaders, parseErrorResponse } from './client';

export interface Note {
  ID: number;
  title: string;
  content: string;
  date: string;
  contact_id?: number;
  CreatedAt: string;
  UpdatedAt: string;
  original_title?: string;
  original_content?: string;
  deleted_at?: string | null;
  author_name?: string;
  contact?: {
    ID: number;
    firstname: string;
    lastname: string;
    nickname?: string;
  };
  company_id?: number;
  company?: {
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
  };
}

export interface NotesResponse {
  notes: Note[];
  total?: number;
  page?: number;
  limit?: number;
}

export interface GetNotesParams {
  page?: number;
  limit?: number;
  search?: string;
  fromDate?: string;
  toDate?: string;
}

// Get notes for a contact
export async function getContactNotes(
  contactId: string | number
): Promise<NotesResponse> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/notes`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Get soft-deleted notes for a contact (discreet "ver eliminadas" view)
export async function getDeletedContactNotes(
  contactId: string | number
): Promise<NotesResponse> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/notes?deleted=true`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Get all notes (assigned and unassigned) for the global timeline
export async function getAllNotes(
  params: GetNotesParams = {}
): Promise<NotesResponse> {
  const { page = 1, limit = 25 } = params;
  const search = params.search?.trim();

  const queryParams = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (search) {
    queryParams.append('search', search);
  }

  if (params.fromDate) {
    queryParams.append('fromDate', params.fromDate);
  }

  if (params.toDate) {
    queryParams.append('toDate', params.toDate);
  }

  const response = await apiFetch(
    `${API_BASE_URL}/notes?${queryParams.toString()}`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Get single note
export async function getNote(
  id: string | number
): Promise<Note> {
  const response = await apiFetch(
    `${API_BASE_URL}/notes/${id}`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Create note for contact
export async function createNote(
  contactId: string | number,
  data: { title: string; content: string; date: string; contact_id?: number }
): Promise<Note> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/notes`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Create unassigned note
export async function createUnassignedNote(
  data: { title: string; content: string; date: string; contact_id?: number }
): Promise<Note> {
  const response = await apiFetch(
    `${API_BASE_URL}/notes`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Update note
export async function updateNote(
  id: string | number,
  data: { title: string; content: string; date: string; contact_id?: number }
): Promise<Note> {
  const response = await apiFetch(
    `${API_BASE_URL}/notes/${id}`,
    {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Delete note
export async function deleteNote(
  id: string | number
): Promise<void> {
  const response = await apiFetch(
    `${API_BASE_URL}/notes/${id}`,
    {
      method: 'DELETE',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }
}
