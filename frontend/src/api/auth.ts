import { apiClient } from './client';
import type { User } from '../types';

interface LoginPayload { email: string; password: string }
interface RegisterPayload { email: string; password: string; name: string }
interface AuthResponse { access_token: string; user: User }

export const authApi = {
  login: (data: LoginPayload) =>
    apiClient.post<AuthResponse>('/auth/login', data).then(r => r.data),

  register: (data: RegisterPayload) =>
    apiClient.post<AuthResponse>('/auth/register', data).then(r => r.data),

  me: () =>
    apiClient.get<User>('/auth/me').then(r => r.data),
};