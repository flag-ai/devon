import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deleteModel, listModels } from "../api";

export default function Models() {
  const qc = useQueryClient();
  const models = useQuery({ queryKey: ["models"], queryFn: listModels });
  const remove = useMutation({
    mutationFn: (m: { source: string; model_id: string }) => deleteModel(m.source, m.model_id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["models"] }),
  });

  return (
    <div>
      <h2 className="mb-6 text-2xl font-semibold text-mocha-lavender">Tracked models</h2>

      {models.isLoading && <p className="text-mocha-subtext0">Loading…</p>}
      {models.error && <p className="text-mocha-red">{String(models.error)}</p>}

      <div className="space-y-3">
        {models.data?.map((m) => (
          <div key={m.id} className="rounded border border-mocha-surface1 bg-mocha-mantle p-4">
            <div className="flex items-start justify-between">
              <div>
                <div className="font-mono">{m.model_id}</div>
                <div className="text-xs text-mocha-overlay1">
                  {m.source} · {m.metadata.pipeline_tag ?? "?"} ·{" "}
                  {m.metadata.params_billions ? `${m.metadata.params_billions.toFixed(1)}B` : "?B"}
                </div>
              </div>
              <button
                type="button"
                className="text-xs text-mocha-red hover:underline"
                onClick={() => remove.mutate({ source: m.source, model_id: m.model_id })}
              >
                Delete everywhere
              </button>
            </div>
            <div className="mt-3 space-y-1 text-xs">
              {m.placements.length === 0 ? (
                <div className="text-mocha-overlay1">No placements yet</div>
              ) : (
                m.placements.map((p) => (
                  <div key={p.id} className="font-mono text-mocha-subtext1">
                    {p.agent_id.slice(0, 8)} · {p.host_path}
                  </div>
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
