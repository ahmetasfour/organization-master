import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  recordConsent,
  getConsentStatus,
  getPublishedMembers,
  RecordConsentRequest,
} from '@/lib/api/webpublish';

export function useConsentStatus(applicationId: string) {
  return useQuery({
    queryKey: ['consent-status', applicationId],
    queryFn: () => getConsentStatus(applicationId),
    enabled: !!applicationId,
  });
}

export function useRecordConsent(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: RecordConsentRequest) =>
      recordConsent(applicationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consent-status', applicationId] });
      queryClient.invalidateQueries({ queryKey: ['application', applicationId] });
      queryClient.invalidateQueries({ queryKey: ['published-members'] });
    },
    onError: (error: any) => {
      console.error('Failed to record consent:', error);
    },
  });
}

export function usePublishedMembers() {
  return useQuery({
    queryKey: ['published-members'],
    queryFn: getPublishedMembers,
  });
}