import { useQuery } from "@tanstack/react-query";
import { getReady, listAgents, listDownloads, listModels } from "../api";

export default function Dashboard() {
  const ready = useQuery({ queryKey: ["ready"], queryFn: getReady });
  const models = useQuery({ queryKey: ["models"], queryFn: listModels });
  const agents = useQuery({ queryKey: ["agents"], queryFn: listAgents });
  const downloads = useQuery({ queryKey: ["downloads"], queryFn: listDownloads });

  const pending =
    downloads.data?.filter((j) => j.status === "pending" || j.status === "running").length ?? 0;
  const failed = downloads.data?.filter((j) => j.status === "failed").length ?? 0;

  return (
    <div>
      <h2 className="mb-6 text-2xl font-semibold text-mocha-lavender">Dashboard</h2>

      <div className="grid grid-cols-4 gap-4">
        <Card title="Tracked models" value={models.data?.length ?? "—"} />
        <Card title="Registered agents" value={agents.data?.length ?? "—"} />
        <Card title="In-flight downloads" value={pending} />
        <Card title="Failed downloads" value={failed} />
      </div>

      <div className="mt-8 rounded border border-mocha-surface1 bg-mocha-mantle p-4">
        <h3 className="mb-3 text-lg font-medium">Readiness</h3>
        {ready.isLoading && <p className="text-mocha-subtext0">Checking…</p>}
        {ready.data && (
          <table className="w-full text-sm">
            <tbody>
              {ready.data.checks.map((c) => (
                <tr key={c.name} className="border-b border-mocha-surface0 last:border-0">
                  <td className="py-1 font-mono text-mocha-subtext1">{c.name}</td>
                  <td className={c.healthy ? "text-mocha-green" : "text-mocha-red"}>
                    {c.healthy ? "ok" : c.error ?? "failing"}
                  </td>
                  <td className="text-right text-mocha-overlay1">{c.latency_ms} ms</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

function Card({ title, value }: { title: string; value: string | number }) {
  return (
    <div className="rounded border border-mocha-surface1 bg-mocha-mantle p-4">
      <div className="text-xs uppercase tracking-wide text-mocha-overlay1">{title}</div>
      <div className="mt-2 text-2xl font-semibold text-mocha-text">{value}</div>
    </div>
  );
}
