import { apiClient } from './client';

// ─── Types ────────────────────────────────────────────────────────────────────

export type ReferenceStatus = 'pending' | 'positive' | 'negative' | 'unknown';

export interface Reference {
  id: string;
  application_id: string;
  referee_user_id: string;
  referee_name: string;
  referee_email: string;
  status: ReferenceStatus;
  response_type?: 'positive' | 'negative' | 'unknown';
  responded_at?: string;
  created_at: string;
}

export interface ReferenceListResponse {
  references: Reference[];
  total: number;
  responded: number;
  positive: number;
  negative: number;
  unknown: number;
}

// ─── API Functions ────────────────────────────────────────────────────────────

export async function getReferences(
  applicationId: string
): Promise<ReferenceListResponse> {
  const response = await apiClient.get<{ data: ReferenceListResponse }>(
    `/applications/${applicationId}/references`
  );
  return response.data.data;
}

export async function resendReferenceEmail(
  applicationId: string,
  referenceId: string
): Promise<void> {
  await apiClient.post(
    `/applications/${applicationId}/references/${referenceId}/resend`
  );
}
