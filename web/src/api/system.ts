import request from './request';
import type { SystemWebhookConfig } from '../types';

export const systemApi = {
  // Get system webhook configuration
  getWebhookConfig: () => {
    return request.get<SystemWebhookConfig>('/system/webhook');
  },

  // Update system webhook configuration
  updateWebhookConfig: (config: SystemWebhookConfig) => {
    return request.put<{ message: string }>('/system/webhook', config);
  },
};
