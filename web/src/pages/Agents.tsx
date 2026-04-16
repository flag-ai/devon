import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createAgent, deleteAgent, listAgents, runScan } from "../api";

export default function Agents() {
  const qc = useQueryClient();
  const agents = useQuery({ queryKey: ["agents"], queryFn: listAgents, refetchInterval: 10_000 });
  const create = useMutation({
    mutationFn: createAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
  const remove = useMutation({
    mutationFn: deleteAgent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["agents"] }),
  });
  const scan = useMutation({
    mutationFn: (id?: string) => runScan(id),
  });

  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [token, setToken] = useState("");

  return (
    <div>
      <h2 className="mb-6 text-2xl font-semibold text-mocha-lavender">Bonnie agents</h2>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          create.mutate(
            { name, url, token },
            {
              onSuccess: () => {
                setName("");
                setUrl("");
                setToken("");
              },
            },
          );
        }}
        className="mb-6 grid grid-cols-4 gap-2"
      >
        <input
          placeholder="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          required
        />
        <input
          placeholder="https://agent.host:8000"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          className="col-span-2 rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          required
        />
        <input
          placeholder="agent token"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          type="password"
          className="rounded border border-mocha-surface1 bg-mocha-mantle p-2"
        />
        <button
          type="submit"
          className="col-span-4 rounded bg-mocha-blue px-4 py-2 font-medium text-mocha-crust hover:bg-mocha-sky"
        >
          Register agent
        </button>
      </form>

      <div className="mb-6 flex gap-2">
        <button
          type="button"
          onClick={() => scan.mutate(undefined)}
          className="rounded border border-mocha-surface1 px-3 py-1 text-xs hover:bg-mocha-surface0"
        >
          Scan all agents
        </button>
        {scan.data && (
          <span className="text-xs text-mocha-subtext0">
            last scan: {scan.data.map((r) => `${r.bonnie_agent_name}: ${r.persisted}/${r.discovered}`).join(", ")}
          </span>
        )}
      </div>

      <div className="space-y-2">
        {agents.data?.map((a) => (
          <div
            key={a.id}
            className="flex items-center justify-between rounded border border-mocha-surface1 bg-mocha-mantle p-3"
          >
            <div>
              <div className="text-mocha-text">
                {a.name}{" "}
                <span
                  className={
                    a.status === "online" ? "text-xs text-mocha-green" : "text-xs text-mocha-red"
                  }
                >
                  ({a.status})
                </span>
              </div>
              <div className="text-xs text-mocha-overlay1">{a.url}</div>
            </div>
            <button
              type="button"
              className="text-xs text-mocha-red hover:underline"
              onClick={() => remove.mutate(a.id)}
            >
              Remove
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
