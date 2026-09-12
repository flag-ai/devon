import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  startDownload,
  listDownloadJobs,
  restartDownload,
  type DownloadStartResponse,
  type DownloadJobResponse,
} from "../api/client";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const val = bytes / Math.pow(1024, i);
  return `${val.toFixed(i > 1 ? 1 : 0)} ${units[i]}`;
}

function timeAgo(iso: string): string {
  const seconds = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}

function StatusBadge({ status }: { status: DownloadJobResponse["status"] }) {
  if (status === "downloading") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full bg-ctp-blue/10 px-2.5 py-0.5 text-xs font-medium text-ctp-blue">
        <svg className="animate-spin h-3 w-3" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        Downloading
      </span>
    );
  }
  if (status === "completed") {
    return (
      <span className="inline-flex items-center rounded-full bg-ctp-green/10 px-2.5 py-0.5 text-xs font-medium text-ctp-green">
        Completed
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-full bg-ctp-red/10 px-2.5 py-0.5 text-xs font-medium text-ctp-red">
      Failed
    </span>
  );
}

export default function Downloads() {
  const queryClient = useQueryClient();
  const [modelInput, setModelInput] = useState("");
  const [source, setSource] = useState("huggingface");
  const [force, setForce] = useState(false);
  const [includePatterns, setIncludePatterns] = useState("");
  const [formFeedback, setFormFeedback] = useState<{ type: "success" | "cached" | "error"; message: string } | null>(null);

  // Track previous job statuses for transition detection
  const prevJobsRef = useRef<Record<string, string>>({});

  const { data: jobsData } = useQuery({
    queryKey: ["download-jobs"],
    queryFn: listDownloadJobs,
    refetchInterval: (query) => {
      const jobs = query.state.data?.jobs;
      if (jobs?.some((j) => j.status === "downloading")) return 3000;
      return 30000;
    },
  });

  // Invalidate models/storage queries when a job transitions to completed
  useEffect(() => {
    if (!jobsData?.jobs) return;
    const prev = prevJobsRef.current;
    let shouldInvalidate = false;

    for (const job of jobsData.jobs) {
      if (prev[job.id] === "downloading" && job.status === "completed") {
        shouldInvalidate = true;
      }
    }

    // Update ref
    const next: Record<string, string> = {};
    for (const job of jobsData.jobs) {
      next[job.id] = job.status;
    }
    prevJobsRef.current = next;

    if (shouldInvalidate) {
      queryClient.invalidateQueries({ queryKey: ["models"] });
      queryClient.invalidateQueries({ queryKey: ["storage-status"] });
    }
  }, [jobsData, queryClient]);

  const dlMut = useMutation({
    mutationFn: startDownload,
    onSuccess: (data: DownloadStartResponse) => {
      if (data.cached) {
        setFormFeedback({ type: "cached", message: `${data.cached.model_id} is already downloaded.` });
      } else {
        setFormFeedback({ type: "success", message: "Download started. See the table below for progress." });
      }
      queryClient.invalidateQueries({ queryKey: ["download-jobs"] });
    },
    onError: (err: Error) => {
      setFormFeedback({ type: "error", message: err.message });
    },
  });

  const restartMut = useMutation({
    mutationFn: restartDownload,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["download-jobs"] });
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!modelInput.trim()) return;

    setFormFeedback(null);
    const patterns = includePatterns
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    dlMut.mutate({
      model_id: modelInput.trim(),
      source,
      force,
      include_patterns: patterns.length > 0 ? patterns : undefined,
    });
  }

  const jobs = jobsData?.jobs ?? [];

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Downloads</h2>

      {/* Manual download form */}
      <div className="max-w-2xl mb-8">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle p-5 space-y-4">
            <label className="block">
              <span className="text-sm font-medium text-ctp-subtext1">Model ID or URL</span>
              <input
                type="text"
                value={modelInput}
                onChange={(e) => setModelInput(e.target.value)}
                placeholder="e.g. Qwen/Qwen2.5-7B-Instruct or https://huggingface.co/..."
                className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
              />
            </label>

            <div className="grid grid-cols-2 gap-4">
              <label className="block">
                <span className="text-sm font-medium text-ctp-subtext1">Source</span>
                <select
                  value={source}
                  onChange={(e) => setSource(e.target.value)}
                  className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
                >
                  <option value="huggingface">HuggingFace</option>
                </select>
              </label>

              <label className="block">
                <span className="text-sm font-medium text-ctp-subtext1">Include patterns</span>
                <input
                  type="text"
                  value={includePatterns}
                  onChange={(e) => setIncludePatterns(e.target.value)}
                  placeholder="e.g. *.safetensors, config.json"
                  className="mt-1 block w-full rounded-md border border-ctp-surface1 bg-ctp-surface0 px-3 py-2 text-sm text-ctp-text placeholder-ctp-overlay0 focus:border-ctp-blue focus:outline-none focus:ring-1 focus:ring-ctp-blue"
                />
                <span className="text-xs text-ctp-overlay1 mt-1 block">Comma-separated glob patterns (leave empty for all files)</span>
              </label>
            </div>

            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={force}
                onChange={(e) => setForce(e.target.checked)}
                className="rounded border-ctp-surface1 bg-ctp-surface0 text-ctp-blue focus:ring-ctp-blue"
              />
              <span className="text-sm text-ctp-subtext1">Force re-download if already exists</span>
            </label>
          </div>

          <button
            type="submit"
            disabled={!modelInput.trim() || dlMut.isPending}
            className="rounded-lg bg-ctp-blue px-5 py-2.5 text-sm font-medium text-ctp-crust hover:bg-ctp-blue/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {dlMut.isPending ? "Starting..." : "Download"}
          </button>
        </form>

        {/* Form feedback */}
        {formFeedback?.type === "error" && (
          <div className="mt-4 rounded-lg bg-ctp-red/10 border border-ctp-red/30 px-4 py-3 text-sm text-ctp-red">
            {formFeedback.message}
          </div>
        )}
        {formFeedback?.type === "cached" && (
          <div className="mt-4 rounded-lg bg-ctp-yellow/10 border border-ctp-yellow/30 px-4 py-3 text-sm text-ctp-yellow">
            {formFeedback.message}
          </div>
        )}
        {formFeedback?.type === "success" && (
          <div className="mt-4 rounded-lg bg-ctp-green/10 border border-ctp-green/30 px-4 py-3 text-sm text-ctp-green">
            {formFeedback.message}
          </div>
        )}
      </div>

      {/* Job table */}
      {jobs.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3 text-ctp-text">Download History</h3>
          <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-ctp-surface0 text-ctp-subtext0">
                  <th className="text-left px-4 py-2 font-medium">Model</th>
                  <th className="text-left px-4 py-2 font-medium">Status</th>
                  <th className="text-right px-4 py-2 font-medium">Size</th>
                  <th className="text-right px-4 py-2 font-medium">Started</th>
                  <th className="px-4 py-2 font-medium w-28"></th>
                </tr>
              </thead>
              <tbody>
                {jobs.map((job) => (
                  <tr
                    key={job.id}
                    className="border-b border-ctp-surface0 last:border-0 hover:bg-ctp-surface0/50"
                  >
                    <td className="px-4 py-2">
                      <div className="text-ctp-text">{job.model_id}</div>
                      <div className="text-xs text-ctp-overlay1">{job.source}</div>
                    </td>
                    <td className="px-4 py-2">
                      <StatusBadge status={job.status} />
                      {job.error && (
                        <div className="text-xs text-ctp-red mt-1 max-w-xs truncate" title={job.error}>
                          {job.error}
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right text-ctp-subtext0">
                      {job.result ? formatBytes(job.result.size_bytes) : "--"}
                    </td>
                    <td className="px-4 py-2 text-right text-ctp-subtext0">
                      {timeAgo(job.started_at)}
                    </td>
                    <td className="px-4 py-2 text-right">
                      {job.status === "failed" && (
                        <button
                          onClick={() => restartMut.mutate(job.id)}
                          disabled={restartMut.isPending}
                          className="rounded bg-ctp-red/10 px-2.5 py-1 text-xs text-ctp-red hover:bg-ctp-red/20 transition-colors disabled:opacity-50"
                        >
                          Restart
                        </button>
                      )}
                      {job.status === "completed" && job.result && (
                        <span className="text-xs text-ctp-subtext0">
                          {job.result.files.length} file{job.result.files.length !== 1 ? "s" : ""}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {jobs.length === 0 && (
        <div className="text-center text-ctp-overlay0 text-sm py-8">
          No downloads yet. Use the form above or start a download from the Search page.
        </div>
      )}
    </div>
  );
}
