'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getReputationStatus,
  addReputationContacts,
  type AddContactsPayload,
} from '../api/reputation';

/**
 * Fetches and caches the reputation screening status for an application.
 * Refetches every 30 seconds so admins see live progress.
 */
export function useReputationStatus(applicationId: string) {
  return useQuery({
    queryKey: ['reputation', applicationId],
    queryFn: () => getReputationStatus(applicationId),
    refetchInterval: 30_000,
    enabled: Boolean(applicationId),
  });
}

/**
 * Mutation to add exactly 10 reputation contacts for an application.
 * Invalidates the reputation status cache on success.
 */
export function useAddReputationContacts(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: AddContactsPayload) =>
      addReputationContacts(applicationId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reputation', applicationId] });
    },
  });
}
