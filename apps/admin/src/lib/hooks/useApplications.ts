'use client';

import { useQuery } from '@tanstack/react-query';
import {
  getApplications,
  getApplication,
  getTimeline,
  getRedHistory,
  ApplicationFilters,
} from '../api/applications';

export const useApplications = (filters: ApplicationFilters = {}) => {
  return useQuery({
    queryKey: ['applications', filters],
    queryFn: () => getApplications(filters),
    placeholderData: (prev) => prev,
  });
};

export const useApplication = (id: string) => {
  return useQuery({
    queryKey: ['application', id],
    queryFn: () => getApplication(id),
    enabled: !!id,
  });
};

export const useTimeline = (id: string) => {
  return useQuery({
    queryKey: ['application-timeline', id],
    queryFn: () => getTimeline(id),
    enabled: !!id,
  });
};

export const useRedHistory = (id: string, enabled = true) => {
  return useQuery({
    queryKey: ['application-red-history', id],
    queryFn: () => getRedHistory(id),
    enabled: !!id && enabled,
  });
};
