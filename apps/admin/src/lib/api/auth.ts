import { apiClient } from './client';
import { User } from '../store/auth.store';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export interface RefreshRequest {
  refreshToken: string;
}

export interface RefreshResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
}

export const loginApi = async (email: string, password: string): Promise<LoginResponse> => {
  const response = await apiClient.post<{ data: LoginResponse }>('/auth/login', {
    email,
    password,
  });
  return response.data.data;
};

export const refreshApi = async (refreshToken: string): Promise<RefreshResponse> => {
  const response = await apiClient.post<{ data: RefreshResponse }>('/auth/refresh', {
    refreshToken,
  });
  return response.data.data;
};

export const logoutApi = async (): Promise<void> => {
  await apiClient.post('/auth/logout');
};
