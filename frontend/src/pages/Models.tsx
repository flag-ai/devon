import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { listModels, deleteModel, scanModels, type LocalModel, type ScanResponse } from "../api/client";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const val = bytes / Math.pow(1024, i);
  return `${val.toFixed(i > 1 ? 1 : 0)} ${units[i]}`;
}

function ModelDrawer({ model, onClose }: { model: LocalModel; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative w-full max-w-md bg-ctp-mantle border-l border-ctp-surface0 p-6 overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-bold text-ctp-text">Model Details</h3>
          <button
            onClick={onClose}
            className="text-ctp-overlay1 hover:text-ctp-text transition-colors"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Model ID</p>
            <p className="text-sm text-ctp-text mt-1">{model.model_id}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Source</p>
            <p className="text-sm text-ctp-text mt-1">{model.source}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Path</p>
            <p className="text-sm text-ctp-text mt-1 font-mono break-all">{model.path}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Size</p>
            <p className="text-sm text-ctp-text mt-1">{formatBytes(model.size_bytes)}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Downloaded</p>
            <p className="text-sm text-ctp-text mt-1">
              {new Date(model.downloaded_at).toLocaleString()}
            </p>
          </div>
          {model.last_used && (
            <div>
              <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">Last Used</p>
              <p className="text-sm text-ctp-text mt-1">
                {new Date(model.last_used).toLocaleString()}
              </p>
            </div>
          )}
          <div>
            <p className="text-xs font-medium text-ctp-subtext0 uppercase tracking-wider">
              Files ({model.files.length})
            </p>
            <ul className="mt-1 text-sm text-ctp-subtext1 space-y-0.5 max-h-48 overflow-y-auto">
              {model.files.map((f) => (
                <li key={f} className="font-mono text-xs">{f}</li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function Models() {
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState<LocalModel | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<LocalModel | null>(null);
  const [scanReconcile, setScanReconcile] = useState(false);
  const [scanResult, setScanResult] = useState<ScanResponse | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["models"],
    queryFn: () => listModels(),
  });

  const deleteMut = useMutation({
    mutationFn: ({ source, modelId }: { source: string; modelId: string }) =>
      deleteModel(source, modelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["models"] });
      queryClient.invalidateQueries({ queryKey: ["storage-status"] });
      setConfirmDelete(null);
    },
  });

  const scanMut = useMutation({
    mutationFn: () => scanModels({ reconcile: scanReconcile }),
    onSuccess: (result) => {
      setScanResult(result);
      if (result.added > 0 || result.removed > 0) {
        queryClient.invalidateQueries({ queryKey: ["models"] });
        queryClient.invalidateQueries({ queryKey: ["storage-status"] });
      }
    },
  });

  const models = data?.models ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-2xl font-bold">Local Models</h2>
        <div className="flex items-center gap-3">
          <span className="text-sm text-ctp-subtext0">{models.length} model{models.length !== 1 ? "s" : ""}</span>
          <label className="flex items-center gap-1.5 text-xs text-ctp-subtext0 cursor-pointer">
            <input
              type="checkbox"
              checked={scanReconcile}
              onChange={(e) => setScanReconcile(e.target.checked)}
              className="accent-ctp-mauve"
            />
            Reconcile
          </label>
          <button
            onClick={() => scanMut.mutate()}
            disabled={scanMut.isPending}
            className="rounded-lg bg-ctp-mauve px-3 py-1.5 text-sm font-medium text-ctp-crust hover:bg-ctp-mauve/80 disabled:opacity-50 transition-colors"
          >
            {scanMut.isPending ? "Scanning..." : "Scan"}
          </button>
        </div>
      </div>

      {scanMut.isError && (
        <div className="rounded-lg border border-ctp-red/30 bg-ctp-red/10 px-4 py-3 mb-4 flex items-center justify-between">
          <p className="text-sm text-ctp-red">
            Scan failed: {scanMut.error instanceof Error ? scanMut.error.message : "Unknown error"}
          </p>
          <button
            onClick={() => scanMut.reset()}
            className="text-ctp-overlay1 hover:text-ctp-text transition-colors text-xs"
          >
            Dismiss
          </button>
        </div>
      )}

      {scanResult && (
        <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle px-4 py-3 mb-4 flex items-center justify-between">
          <p className="text-sm text-ctp-subtext1">
            Scan complete: <span className="text-ctp-green">{scanResult.added} added</span>
            {scanResult.stale > 0 && (
              <>, <span className="text-ctp-yellow">{scanResult.stale} stale</span></>
            )}
            {scanResult.removed > 0 && (
              <>, <span className="text-ctp-red">{scanResult.removed} removed</span></>
            )}
            , {scanResult.existing} already tracked
          </p>
          <button
            onClick={() => setScanResult(null)}
            className="text-ctp-overlay1 hover:text-ctp-text transition-colors text-xs"
          >
            Dismiss
          </button>
        </div>
      )}

      {isLoading && <p className="text-ctp-subtext0">Loading models...</p>}

      {!isLoading && models.length === 0 && (
        <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle px-4 py-12 text-center text-ctp-overlay0 text-sm">
          No models downloaded yet.
        </div>
      )}

      {models.length > 0 && (
        <div className="rounded-lg border border-ctp-surface0 bg-ctp-mantle overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-ctp-surface0 text-ctp-subtext0">
                <th className="text-left px-4 py-2 font-medium">Model</th>
                <th className="text-left px-4 py-2 font-medium">Source</th>
                <th className="text-right px-4 py-2 font-medium">Size</th>
                <th className="text-right px-4 py-2 font-medium">Downloaded</th>
                <th className="text-right px-4 py-2 font-medium w-20">Actions</th>
              </tr>
            </thead>
            <tbody>
              {models.map((m) => (
                <tr
                  key={`${m.source}::${m.model_id}`}
                  className="border-b border-ctp-surface0 last:border-0 hover:bg-ctp-surface0/50"
                >
                  <td className="px-4 py-2">
                    <button
                      onClick={() => setSelected(m)}
                      className="text-ctp-blue hover:underline text-left"
                    >
                      {m.model_id}
                    </button>
                  </td>
                  <td className="px-4 py-2 text-ctp-subtext0">{m.source}</td>
                  <td className="px-4 py-2 text-right text-ctp-subtext0">{formatBytes(m.size_bytes)}</td>
                  <td className="px-4 py-2 text-right text-ctp-overlay1">
                    {new Date(m.downloaded_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={() => setConfirmDelete(m)}
                      className="text-ctp-red hover:text-ctp-red/80 transition-colors"
                      title="Delete model"
                    >
                      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Detail drawer */}
      {selected && <ModelDrawer model={selected} onClose={() => setSelected(null)} />}

      {/* Delete confirmation */}
      {confirmDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/40" onClick={() => setConfirmDelete(null)} />
          <div className="relative rounded-lg bg-ctp-mantle border border-ctp-surface0 p-6 max-w-sm w-full mx-4">
            <h3 className="text-lg font-bold text-ctp-text mb-2">Delete Model</h3>
            <p className="text-sm text-ctp-subtext0 mb-1">
              Are you sure you want to delete this model?
            </p>
            <p className="text-sm text-ctp-text font-mono mb-4">{confirmDelete.model_id}</p>
            <p className="text-xs text-ctp-overlay1 mb-4">
              This will remove {formatBytes(confirmDelete.size_bytes)} from disk. This action cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setConfirmDelete(null)}
                className="rounded-lg border border-ctp-surface1 px-4 py-2 text-sm text-ctp-subtext1 hover:bg-ctp-surface0 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  deleteMut.mutate({
                    source: confirmDelete.source,
                    modelId: confirmDelete.model_id,
                  })
                }
                disabled={deleteMut.isPending}
                className="rounded-lg bg-ctp-red px-4 py-2 text-sm font-medium text-ctp-crust hover:bg-ctp-red/80 disabled:opacity-50 transition-colors"
              >
                {deleteMut.isPending ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
