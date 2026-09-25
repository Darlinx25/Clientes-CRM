// Contact-related API calls
import { apiFetch, API_BASE_URL, getAuthHeaders, parseErrorResponse } from './client';
import type { CompanyType } from './companyTypes';

export interface ContactValue {
  type: string;
  value: string;
}

export interface ContactAddress {
  type: string;
  street: string;
  city: string;
  region: string;
  postal: string;
  country: string;
}

export interface Company {
  ID: number;
  contact_id: number;
  company_number: string;
  custom_type?: string;
  types: CompanyType[];
}

// Payload for creating/updating a company (número de empresa). custom_type is
// the free-text "Otro" exception stored on the company itself; providing it
// does not add a new shared CompanyType.
export interface CompanyInput {
  company_number: string;
  type_ids: number[];
  custom_type?: string;
}

export interface Contact {
  ID: number;
  firstname: string;
  lastname: string;
  nickname?: string;
  gender?: string;
  email?: string;
  phone?: string;
  birthday?: string;
  photo?: string;
  address?: string;
  how_we_met?: string;
  food_preference?: string;
  work_information?: string;
  contact_information?: string;
  photo_thumbnail?: string;
  custom_fields?: Record<string, string>;
  archived?: boolean;
  // Multi-valued vCard fields
  emails?: ContactValue[];
  phones?: ContactValue[];
  addresses?: ContactAddress[];
  urls?: ContactValue[];
  impps?: ContactValue[];
  // Structured name parts
  prefix?: string;
  middle_name?: string;
  suffix?: string;
  // Organizational fields
  organization?: string;
  department?: string;
  job_title?: string;
  role?: string;
  anniversary?: string;
  // Client-specific fields
  rut?: string;
  documento?: string;
  contact_person?: string;
  companies?: Company[];
}

export interface Birthday {
  type: 'contact' | 'relationship';
  name: string;
  birthday: string;
  photo_thumbnail?: string;
  contact_id: number;
  relationship_type?: string;
  associated_contact_name?: string;
}

export interface ContactsResponse {
  contacts: Contact[];
  total: number;
  page: number;
  limit: number;
}

export interface GetContactsParams {
  page?: number;
  limit?: number;
  search?: string;
  sort?: string;
  order?: string;
  includeArchived?: boolean;
  archived?: boolean;
}

// Get all contacts with pagination and filters
export async function getContacts(
  params: GetContactsParams
): Promise<ContactsResponse> {
  const { page = 1, limit = 25, search = '', sort, order, includeArchived, archived } = params;

  const queryParams = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (search) queryParams.append('search', search);
  if (sort) queryParams.append('sort', sort);
  if (order) queryParams.append('order', order);
  if (includeArchived) queryParams.append('include_archived', 'true');
  if (archived !== undefined) queryParams.append('archived', archived.toString());

  queryParams.append('fields', 'ID,firstname,lastname,nickname,photo_thumbnail,archived,rut');

  const response = await apiFetch(
    `${API_BASE_URL}/contacts?${queryParams.toString()}`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Get single contact
export async function getContact(
  id: string | number,
  fields?: string[]
): Promise<Contact> {
  let url = `${API_BASE_URL}/contacts/${id}`;
  if (fields && fields.length > 0) {
    url += `?fields=${fields.join(',')}`;
  }

  const response = await apiFetch(
    url,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Create contact
export async function createContact(
  data: Partial<Contact>
): Promise<Contact> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  const result = await response.json();
  return result.contact || result;
}

// Update contact
export async function updateContact(
  id: string | number,
  data: Partial<Contact>
): Promise<Contact> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${id}`,
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

// Delete contact
export async function deleteContact(
  id: string | number
): Promise<void> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${id}`,
    {
      method: 'DELETE',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }
}

// Get contact profile picture
export async function getContactProfilePicture(
  id: string | number,
  thumbnail: boolean = false
): Promise<Blob | null> {
  const url = thumbnail
    ? `${API_BASE_URL}/contacts/${id}/profile_picture?thumbnail=true`
    : `${API_BASE_URL}/contacts/${id}/profile_picture`;
  const response = await apiFetch(url);

  if (!response.ok) {
    return null;
  }

  return response.blob();
}

// Upload contact profile picture
export async function uploadProfilePicture(
  id: string | number,
  imageBlob: Blob
): Promise<void> {
  const formData = new FormData();
  formData.append('photo', imageBlob, 'profile.jpg');

  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${id}/profile_picture`,
    {
      method: 'POST',
      body: formData
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }
}

// Get random contacts (returns 5 contacts)
export async function getRandomContacts(): Promise<Contact[]> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/random`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  const data = await response.json();
  return data.contacts || [];
}

// Get upcoming birthdays (returns up to 10 birthdays for contacts and relationships)
export async function getUpcomingBirthdays(): Promise<Birthday[]> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/birthdays`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  const data = await response.json();
  return data.birthdays || [];
}

// Archive a contact (deletes all reminders)
export async function archiveContact(
  id: string | number
): Promise<Contact> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${id}/archive`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Unarchive a contact
export async function unarchiveContact(
  id: string | number
): Promise<Contact> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${id}/unarchive`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// ---- Companies (client companies) ----

// Get all companies for a contact
export async function getCompanies(contactId: string | number): Promise<Company[]> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/companies`,
    { headers: getAuthHeaders() }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json();
}

// Create a company for a contact
export async function createCompany(
  contactId: string | number,
  data: CompanyInput
): Promise<Company> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/companies`,
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

// Update a company
export async function updateCompany(
  contactId: string | number,
  companyId: number,
  data: CompanyInput
): Promise<Company> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/companies/${companyId}`,
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

// Delete a company
export async function deleteCompany(
  contactId: string | number,
  companyId: number
): Promise<void> {
  const response = await apiFetch(
    `${API_BASE_URL}/contacts/${contactId}/companies/${companyId}`,
    {
      method: 'DELETE',
      headers: getAuthHeaders(),
    }
  );

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }
}
