// User types
export interface User {
  id: number;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

// API Response types
export interface ApiError {
  error: string;
}

export interface HealthResponse {
  status: string;
  time: string;
  database: string;
  version: string;
}

// Git Platform types
export interface GitPlatformConfig {
  id?: number;
  platform_type: string;
  base_url: string;
  access_token: string;
  webhook_secret?: string;
  is_active: boolean;
  last_tested_at?: string;
  last_test_status?: string;
  last_test_message?: string;
  created_at?: string;
  updated_at?: string;
}

// LLM types
export interface LLMProvider {
  id?: number;
  name: string; // User-defined name, e.g., "OpenAI Official", "DeepSeek China"
  base_url: string;
  api_key: string;
  model: string; // Model name, e.g., "gpt-4", "deepseek-chat"
  is_active: boolean;
  last_tested_at?: string;
  last_test_status?: string;
  last_test_message?: string;
  created_at?: string;
  updated_at?: string;
}

// System Configuration types
export interface SystemWebhookConfig {
  webhook_callback_url: string;
}

// Repository types
export interface Repository {
  id?: number;
  platform_id: number;
  platform_repo_id: number;
  name: string;
  full_path: string;
  http_url: string;
  ssh_url: string;
  default_branch: string;
  llm_provider_id?: number;
  webhook_id?: number;
  webhook_url: string;
  is_active: boolean;

  // Webhook status tracking
  last_webhook_test_at?: string;
  last_webhook_test_status?: "success" | "failed" | "never";
  last_webhook_test_error?: string;

  created_at?: string;
  updated_at?: string;
  platform?: GitPlatformConfig;
  llm_provider?: LLMProvider;
}

export interface GitLabRepository {
  id: number;
  name: string;
  full_path: string;
  http_url: string;
  ssh_url: string;
  default_branch: string;
  description: string;
}
