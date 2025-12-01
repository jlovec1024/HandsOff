import request from './request';
import type { LoginRequest, LoginResponse, User } from '../types';

export const authApi = {
  // Login
  login: (data: LoginRequest) => {
    return request.post<LoginResponse>('/auth/login', data);
  },

  // Logout
  logout: () => {
    return request.post('/auth/logout');
  },

  // Get current user
  getCurrentUser: () => {
    return request.get<User>('/auth/user');
  },
};
