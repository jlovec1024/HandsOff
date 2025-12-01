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
  name: string;
  type: string; // openai, deepseek, claude, etc.
  base_url: string;
  api_key: string;
  is_active: boolean;
  last_tested_at?: string;
  last_test_status?: string;
  last_test_message?: string;
  created_at?: string;
  updated_at?: string;
}

export interface LLMModel {
  id?: number;
  provider_id: number;
  model_name: string;
  display_name: string;
  description?: string;
  max_tokens: number;
  temperature: number;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
  provider?: LLMProvider;
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
  llm_model_id?: number;
  webhook_id?: number;
  webhook_url: string;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
  platform?: GitPlatformConfig;
  llm_model?: LLMModel;
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
