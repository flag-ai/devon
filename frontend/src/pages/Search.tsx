import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { searchModels, startDownload, type SearchParams, type ModelResult } from "../api/client";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const val = bytes / Math.pow(1024, i);
  return `${val.toFixed(i > 1 ? 1 : 0)} ${units[i]}`;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

export default function Search() {
  const [searchParams, setSearchParams] = useSearchParams();

  const [query, setQuery] = useState(searchParams.get("query") ?? "");
  const [provider, setProvider] = useState("");
  const [params, setParams] = useState("");
  const [size, setSize] = useState("");
  const [format, setFormat] = useState("");
  const [task, setTask] = useState("");
  const [license, setLicense] = useState("");
  const [limit, setLimit] = useState(20);
  const [showFilters, setShowFilters] = useState(false);

  const [activeSearch, setActiveSearch] = useState<SearchParams | null>(
    searchParams.get("query") ? { query: searchParams.get("query") ?? undefined } : null,
  );
  const [downloading, setDownloading] = useState<Record<string, "pending" | "started" | "done" | "error">>({});
  const [dlErrors, setDlErrors] = useState<Record<string, string>>({});

  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ["search", activeSearch],
    queryFn: () => searchModels(activeSearch!),
    enabled: activeSearch !== null,
  });

  const dlMut = useMutation({
    mutationFn: startDownload,
    onSuccess: (data, variables) => {
      if (data.cached) {
        setDownloading((prev) => ({ ...prev, [variables.model_id]: "done" }));
        queryClient.invalidateQueries({ queryKey: ["models"] });
        queryClient.invalidateQueries({ queryKey: ["storage-status"] });
      } else {
        setDownloading((prev) => ({ ...prev, [variables.model_id]: "started" }));
        queryClient.invalidateQueries({ queryKey: ["download-jobs"] });
      }
    },
    onError: (err, variables) => {
      setDownloading((prev) => ({ ...prev, [variables.model_id]: "error" }));
      setDlErrors((prev) => ({ ...prev, [variables.model_id]: (err as Error).message }));
    },
  });

  function handleDownload(model: ModelResult) {
    setDownloading((prev) => ({ ...prev, [model.model_id]: "pending" }));
    dlMut.mutate({ model_id: model.model_id, source: model.source });
  }

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    const p: SearchParams = { limit };
    if (query) p.query = query;
    if (provider) p.provider = provider;
    if (params) p.params = params;
    if (size) p.size = size;
    if (format) p.format = format;
    if (task) p.task = task;
    if (license) p.license = license;
    setActiveSearch(p);
    setSearchParams(query ? { query } : {});
  }

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Search Models</h2>

      {/* Search form */}
      <form onSubmit={handleSearch} className="mb-6 space-y-3">
        <div className="flex gap-2">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search models (e.g. llama 7b gguf)..."
            className="flex-1 rounded-lg border border-ctp-surface1 bg-ctp-mantle px-4 py-2.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
          />
          <button
            type="submit"
            className="rounded-lg bg-ctp-blue px-5 py-2.5 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 transition-colors"
          >
            Search
          </button>
          <button
            type="button"
            onClick={() => setShowFilters(!showFilters)}
            className={`rounded-lg border px-3 py-2.5 text-sm transition-colors ${
              showFilters
                ? "border-ctp-blue text-ctp-blue bg-ctp-blue/10"
                : "border-ctp-surface1 text-ctp-subtext0 hover:border-ctp-overlay0"
            }`}
          >
            Filters
          </button>
        </div>

        {/* Filter panel */}
        {showFilters && (
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3 rounded-lg border border-ctp-surface0 bg-ctp-mantle p-4">
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Provider</span>
              <input
                type="text"
                value={provider}
                onChange={(e) => setProvider(e.target.value)}
                placeholder="e.g. meta-llama"
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none"
              />
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Parameters</span>
              <input
                type="text"
                value={params}
                onChange={(e) => setParams(e.target.value)}
                placeholder="e.g. 7b, 13b"
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none"
              />
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Size</span>
              <input
                type="text"
                value={size}
                onChange={(e) => setSize(e.target.value)}
                placeholder="e.g. <100gb"
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none"
              />
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Format</span>
              <select
                value={format}
                onChange={(e) => setFormat(e.target.value)}
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text focus:border-ctp-blue focus:outline-none"
              >
                <option value="">Any</option>
                <option value="gguf">GGUF</option>
                <option value="safetensors">SafeTensors</option>
                <option value="pytorch">PyTorch</option>
                <option value="onnx">ONNX</option>
              </select>
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Task</span>
              <input
                type="text"
                value={task}
                onChange={(e) => setTask(e.target.value)}
                placeholder="e.g. text-generation"
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none"
              />
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">License</span>
              <input
                type="text"
                value={license}
                onChange={(e) => setLicense(e.target.value)}
                placeholder="e.g. apache-2.0"
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none"
              />
            </label>
            <label className="block">
              <span className="text-xs font-medium text-ctp-subtext0">Result Limit</span>
              <input
                type="number"
                value={limit}
                onChange={(e) => setLimit(Number(e.target.value))}
                min={1}
                max={100}
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-2 py-1.5 text-sm text-ctp-text focus:border-ctp-blue focus:outline-none"
              />
            </label>
          </div>
        )}
      </form>

      {/* Results */}
      {isLoading && <p className="text-ctp-subtext0">Searching...</p>}
      {error && (
        <div className="rounded-lg bg-ctp-red/10 border border-ctp-red/30 px-4 py-3 text-sm text-ctp-red">
          {(error as Error).message}
        </div>
      )}

      {data && (
        <div>
          <p className="text-sm text-ctp-subtext0 mb-3">
            {data.count} result{data.count !== 1 ? "s" : ""} from {data.source}
          </p>

          <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-ctp-surface0 text-ctp-subtext0">
                  <th className="text-left px-4 py-2 font-medium">Model</th>
                  <th className="text-left px-4 py-2 font-medium">Format</th>
                  <th className="text-right px-4 py-2 font-medium">Size</th>
                  <th className="text-right px-4 py-2 font-medium">Downloads</th>
                  <th className="text-right px-4 py-2 font-medium">Likes</th>
                  <th className="px-4 py-2 font-medium w-24"></th>
                </tr>
              </thead>
              <tbody>
                {data.results.map((r: ModelResult) => (
                  <tr
                    key={r.model_id}
                    className="border-b border-ctp-surface0 last:border-0 hover:bg-ctp-surface0/50"
                  >
                    <td className="px-4 py-2">
                      <div>
                        <a
                          href={r.web_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-ctp-blue hover:underline"
                        >
                          {r.model_id}
                        </a>
                        <div className="text-xs text-ctp-overlay1 mt-0.5">
                          {r.author}
                          {r.parameter_count && ` · ${r.parameter_count}B params`}
                          {r.quantization && ` · ${r.quantization}`}
                          {r.license && ` · ${r.license}`}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex flex-wrap gap-1">
                        {r.format.map((f) => (
                          <span
                            key={f}
                            className="inline-block rounded bg-ctp-surface0 px-1.5 py-0.5 text-xs text-ctp-subtext0"
                          >
                            {f}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-2 text-right text-ctp-subtext0">
                      {formatBytes(r.total_size_bytes)}
                    </td>
                    <td className="px-4 py-2 text-right text-ctp-subtext0">
                      {formatNumber(r.downloads)}
                    </td>
                    <td className="px-4 py-2 text-right text-ctp-subtext0">
                      {formatNumber(r.likes)}
                    </td>
                    <td className="px-4 py-2 text-right">
                      {downloading[r.model_id] === "done" ? (
                        <span className="text-xs text-ctp-green">Downloaded</span>
                      ) : downloading[r.model_id] === "started" ? (
                        <Link
                          to="/downloads"
                          className="text-xs text-ctp-blue hover:underline"
                        >
                          View in Downloads
                        </Link>
                      ) : downloading[r.model_id] === "error" ? (
                        <div className="flex flex-col items-end gap-1">
                          {dlErrors[r.model_id] && (
                            <span className="text-xs text-ctp-red max-w-48 truncate" title={dlErrors[r.model_id]}>
                              {dlErrors[r.model_id]}
                            </span>
                          )}
                          <button
                            onClick={() => handleDownload(r)}
                            className="rounded bg-ctp-red/10 px-2.5 py-1 text-xs text-ctp-red hover:bg-ctp-red/20 transition-colors"
                          >
                            Retry
                          </button>
                        </div>
                      ) : downloading[r.model_id] === "pending" ? (
                        <span className="text-xs text-ctp-blue animate-pulse">Starting...</span>
                      ) : (
                        <button
                          onClick={() => handleDownload(r)}
                          className="rounded bg-ctp-blue/10 px-2.5 py-1 text-xs text-ctp-blue hover:bg-ctp-blue/20 transition-colors"
                        >
                          Download
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {data.results.length === 0 && (
              <div className="px-4 py-8 text-center text-ctp-overlay0 text-sm">
                No models found matching your query.
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
