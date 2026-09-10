import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { getStorageStatus, listModels } from "../api/client";
import SetupBanner from "../components/SetupBanner";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const val = bytes / Math.pow(1024, i);
  return `${val.toFixed(i > 1 ? 1 : 0)} ${units[i]}`;
}

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle p-4">
      <p className="text-sm text-ctp-subtext0">{label}</p>
      <p className="text-2xl font-bold text-ctp-text mt-1">{value}</p>
      {sub && <p className="text-xs text-ctp-overlay1 mt-1">{sub}</p>}
    </div>
  );
}

export default function Dashboard() {
  const navigate = useNavigate();
  const [quickSearch, setQuickSearch] = useState("");

  const { data: status } = useQuery({
    queryKey: ["storage-status"],
    queryFn: getStorageStatus,
  });

  const { data: modelsData } = useQuery({
    queryKey: ["models"],
    queryFn: () => listModels(),
  });

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (quickSearch.trim()) {
      navigate(`/search?query=${encodeURIComponent(quickSearch.trim())}`);
    }
  }

  const models = modelsData?.models ?? [];
  const recentModels = [...models]
    .sort((a, b) => b.downloaded_at.localeCompare(a.downloaded_at))
    .slice(0, 5);

  const sourceCounts = status?.sources ?? {};

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Dashboard</h2>

      <SetupBanner />

      {/* Quick search */}
      <form onSubmit={handleSearch} className="mb-6">
        <div className="relative">
          <input
            type="text"
            value={quickSearch}
            onChange={(e) => setQuickSearch(e.target.value)}
            placeholder="Quick search models..."
            className="w-full rounded-lg border border-ctp-surface1 bg-ctp-mantle px-4 py-3 pl-10 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
          />
          <svg
            className="absolute left-3 top-3.5 w-4 h-4 text-ctp-overlay0"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </form>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <StatCard
          label="Total Models"
          value={String(status?.model_count ?? 0)}
        />
        <StatCard
          label="Storage Used"
          value={status ? formatBytes(status.total_size_bytes) : "0 B"}
          sub={status?.storage_path}
        />
        <StatCard
          label="Sources"
          value={String(Object.keys(sourceCounts).length)}
          sub={Object.entries(sourceCounts)
            .map(([name, info]) => `${name}: ${info.count}`)
            .join(", ") || "No models yet"}
        />
      </div>

      {/* Recent models */}
      <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle">
        <div className="px-4 py-3 border-b border-ctp-surface0">
          <h3 className="text-sm font-semibold text-ctp-subtext1">Recent Models</h3>
        </div>
        {recentModels.length === 0 ? (
          <div className="px-4 py-8 text-center text-ctp-overlay0 text-sm">
            No models downloaded yet. Use Search or Downloads to get started.
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-ctp-surface0 text-ctp-subtext0">
                <th className="text-left px-4 py-2 font-medium">Model</th>
                <th className="text-left px-4 py-2 font-medium">Source</th>
                <th className="text-right px-4 py-2 font-medium">Size</th>
                <th className="text-right px-4 py-2 font-medium">Downloaded</th>
              </tr>
            </thead>
            <tbody>
              {recentModels.map((m) => (
                <tr
                  key={`${m.source}::${m.model_id}`}
                  className="border-b border-ctp-surface0 last:border-0 hover:bg-ctp-surface0/50 cursor-pointer"
                  onClick={() => navigate(`/models`)}
                >
                  <td className="px-4 py-2 text-ctp-text">{m.model_id}</td>
                  <td className="px-4 py-2 text-ctp-subtext0">{m.source}</td>
                  <td className="px-4 py-2 text-right text-ctp-subtext0">{formatBytes(m.size_bytes)}</td>
                  <td className="px-4 py-2 text-right text-ctp-overlay1">
                    {new Date(m.downloaded_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
