import request from "./request";
import type { Repository, GitLabRepository } from "../types";

export const repositoryApi = {
  // List repositories from GitLab
  listFromGitLab: (
    page: number = 1,
    perPage: number = 20,
    search: string = ""
  ) => {
    return request.get<{
      repositories: GitLabRepository[];
      page: number;
      per_page: number;
      total_pages: number;
    }>("/repositories/gitlab", {
      params: { page, per_page: perPage, search },
    });
  },

  // List imported repositories
  list: (page: number = 1, pageSize: number = 20) => {
    return request.get<{
      repositories: Repository[];
      page: number;
      page_size: number;
      total: number;
    }>("/repositories", {
      params: { page, page_size: pageSize },
    });
  },

  // Get repository details
  get: (id: number) => {
    return request.get<Repository>(`/repositories/${id}`);
  },

  // Batch import repositories
  batchImport: (repositoryIDs: number[], webhookCallbackURL: string) => {
    return request.post<{ message: string; count: number }>(
      "/repositories/batch",
      {
        repository_ids: repositoryIDs,
        webhook_callback_url: webhookCallbackURL,
      }
    );
  },

  // Update LLM provider for repository
  updateLLMProvider: (id: number, llmProviderID: number | null) => {
    return request.put<{ message: string }>(`/repositories/${id}/llm`, {
      llm_provider_id: llmProviderID,
    });
  },

  // Delete repository
  delete: (id: number) => {
    return request.delete<{ message: string }>(`/repositories/${id}`);
  },

  // Test webhook for repository
  testWebhook: (id: number) => {
    return request.post<{ status: string; message: string }>(
      `/repositories/${id}/webhook/test`
    );
  },

  // Recreate webhook for repository
  recreateWebhook: (id: number) => {
    return request.put<{ message: string }>(`/repositories/${id}/webhook`);
  },
};
