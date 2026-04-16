import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getConfig, getSecrets, putConfig, putSecrets } from "../api";

export default function Settings() {
  const qc = useQueryClient();
  const config = useQuery({ queryKey: ["config"], queryFn: getConfig });
  const secrets = useQuery({ queryKey: ["secrets"], queryFn: getSecrets });

  const [configText, setConfigText] = useState("");
  const [hfToken, setHfToken] = useState("");

  useEffect(() => {
    if (config.data) setConfigText(JSON.stringify(config.data, null, 2));
  }, [config.data]);

  const updateConfig = useMutation({
    mutationFn: putConfig,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["config"] }),
  });
  const updateSecrets = useMutation({
    mutationFn: putSecrets,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["secrets"] });
      setHfToken("");
    },
  });

  return (
    <div className="space-y-8">
      <h2 className="text-2xl font-semibold text-mocha-lavender">Settings</h2>

      <section>
        <h3 className="mb-3 text-lg font-medium">Config</h3>
        <textarea
          value={configText}
          onChange={(e) => setConfigText(e.target.value)}
          rows={10}
          className="w-full rounded border border-mocha-surface1 bg-mocha-mantle p-3 font-mono text-xs"
        />
        <button
          type="button"
          className="mt-2 rounded bg-mocha-blue px-4 py-2 text-sm font-medium text-mocha-crust hover:bg-mocha-sky"
          onClick={() => {
            try {
              updateConfig.mutate(JSON.parse(configText));
            } catch {
              alert("Invalid JSON");
            }
          }}
        >
          Save config
        </button>
      </section>

      <section>
        <h3 className="mb-3 text-lg font-medium">Secrets</h3>
        <table className="mb-4 w-full text-sm">
          <tbody>
            {secrets.data &&
              Object.entries(secrets.data).map(([k, v]) => (
                <tr key={k} className="border-b border-mocha-surface0 last:border-0">
                  <td className="py-1 font-mono">{k}</td>
                  <td className="text-mocha-overlay1">{v || "(unset)"}</td>
                </tr>
              ))}
          </tbody>
        </table>
        <div className="grid grid-cols-3 gap-2">
          <input
            value={hfToken}
            onChange={(e) => setHfToken(e.target.value)}
            placeholder="New HF token"
            type="password"
            className="col-span-2 rounded border border-mocha-surface1 bg-mocha-mantle p-2"
          />
          <button
            type="button"
            onClick={() => updateSecrets.mutate({ hf_token: hfToken })}
            className="rounded bg-mocha-blue px-4 py-2 text-sm font-medium text-mocha-crust hover:bg-mocha-sky"
          >
            Rotate hf_token
          </button>
        </div>
      </section>
    </div>
  );
}
