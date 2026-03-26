import { apiClient } from './client';

// ─── Types ────────────────────────────────────────────────────────────────────

export type VoteStage = 'yk_prelim' | 'yik' | 'yk_final';
export type VoteType = 'approve' | 'abstain' | 'reject';

export interface CastVotePayload {
  vote_type: VoteType;
  reason?: string;
}

export interface VoteResponse {
  id: string;
  voter_id: string;
  voter_name: string;
  vote_stage: VoteStage;
  vote_type: VoteType;
  is_veto: boolean;
  reason?: string;
  created_at: string;
}

export interface VoteSummary {
  stage: VoteStage;
  total_voters: number;
  approved: number;
  abstained: number;
  rejected: number;
  is_terminated: boolean;
  votes: VoteResponse[]; // populated for yk / admin only
}

// Stage → URL segment mapping
const stageSegment: Record<VoteStage, string> = {
  yk_prelim: 'yk-prelim',
  yik: 'yik',
  yk_final: 'yk-final',
};

// ─── API functions ────────────────────────────────────────────────────────────

export const castVote = async (
  appId: string,
  stage: VoteStage,
  payload: CastVotePayload,
): Promise<void> => {
  await apiClient.post(
    `/applications/${appId}/votes/${stageSegment[stage]}`,
    payload,
  );
};

export const getVotes = async (
  appId: string,
  stage: VoteStage,
): Promise<VoteSummary> => {
  const { data } = await apiClient.get<{ data: VoteSummary }>(
    `/applications/${appId}/votes`,
    { params: { stage } },
  );
  return data.data;
};
