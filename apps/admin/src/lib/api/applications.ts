import { apiClient } from './client';

// ─── Types ────────────────────────────────────────────────────────────────────

export type MembershipType =
  | 'asil'
  | 'akademik'
  | 'profesyonel'
  | 'öğrenci'
  | 'onursal';

export type ApplicationStatus =
  | 'başvuru_alındı'
  | 'referans_bekleniyor'
  | 'referans_tamamlandı'
  | 'referans_red'
  | 'yk_ön_incelemede'
  | 'ön_onaylandı'
  | 'yk_red'
  | 'itibar_taramasında'
  | 'itibar_temiz'
  | 'itibar_red'
  | 'danışma_sürecinde'
  | 'danışma_red'
  | 'öneri_alındı'
  | 'yik_değerlendirmede'
  | 'yik_red'
  | 'gündemde'
  | 'kabul'
  | 'reddedildi';

export interface ApplicationSummary {
  id: string;
  applicant_name: string;
  applicant_email: string;
  membership_type: MembershipType;
  status: ApplicationStatus;
  created_at: string;
}

export interface ApplicationDetail {
  id: string;
  applicant_name: string;
  applicant_email: string;
  applicant_phone?: string;
  linkedin_url?: string;
  photo_url?: string;
  membership_type: MembershipType;
  status: ApplicationStatus;
  proposal_reason?: string;
  rejection_reason?: string;
  rejected_by_role?: string;
  repeat_applicant?: boolean;
  previous_app_id?: string;
  created_at: string;
  updated_at: string;
  allowed_next_statuses: ApplicationStatus[];
}

export interface TimelineEntry {
  status: string;
  changed_at?: string;
  changed_by?: string;
  notes?: string;
}

export interface ApplicationListResponse {
  data: ApplicationSummary[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ApplicationFilters {
  membership_type?: string;
  status?: string;
  search?: string;
  page?: number;
  page_size?: number;
}

export interface ReferenceInput {
  user_id: string;
}

export interface CreateApplicationRequest {
  applicant_name: string;
  applicant_email: string;
  applicant_phone?: string;
  linkedin_url?: string;
  photo_url?: string;
  membership_type: MembershipType;
  proposal_reason?: string;
  proposed_by_user_id?: string;
  references?: ReferenceInput[];
}

// ─── API Functions ────────────────────────────────────────────────────────────

export const getApplications = async (
  filters: ApplicationFilters = {}
): Promise<ApplicationListResponse> => {
  const params = new URLSearchParams();
  if (filters.membership_type) params.set('membership_type', filters.membership_type);
  if (filters.status) params.set('status', filters.status);
  if (filters.search) params.set('search', filters.search);
  if (filters.page) params.set('page', String(filters.page));
  if (filters.page_size) params.set('page_size', String(filters.page_size));

  const response = await apiClient.get<{ data: ApplicationListResponse }>(
    `/applications?${params.toString()}`
  );
  return response.data.data;
};

export const getApplication = async (id: string): Promise<ApplicationDetail> => {
  const response = await apiClient.get<{ data: ApplicationDetail }>(`/applications/${id}`);
  return response.data.data;
};

export const getTimeline = async (id: string): Promise<TimelineEntry[]> => {
  const response = await apiClient.get<{ data: TimelineEntry[] }>(
    `/applications/${id}/timeline`
  );
  return response.data.data ?? [];
};

export const getRedHistory = async (id: string): Promise<ApplicationDetail[]> => {
  const response = await apiClient.get<{ data: ApplicationDetail[] }>(
    `/applications/${id}/red-history`
  );
  return response.data.data ?? [];
};

export const createApplication = async (
  req: CreateApplicationRequest
): Promise<{ application: ApplicationDetail; repeat_applicant: boolean; previous_app_id?: string }> => {
  const response = await apiClient.post<{
    data: { application: ApplicationDetail; repeat_applicant: boolean; previous_app_id?: string };
  }>('/applications', req);
  return response.data.data;
};
