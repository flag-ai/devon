import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { listAgents, search, startDownload } from "../api";

export default function Search() {
  const [query, setQuery] = useState("");
  const [task, setTask] = useState("");
  const [format, setFormat] = useState("");
  const [limit, setLimit] = useState(30);
  type Submit = Record<string, string | number | undefined>;
  const [submitted, setSubmitted] = useState<Submit | null>(null);

  const agents = useQuery({ queryKey: ["agents"], queryFn: listAgents });
  const results = useQuery({
    queryKey: ["search", submitted],
    queryFn: () => search(submitted ?? {}),
    enabled: submitted !== null,
  });

  const download = useMutation({
    mutationFn: startDownload,
  });

  return (
    <div>
      <h2 className="mb-6 text-2xl font-semibold text-mocha-lavender">Search</h2>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          setSubmitted({ query, task, format, limit });
        }}
        className="mb-6 grid grid-cols-4 gap-3"
      >
        <input
          className="col-span-2 rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          placeholder="Query (e.g. Qwen, llama)"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <input
          className="rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          placeholder="Task (text-generation)"
          value={task}
          onChange={(e) => setTask(e.target.value)}
        />
        <input
          className="rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          placeholder="Format (gguf, safetensors)"
          value={format}
          onChange={(e) => setFormat(e.target.value)}
        />
        <input
          type="number"
          className="rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          min={1}
          max={200}
        />
        <button
          type="submit"
          className="col-span-3 rounded bg-mocha-blue px-4 py-2 font-medium text-mocha-crust hover:bg-mocha-sky"
        >
          Search
        </button>
      </form>

      {results.isLoading && <p className="text-mocha-subtext0">Searching…</p>}
      {results.error && <p className="text-mocha-red">{String(results.error)}</p>}

      <div className="space-y-3">
        {results.data?.results.map((m) => (
          <div
            key={`${m.source}:${m.model_id}`}
            className="flex items-start justify-between rounded border border-mocha-surface1 bg-mocha-mantle p-4"
          >
            <div>
              <div className="font-mono text-mocha-text">{m.model_id}</div>
              <div className="mt-1 text-xs text-mocha-overlay1">
                {m.pipeline_tag ?? "?"} · {m.license ?? "no license"} ·{" "}
                {m.params_billions ? `${m.params_billions.toFixed(1)}B` : "?B"} ·{" "}
                {(m.formats ?? []).join("/")}
              </div>
            </div>
            <select
              onChange={(e) => {
                if (!e.target.value) return;
                download.mutate({
                  source: m.source,
                  model_id: m.model_id,
                  bonnie_agent_id: e.target.value,
                });
                e.target.value = "";
              }}
              className="rounded border border-mocha-surface1 bg-mocha-base p-1 text-xs"
              defaultValue=""
            >
              <option value="">Download to…</option>
              {agents.data?.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name} ({a.status})
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>

      {download.isSuccess && <p className="mt-4 text-mocha-green">Download queued.</p>}
      {download.error && <p className="mt-4 text-mocha-red">{String(download.error)}</p>}
    </div>
  );
}
