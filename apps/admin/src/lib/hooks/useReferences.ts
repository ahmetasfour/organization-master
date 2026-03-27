'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getReferences, resendReferenceEmail } from '../api/references';

export function useReferences(applicationId: string) {
  return useQuery({
    queryKey: ['references', applicationId],
    queryFn: () => getReferences(applicationId),
    enabled: !!applicationId,
    refetchInterval: 30_000, // Auto-refresh every 30s
  });
}

export function useResendReferenceEmail(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (referenceId: string) =>
      resendReferenceEmail(applicationId, referenceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['references', applicationId] });
    },
  });
}
