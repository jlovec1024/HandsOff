import request from './request';
import type { GitPlatformConfig } from '../types';

export const platformApi = {
  // Get platform config
  getConfig: () => {
    return request.get<GitPlatformConfig | { exists: false; message: string }>('/platform/config');
  },

  // Update platform config
  updateConfig: (data: GitPlatformConfig) => {
    return request.put<{ message: string }>('/platform/config', data);
  },

  // Test connection
  testConnection: () => {
    return request.post<{ success: boolean; message: string; tested_at: string }>('/platform/test');
  },
};
