'use client';

import { useQuery } from '@tanstack/react-query';
import { getLogs, LogFilters } from '../api/logs';

export function useLogs(filters: LogFilters = {}) {
  return useQuery({
    queryKey: ['logs', filters],
    queryFn: () => getLogs(filters),
    placeholderData: (prev) => prev,
  });
}
