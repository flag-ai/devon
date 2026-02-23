const BASE_URL = "";

let apiKey: string | null = sessionStorage.getItem("devon_api_key");

export function setApiKey(key: string) {
  apiKey = key;
  sessionStorage.setItem("devon_api_key", key);
}

export function getApiKey(): string | null {
  return apiKey;
}

export function clearApiKey() {
  apiKey = null;
  sessionStorage.removeItem("devon_api_key");
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (apiKey) {
    headers["Authorization"] = `Bearer ${apiKey}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ detail: res.statusText }));
    throw new ApiError(res.status, body.detail ?? "Request failed");
  }

  return res.json() as Promise<T>;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

// --- Health ---

export interface HealthResponse {
  status: string;
  version: string;
}

export const health = () => request<HealthResponse>("/health");

// --- Config ---

export interface ConfigResponse {
  config: Record<string, unknown>;
}

export interface SetupStatusResponse {
  configured: boolean;
  missing: string[];
}

export const getConfig = () => request<ConfigResponse>("/api/v1/config");

export const updateConfig = (config: Record<string, unknown>) =>
  request<ConfigResponse>("/api/v1/config", {
    method: "PUT",
    body: JSON.stringify({ config }),
  });

export const getSetupStatus = () =>
  request<SetupStatusResponse>("/api/v1/config/setup-status");

export const updateSecrets = (secrets: { hf_token?: string; api_key?: string }) =>
  request<{ updated: string[] }>("/api/v1/config/secrets", {
    method: "PUT",
    body: JSON.stringify(secrets),
  });

// --- Search ---

export interface ModelResult {
  source: string;
  model_id: string;
  model_name: string;
  author: string;
  total_size_bytes: number;
  file_count: number;
  parameter_count: number | null;
  architecture: string | null;
  format: string[];
  quantization: string | null;
  tags: string[];
  license: string | null;
  downloads: number;
  likes: number;
  created_at: string;
  updated_at: string;
  web_url: string;
  repo_url: string;
}

export interface SearchResponse {
  query: string | null;
  source: string;
  count: number;
  results: ModelResult[];
}

export interface SearchParams {
  query?: string;
  source?: string;
  provider?: string;
  params?: string;
  size?: string;
  format?: string;
  task?: string;
  license?: string;
  limit?: number;
}

export const searchModels = (p: SearchParams) => {
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(p)) {
    if (v !== undefined && v !== "") qs.set(k, String(v));
  }
  return request<SearchResponse>(`/api/v1/search?${qs.toString()}`);
};

// --- Local Models ---

export interface LocalModel {
  source: string;
  model_id: string;
  path: string;
  size_bytes: number;
  downloaded_at: string;
  last_used: string | null;
  files: string[];
  metadata: Record<string, unknown>;
}

export interface LocalModelsResponse {
  count: number;
  models: LocalModel[];
}

export interface ModelInfoResponse {
  local: LocalModel | null;
  remote: ModelResult | null;
}

export const listModels = (source?: string) => {
  const qs = source ? `?source=${source}` : "";
  return request<LocalModelsResponse>(`/api/v1/models${qs}`);
};

export const getModelInfo = (source: string, modelId: string) =>
  request<ModelInfoResponse>(`/api/v1/models/${source}/${modelId}`);

export const deleteModel = (source: string, modelId: string) =>
  request<{ deleted: boolean; model_id: string; source: string }>(
    `/api/v1/models/${source}/${modelId}`,
    { method: "DELETE" },
  );

// --- Downloads ---

export interface DownloadRequest {
  model_id: string;
  source?: string;
  force?: boolean;
  include_patterns?: string[];
}

export interface DownloadResponse {
  model_id: string;
  source: string;
  path: string;
  files: string[];
  size_bytes: number;
}

export const downloadModel = (req: DownloadRequest) =>
  request<DownloadResponse>("/api/v1/downloads", {
    method: "POST",
    body: JSON.stringify(req),
  });

// --- Storage ---

export interface StorageStatusResponse {
  model_count: number;
  total_size_bytes: number;
  storage_path: string;
  sources: Record<string, { count: number; size_bytes: number }>;
}

export interface CleanRequest {
  unused?: boolean;
  days?: number;
  all?: boolean;
  dry_run?: boolean;
}

export interface CleanResponse {
  removed: number;
  freed_bytes: number;
  dry_run: boolean;
  models: string[];
}

export interface ExportResponse {
  format: string;
  count: number;
  content: unknown;
}

export const getStorageStatus = () =>
  request<StorageStatusResponse>("/api/v1/status");

export const cleanModels = (req: CleanRequest) =>
  request<CleanResponse>("/api/v1/clean", {
    method: "POST",
    body: JSON.stringify(req),
  });

export const exportModels = (format: string = "kitt") =>
  request<ExportResponse>("/api/v1/export", {
    method: "POST",
    body: JSON.stringify({ format }),
  });
