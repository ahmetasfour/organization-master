import { apiClient } from './client';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface ContactStatus {
  id: string;
  contact_name: string;
  email: string; // masked: j***@example.com
  status: 'pending' | 'clean' | 'flagged';
  responded_at?: string;
}

export interface ReputationStatus {
  application_id: string;
  total_contacts: number;
  responded: number;
  clean: number;
  flagged: number;
  contacts: ContactStatus[];
}

export interface AddContactsPayload {
  contacts: { name: string; email: string }[];
}

// ─── API calls ────────────────────────────────────────────────────────────────

/**
 * GET /api/v1/applications/:id/reputation
 * Returns aggregated reputation screening status with masked emails.
 */
export async function getReputationStatus(applicationId: string): Promise<ReputationStatus> {
  const res = await apiClient.get<{ data: ReputationStatus }>(
    `/applications/${applicationId}/reputation`
  );
  return res.data.data;
}

/**
 * POST /api/v1/applications/:id/reputation/contacts
 * Adds exactly 10 reputation contacts and triggers email sending.
 */
export async function addReputationContacts(
  applicationId: string,
  payload: AddContactsPayload
): Promise<void> {
  await apiClient.post(`/applications/${applicationId}/reputation/contacts`, payload);
}
