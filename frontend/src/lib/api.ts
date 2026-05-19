const API_BASE = import.meta.env.VITE_API_URL || "";

function getToken(): string | null {
  return localStorage.getItem("throtl_token");
}

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(opts?.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...opts,
    headers,
  });
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      if (typeof body.error === "string") message = body.error;
      else if (typeof body.message === "string") message = body.message;
    } catch {
      // Response body is not valid JSON — use the default status-based message
    }
    throw new Error(message);
  }
  if (res.status === 204) return undefined as unknown as T;
  return res.json();
}

// --- Types ---

export interface Provider {
  id: string;
  name: string;
  type: string;  // "openai" | "anthropic"
  base_url: string;
  api_key: string;
  created_at: string;
}

export interface RateLimitStatus {
  daily_tokens_in_count: number;
  daily_tokens_in_limit: number;
  daily_tokens_out_count: number;
  daily_tokens_out_limit: number;
  daily_tokens_reset?: string;
}

export interface APIKey {
  id: string;
  name: string;
  key: string;
  limit_daily: number;
  limit_tokens_in_daily: number;
  limit_tokens_out_daily: number;
  allowed_models: string;
  active: boolean;
  created_at: string;
  last_used_at?: string;
  rate_limit: RateLimitStatus;
}

export interface UsageLog {
  id: string;
  api_key_id: string;
  provider: string;
  model: string;
  status: number;
  tokens_in: number;
  tokens_out: number;
  latency_ms: number;
  created_at: string;
}

export interface DashboardStats {
  total_keys: number;
  active_keys: number;
  total_providers: number;
  total_requests: number;
  requests_today: number;
  total_tokens_in: number;
  total_tokens_out: number;
  recent_requests: UsageLog[];
  key_usage: KeyUsageStat[];
  model_breakdown: ModelStat[];
}

export interface KeyUsageStat {
  key_name: string;
  key_prefix: string;
  requests: number;
  tokens_in: number;
  tokens_out: number;
}

export interface ModelStat {
  model: string;
  requests: number;
  tokens_in: number;
  tokens_out: number;
}

export interface Model {
  id: string;
  object: string;
  created: number;
  owned_by: string;
  provider_id: string;
  active: boolean;
  request_multiplier: number;
}

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}

// --- API Calls ---

export const api = {
  // Auth (public)
  checkSetup: () =>
    request<{ setup_required: boolean }>("/api/auth/check"),
  setup: (data: { email: string; password: string; name: string }) =>
    request<AuthResponse>("/api/auth/setup", { method: "POST", body: JSON.stringify(data) }),
  login: (data: { email: string; password: string }) =>
    request<AuthResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(data) }),
  getMe: () =>
    request<AuthUser>("/api/me"),

  // Stats
  getStats: () => request<DashboardStats>("/api/stats"),
  getUsageLogs: () => request<UsageLog[]>("/api/usage"),

  // Providers
  listProviders: () => request<Provider[]>("/api/providers"),
  createProvider: (data: { id: string; name: string; type: string; base_url: string; api_key: string }) =>
    request<Provider>("/api/providers", { method: "POST", body: JSON.stringify(data) }),
  deleteProvider: (id: string) =>
    request<void>(`/api/providers/${id}`, { method: "DELETE" }),

  // API Keys
  listKeys: () => request<APIKey[]>("/api/keys"),
  createKey: (data: {
    name: string;
    limit_daily?: number;
    limit_tokens_in_daily?: number;
    limit_tokens_out_daily?: number;
    allowed_models?: string;
  }) => request<APIKey>("/api/keys", { method: "POST", body: JSON.stringify(data) }),
  toggleKey: (id: string, active: boolean) =>
    request<void>(`/api/keys/${id}?active=${active}`, { method: "PATCH" }),
  deleteKey: (id: string) =>
    request<void>(`/api/keys/${id}`, { method: "DELETE" }),

  // Models
  listModels: () => request<{ object: string; data: Model[] }>("/api/models"),
  toggleModel: (id: string, active: boolean) =>
    request<void>(`/api/models/${id}?active=${active}`, { method: "PATCH" }),
  updateModel: (id: string, data: { active?: boolean; request_multiplier?: number }) =>
    request<Model>(`/api/models/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
};
