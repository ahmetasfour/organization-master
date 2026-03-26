// ============================================================================
// MEMBERSHIP TYPES
// ============================================================================

export enum MembershipType {
  Asil = "asil",
  Akademik = "akademik",
  Profesyonel = "profesyonel",
  Ogrenci = "ogrenci",
  Onursal = "onursal",
}

// ============================================================================
// APPLICATION STATUS
// ============================================================================

export enum ApplicationStatus {
  // Common statuses
  BasvuruAlindi = "başvuru_alındı",
  Gundemde = "gündemde",
  Kabul = "kabul",
  Reddedildi = "reddedildi",

  // Asil & Akademik flow
  ReferansBekleniyor = "referans_bekleniyor",
  ReferansTamamlandi = "referans_tamamlandı",
  ReferansRed = "referans_red",
  YKOnIncelemede = "yk_ön_incelemede",
  OnOnaylandi = "ön_onaylandı",
  YKRed = "yk_red",
  ItibarTaramasinda = "itibar_taramasında",
  ItibarTemiz = "itibar_temiz",
  ItibarRed = "itibar_red",

  // Profesyonel & Öğrenci flow
  DanismaSurecinde = "danışma_sürecinde",
  DanismaRed = "danışma_red",

  // Onursal flow
  OneriAlindi = "öneri_alındı",
  YIKDegerlendirmede = "yik_değerlendirmede",
  YIKRed = "yik_red",
}

// ============================================================================
// USER ROLES
// ============================================================================

export enum UserRole {
  Admin = "admin",
  YK = "yk",
  YIK = "yik",
  Koordinator = "koordinator",
  AsilUye = "asil_uye",
}

// ============================================================================
// VOTE TYPES
// ============================================================================

export enum VoteStage {
  YKPrelim = "yk_prelim",
  YIK = "yik",
  YKFinal = "yk_final",
}

export enum VoteType {
  Approve = "approve",
  Abstain = "abstain",
  Reject = "reject",
}

// ============================================================================
// REFERENCE RESPONSE TYPES
// ============================================================================

export enum ReferenceResponseType {
  Positive = "positive",
  Unknown = "unknown",
  Negative = "negative",
}

export enum ReputationResponseType {
  Clean = "clean",
  Negative = "negative",
}

// ============================================================================
// API RESPONSE
// ============================================================================

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    fields?: Record<string, string>;
  };
}

// ============================================================================
// ENTITY INTERFACES
// ============================================================================

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Application {
  id: string;
  applicant_name: string;
  applicant_email: string;
  applicant_linkedin?: string;
  applicant_photo_url?: string;
  membership_type: MembershipType;
  status: ApplicationStatus;
  proposed_by_user_id?: string;
  proposal_reason?: string;
  rejection_reason?: string;
  rejected_by_role?: string;
  web_publish_consent?: boolean;
  is_published: boolean;
  previous_app_id?: string;
  created_at: string;
  updated_at: string;
}

export interface Reference {
  id: string;
  application_id: string;
  referee_name: string;
  referee_email: string;
  token_hash: string;
  token_expires: string;
  token_used: boolean;
  is_replacement: boolean;
  round: number;
  created_at: string;
}

export interface ReferenceResponse {
  id: string;
  reference_id: string;
  response_type: ReferenceResponseType;
  notes?: string;
  responded_at: string;
}

export interface Vote {
  id: string;
  application_id: string;
  voter_id: string;
  vote_stage: VoteStage;
  vote_type: VoteType;
  is_veto: boolean;
  reason?: string;
  created_at: string;
}

export interface Consultation {
  id: string;
  application_id: string;
  member_id: string;
  token_hash: string;
  token_expires: string;
  token_used: boolean;
  response?: string;
  responded_at?: string;
  created_at: string;
}

export interface ReputationContact {
  id: string;
  application_id: string;
  contact_email: string;
  token_hash: string;
  token_expires: string;
  token_used: boolean;
  response_type?: ReputationResponseType;
  details?: string;
  responded_at?: string;
  created_at: string;
}

export interface Log {
  id: string;
  actor_id?: string;
  actor_role?: string;
  action: string;
  entity_type?: string;
  entity_id?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

export interface WebPublishConsent {
  id: string;
  application_id: string;
  consent_given: boolean;
  recorded_by_user_id: string;
  created_at: string;
}
