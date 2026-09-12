import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getConfig, updateConfig, updateSecrets } from "../api/client";

interface SectionProps {
  title: string;
  description: string;
  children: React.ReactNode;
}

function Section({ title, description, children }: SectionProps) {
  return (
    <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle p-5 mb-4">
      <h3 className="text-lg font-semibold text-ctp-text mb-1">{title}</h3>
      <p className="text-sm text-ctp-subtext0 mb-4">{description}</p>
      <div className="space-y-3">{children}</div>
    </div>
  );
}

interface FieldProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  placeholder?: string;
  help?: string;
}

function Field({ label, value, onChange, type = "text", placeholder, help }: FieldProps) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-ctp-subtext1">{label}</span>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
      />
      {help && <span className="text-xs text-ctp-overlay1 mt-1 block">{help}</span>}
    </label>
  );
}

interface ToggleProps {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}

function Toggle({ label, checked, onChange }: ToggleProps) {
  return (
    <label className="flex items-center justify-between cursor-pointer">
      <span className="text-sm font-medium text-ctp-subtext1">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
          checked ? "bg-ctp-blue" : "bg-ctp-surface2"
        }`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-ctp-text transition-transform ${
            checked ? "translate-x-6" : "translate-x-1"
          }`}
        />
      </button>
    </label>
  );
}

interface SelectProps {
  label: string;
  value: string;
  options: string[];
  onChange: (v: string) => void;
}

function Select({ label, value, options, onChange }: SelectProps) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-ctp-subtext1">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
      >
        {options.map((o) => (
          <option key={o} value={o}>
            {o}
          </option>
        ))}
      </select>
    </label>
  );
}

type Config = Record<string, Record<string, unknown>>;

export default function Settings() {
  const queryClient = useQueryClient();
  const [toast, setToast] = useState<{ message: string; type: "success" | "error" } | null>(null);

  // Secrets are separate — never shown from config
  const [hfToken, setHfToken] = useState("");
  const [apiKeyValue, setApiKeyValue] = useState("");

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["config"],
    queryFn: getConfig,
  });

  const configMut = useMutation({
    mutationFn: updateConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["config"] });
      queryClient.invalidateQueries({ queryKey: ["setup-status"] });
      showToast("Settings saved", "success");
    },
    onError: (err: Error) => showToast(err.message, "error"),
  });

  const secretsMut = useMutation({
    mutationFn: updateSecrets,
    onSuccess: () => {
      setHfToken("");
      setApiKeyValue("");
      queryClient.invalidateQueries({ queryKey: ["setup-status"] });
      showToast("Secrets updated", "success");
    },
    onError: (err: Error) => showToast(err.message, "error"),
  });

  function showToast(message: string, type: "success" | "error") {
    setToast({ message, type });
    setTimeout(() => setToast(null), 3000);
  }

  if (isLoading) {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-4">Settings</h2>
        <p className="text-ctp-subtext0">Loading configuration...</p>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-4">Settings</h2>
        <div className="rounded-lg border border-ctp-red/30 bg-ctp-red/10 px-4 py-3 text-sm text-ctp-red">
          Failed to load configuration: {error?.message ?? "Unknown error"}
        </div>
      </div>
    );
  }

  const cfg = data.config as Config;

  function update(section: string, key: string, value: unknown) {
    const updated = { [section]: { [key]: value } };
    configMut.mutate(updated);
  }

  function saveSecrets() {
    const payload: { hf_token?: string; api_key?: string } = {};
    if (hfToken) payload.hf_token = hfToken;
    if (apiKeyValue) payload.api_key = apiKeyValue;
    if (Object.keys(payload).length > 0) {
      secretsMut.mutate(payload);
    }
  }

  const storage = (cfg["storage"] ?? {}) as Record<string, unknown>;
  const download = (cfg["download"] ?? {}) as Record<string, unknown>;
  const sources = (cfg["sources"] ?? {}) as Record<string, unknown>;
  const search = (cfg["search"] ?? {}) as Record<string, unknown>;
  const display = (cfg["display"] ?? {}) as Record<string, unknown>;
  const secrets = (cfg["secrets"] ?? {}) as Record<string, unknown>;

  return (
    <div className="max-w-2xl">
      <h2 className="text-2xl font-bold mb-6">Settings</h2>

      {/* Toast */}
      {toast && (
        <div
          className={`mb-4 rounded-lg px-4 py-2 text-sm ${
            toast.type === "success"
              ? "bg-ctp-green/10 text-ctp-green border border-ctp-green/30"
              : "bg-ctp-red/10 text-ctp-red border border-ctp-red/30"
          }`}
        >
          {toast.message}
        </div>
      )}

      <Section title="Storage" description="Where models are stored on disk.">
        <Field
          label="Base path"
          value={String(storage["base_path"] ?? "")}
          onChange={(v) => update("storage", "base_path", v)}
          placeholder="/data/models"
          help="Absolute path to the model storage directory"
        />
        <Field
          label="Max size (GB)"
          value={storage["max_size_gb"] != null ? String(storage["max_size_gb"]) : ""}
          onChange={(v) => update("storage", "max_size_gb", v ? Number(v) : null)}
          type="number"
          placeholder="No limit"
          help="Maximum total size for stored models (leave empty for unlimited)"
        />
      </Section>

      <Section title="Download" description="Download behavior settings.">
        <Toggle
          label="Resume interrupted downloads"
          checked={Boolean(download["resume"])}
          onChange={(v) => update("download", "resume", v)}
        />
        <Toggle
          label="Verify checksums"
          checked={Boolean(download["verify_checksums"])}
          onChange={(v) => update("download", "verify_checksums", v)}
        />
      </Section>

      <Section title="Sources" description="Model source configuration.">
        <Select
          label="Default source"
          value={String(sources["default"] ?? "huggingface")}
          options={["huggingface"]}
          onChange={(v) => update("sources", "default", v)}
        />
      </Section>

      <Section title="Search" description="Default search behavior.">
        <Field
          label="Default result limit"
          value={String(search["default_limit"] ?? 20)}
          onChange={(v) => update("search", "default_limit", Number(v))}
          type="number"
        />
        <Select
          label="Sort by"
          value={String(search["sort_by"] ?? "downloads")}
          options={["downloads", "likes", "created_at", "updated_at"]}
          onChange={(v) => update("search", "sort_by", v)}
        />
      </Section>

      <Section title="Display" description="UI preferences.">
        <Toggle
          label="Color output (CLI)"
          checked={Boolean(display["color"])}
          onChange={(v) => update("display", "color", v)}
        />
      </Section>

      <Section title="Authentication" description="API keys and tokens. Values are write-only and never displayed.">
        <div className="space-y-3">
          <Field
            label="HuggingFace Token"
            value={hfToken}
            onChange={setHfToken}
            type="password"
            placeholder={secrets["hf_token"] ? "Token is set (enter new value to change)" : "Enter HF token"}
            help="Required for accessing gated models"
          />
          <Field
            label="Devon API Key"
            value={apiKeyValue}
            onChange={setApiKeyValue}
            type="password"
            placeholder={secrets["api_key"] ? "Key is set (enter new value to change)" : "Enter API key"}
            help="Bearer token for authenticating API requests"
          />
          <button
            type="button"
            onClick={saveSecrets}
            disabled={(!hfToken && !apiKeyValue) || secretsMut.isPending}
            className="rounded-lg bg-ctp-blue px-4 py-2 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {secretsMut.isPending ? "Saving..." : "Save Secrets"}
          </button>
        </div>
      </Section>
    </div>
  );
}
