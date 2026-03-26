'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  castVote,
  getVotes,
  CastVotePayload,
  VoteStage,
} from '../api/voting';

export const useVotes = (appId: string, stage: VoteStage) => {
  return useQuery({
    queryKey: ['votes', appId, stage],
    queryFn: () => getVotes(appId, stage),
    enabled: !!appId,
    refetchInterval: 10_000, // auto-refresh every 10 s to catch concurrent votes
  });
};

export const useCastVote = (appId: string, stage: VoteStage) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CastVotePayload) => castVote(appId, stage, payload),
    onSuccess: () => {
      // Invalidate this stage's summary and the application detail
      void queryClient.invalidateQueries({ queryKey: ['votes', appId, stage] });
      void queryClient.invalidateQueries({ queryKey: ['application', appId] });
    },
  });
};
