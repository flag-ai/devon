import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listDownloads, restartDownload } from "../api";

const statusColor: Record<string, string> = {
  pending: "text-mocha-yellow",
  running: "text-mocha-blue",
  succeeded: "text-mocha-green",
  failed: "text-mocha-red",
};

export default function Downloads() {
  const qc = useQueryClient();
  const jobs = useQuery({
    queryKey: ["downloads"],
    queryFn: listDownloads,
    refetchInterval: 3000,
  });
  const restart = useMutation({
    mutationFn: restartDownload,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["downloads"] }),
  });

  return (
    <div>
      <h2 className="mb-6 text-2xl font-semibold text-mocha-lavender">Downloads</h2>
      {jobs.isLoading && <p className="text-mocha-subtext0">Loading…</p>}
      <div className="space-y-2">
        {jobs.data?.map((j) => (
          <div
            key={j.id}
            className="flex items-center justify-between rounded border border-mocha-surface1 bg-mocha-mantle p-3"
          >
            <div className="flex-1">
              <div className="font-mono text-xs text-mocha-subtext1">{j.id.slice(0, 8)}</div>
              <div className="text-sm text-mocha-text">model {j.model_id.slice(0, 8)} → agent {j.bonnie_agent_id.slice(0, 8)}</div>
              {j.error && <div className="mt-1 text-xs text-mocha-red">{j.error}</div>}
            </div>
            <div className={`px-3 text-sm font-medium ${statusColor[j.status] ?? "text-mocha-subtext0"}`}>
              {j.status}
            </div>
            {(j.status === "failed" || j.status === "succeeded") && (
              <button
                type="button"
                className="text-xs text-mocha-blue hover:underline"
                onClick={() => restart.mutate(j.id)}
              >
                Restart
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
