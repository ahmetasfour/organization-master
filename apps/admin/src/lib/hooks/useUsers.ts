import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "../api/client";

export interface User {
  id: string;
  full_name: string;
  email: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
}

export interface UserFilters {
  role?: string;
  is_active?: boolean;
  search?: string;
  page?: number;
  page_size?: number;
}

export interface CreateUserInput {
  full_name: string;
  email: string;
  password: string;
  role: string;
}

export interface UpdateUserInput {
  full_name?: string;
  role?: string;
  is_active?: boolean;
}

// List users with filters
export function useUsers(filters: UserFilters = {}) {
  return useQuery({
    queryKey: ["users", filters],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters.role) params.append("role", filters.role);
      if (filters.is_active !== undefined)
        params.append("is_active", String(filters.is_active));
      if (filters.search) params.append("search", filters.search);
      if (filters.page) params.append("page", String(filters.page));
      if (filters.page_size) params.append("page_size", String(filters.page_size));

      const response = await api.get(`/users?${params.toString()}`);
      return response.data.data;
    },
  });
}

// Get single user
export function useUser(id: string) {
  return useQuery({
    queryKey: ["user", id],
    queryFn: async () => {
      const response = await api.get(`/users/${id}`);
      return response.data.data as User;
    },
    enabled: !!id,
  });
}

// Create user
export function useCreateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateUserInput) => {
      try {
        const response = await api.post("/users", input);
        return response.data;
      } catch (error: any) {
        if (error.response?.data?.error?.message) {
          throw new Error(error.response.data.error.message);
        }
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}

// Update user
export function useUpdateUser(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: UpdateUserInput) => {
      try {
        const response = await api.patch(`/users/${id}`, input);
        return response.data;
      } catch (error: any) {
        if (error.response?.data?.error?.message) {
          throw new Error(error.response.data.error.message);
        }
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      queryClient.invalidateQueries({ queryKey: ["user", id] });
    },
  });
}

// List active users (for member selection)
export function useActiveUsers(role?: string) {
  return useQuery({
    queryKey: ["users", "active", role],
    queryFn: async () => {
      const params = role ? `?role=${role}` : "";
      const response = await api.get(`/users/active${params}`);
      return response.data.data as User[];
    },
  });
}
