// Typed wrapper around the DEVON HTTP API. All /api/v1 routes require
// a Bearer token — held in sessionStorage so a page reload survives but
// a new browser tab re-prompts.

const TOKEN_KEY = "devon_admin_token";

export function getToken(): string | null {
  return sessionStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  sessionStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  sessionStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(path, { ...init, headers });
  const text = await res.text();
  const data = text ? safeParse(text) : null;

  if (!res.ok) {
    const msg =
      (data && typeof data === "object" && "error" in data
        ? String((data as { error: unknown }).error)
        : res.statusText) || "request failed";
    throw new ApiError(res.status, msg);
  }
  return data as T;
}

function safeParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

// --- Health ---
export type HealthReport = {
  healthy: boolean;
  version: string;
  checks: { name: string; healthy: boolean; error?: string; latency_ms: number }[];
};

export const getHealth = () => req<{ status: string }>("/health");
export const getReady = () => req<HealthReport>("/ready");

// --- Setup ---
export type SetupResponse = {
  status: "provisioned" | "already_provisioned";
  admin_token?: string;
  message?: string;
};
export const runSetup = (adminToken?: string) =>
  req<SetupResponse>("/api/v1/setup", {
    method: "POST",
    body: adminToken ? JSON.stringify({ admin_token: adminToken }) : undefined,
  });

// --- Search ---
export type ModelMetadata = {
  source: string;
  model_id: string;
  author?: string;
  description?: string;
  tags?: string[];
  license?: string;
  pipeline_tag?: string;
  params_billions?: number;
  downloads?: number;
  likes?: number;
  size_bytes?: number;
  formats?: string[];
  created_at?: string;
  updated_at?: string;
  url?: string;
};
export type SearchResponse = {
  source: string;
  count: number;
  results: ModelMetadata[];
};
export const search = (params: Record<string, string | number | undefined>) => {
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === "") continue;
    qs.set(k, String(v));
  }
  return req<SearchResponse>(`/api/v1/search?${qs.toString()}`);
};

// --- Models ---
export type Placement = {
  id: string;
  model_db_id: string;
  source: string;
  model_id: string;
  agent_id: string;
  agent_name: string;
  agent_url: string;
  remote_entry_id: string;
  host_path: string;
  size_bytes: number;
  fetched_at: string;
};
export type ModelView = {
  id: string;
  source: string;
  model_id: string;
  metadata: ModelMetadata;
  placements: Placement[];
};
export const listModels = () => req<ModelView[]>("/api/v1/models");
export const getModel = (source: string, modelId: string) =>
  req<ModelView>(`/api/v1/models/${encodeURIComponent(source)}/${encodeURIComponent(modelId)}`);
export const deleteModel = (source: string, modelId: string) =>
  req<void>(`/api/v1/models/${encodeURIComponent(source)}/${encodeURIComponent(modelId)}`, {
    method: "DELETE",
  });

// --- Agents ---
export type BonnieAgent = {
  id: string;
  name: string;
  url: string;
  status: string;
  last_seen_at?: string;
  created_at: string;
  updated_at: string;
};
export const listAgents = () => req<BonnieAgent[]>("/api/v1/bonnie-agents");
export const createAgent = (body: { name: string; url: string; token: string }) =>
  req<BonnieAgent>("/api/v1/bonnie-agents", { method: "POST", body: JSON.stringify(body) });
export const deleteAgent = (id: string) =>
  req<void>(`/api/v1/bonnie-agents/${id}`, { method: "DELETE" });

// --- Downloads ---
export type Job = {
  id: string;
  model_id: string;
  bonnie_agent_id: string;
  status: "pending" | "running" | "succeeded" | "failed";
  patterns: string[];
  error?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  updated_at: string;
};
export const listDownloads = () => req<Job[]>("/api/v1/downloads");
export const getDownload = (id: string) => req<Job>(`/api/v1/downloads/${id}`);
export const restartDownload = (id: string) =>
  req<void>(`/api/v1/downloads/${id}/restart`, { method: "POST" });
export const startDownload = (body: {
  source: string;
  model_id: string;
  bonnie_agent_id: string;
  patterns?: string[];
}) =>
  req<Job>("/api/v1/models/download", { method: "POST", body: JSON.stringify(body) });

// --- Config & secrets ---
export const getConfig = () => req<Record<string, unknown>>("/api/v1/config");
export const putConfig = (cfg: Record<string, unknown>) =>
  req<Record<string, unknown>>("/api/v1/config", { method: "PUT", body: JSON.stringify(cfg) });
export const getSecrets = () => req<Record<string, string>>("/api/v1/config/secrets");
export const putSecrets = (s: Record<string, string>) =>
  req<Record<string, string>>("/api/v1/config/secrets", {
    method: "PUT",
    body: JSON.stringify(s),
  });

// --- Scan & export ---
export const runScan = (agentID?: string) =>
  req<
    {
      bonnie_agent_id: string;
      bonnie_agent_name: string;
      discovered: number;
      persisted: number;
      error?: string;
    }[]
  >("/api/v1/scan", {
    method: "POST",
    body: agentID ? JSON.stringify({ bonnie_agent_id: agentID }) : undefined,
  });
