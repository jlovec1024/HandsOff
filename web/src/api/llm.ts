import request from './request';
import type { LLMProvider, LLMModel } from '../types';

export const llmApi = {
  // Provider APIs
  listProviders: () => {
    return request.get<LLMProvider[]>('/llm/providers');
  },

  getProvider: (id: number) => {
    return request.get<LLMProvider>(`/llm/providers/${id}`);
  },

  createProvider: (data: Omit<LLMProvider, 'id'>) => {
    return request.post<LLMProvider>('/llm/providers', data);
  },

  updateProvider: (id: number, data: Partial<LLMProvider>) => {
    return request.put<LLMProvider>(`/llm/providers/${id}`, data);
  },

  deleteProvider: (id: number) => {
    return request.delete<{ message: string }>(`/llm/providers/${id}`);
  },

  testProvider: (id: number) => {
    return request.post<{ success: boolean; message: string }>(`/llm/providers/${id}/test`);
  },

  // Model APIs
  listModels: (providerID?: number) => {
    return request.get<LLMModel[]>('/llm/models', {
      params: providerID ? { provider_id: providerID } : undefined,
    });
  },

  createModel: (data: Omit<LLMModel, 'id'>) => {
    return request.post<LLMModel>('/llm/models', data);
  },

  updateModel: (id: number, data: Partial<LLMModel>) => {
    return request.put<LLMModel>(`/llm/models/${id}`, data);
  },

  deleteModel: (id: number) => {
    return request.delete<{ message: string }>(`/llm/models/${id}`);
  },
};
