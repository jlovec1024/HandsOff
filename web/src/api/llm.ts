import request from "./request";
import type { LLMProvider } from "../types";

export const llmApi = {
  // Provider APIs
  listProviders: () => {
    return request.get<LLMProvider[]>("/llm/providers");
  },

  getProvider: (id: number) => {
    return request.get<LLMProvider>(`/llm/providers/${id}`);
  },

  createProvider: (data: Omit<LLMProvider, "id">) => {
    return request.post<LLMProvider>("/llm/providers", data);
  },

  updateProvider: (id: number, data: Partial<LLMProvider>) => {
    return request.put<LLMProvider>(`/llm/providers/${id}`, data);
  },

  deleteProvider: (id: number) => {
    return request.delete<{ message: string }>(`/llm/providers/${id}`);
  },

  testProvider: (id: number) => {
    return request.post<{ success: boolean; message: string }>(
      `/llm/providers/${id}/test`
    );
  },

  // Fetch available models from provider
  fetchModels: (baseURL: string, apiKey: string) => {
    return request.post<{ models: string[] }>("/llm/providers/models", {
      base_url: baseURL,
      api_key: apiKey,
    });
  },
};
