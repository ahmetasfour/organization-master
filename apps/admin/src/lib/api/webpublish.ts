import { apiClient } from './client';

export interface RecordConsentRequest {
  consented: boolean;
}

export interface ConsentResponse {
  application_id: string;
  consented: boolean;
  is_published: boolean;
  recorded_at: string;
}

export interface ConsentStatusResponse {
  application_id: string;
  recorded: boolean;
  consented?: boolean;
  is_published?: boolean;
  recorded_at?: string;
}

export interface MemberListItem {
  full_name: string;
  membership_type: string;
  accepted_at: string;
}

export async function recordConsent(
  applicationId: string,
  data: RecordConsentRequest
): Promise<ConsentResponse> {
  const response = await apiClient.post<{ data: ConsentResponse }>(
    `/applications/${applicationId}/publish-consent`,
    data
  );
  return response.data.data;
}

export async function getConsentStatus(
  applicationId: string
): Promise<ConsentStatusResponse> {
  const response = await apiClient.get<{ data: ConsentStatusResponse }>(
    `/applications/${applicationId}/publish-consent`
  );
  return response.data.data;
}

export async function getPublishedMembers(): Promise<MemberListItem[]> {
  const response = await apiClient.get<{ data: MemberListItem[] }>('/members');
  return response.data.data;
}