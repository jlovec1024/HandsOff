import request from "./request";
import type { GitPlatformConfig } from "../types";

export const platformApi = {
  // Get platform config
  getConfig: () => {
    return request.get<GitPlatformConfig | { exists: false; message: string }>(
      "/platform/config"
    );
  },
  // Update platform config
  updateConfig: (data: GitPlatformConfig) => {
    return request.put<{
      message: string;
      config?: GitPlatformConfig;
    }>("/platform/config", data);
  },

  // Test connection (accepts configuration data)
  testConnection: (data: {
    platform_type: string;
    base_url: string;
    access_token: string;
  }) => {
    return request.post<{ success: boolean; message: string }>(
      "/platform/test",
      data
    );
  },
};
